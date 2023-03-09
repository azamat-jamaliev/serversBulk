package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
)

type CurrentFileForDownloading struct {
	Server         string
	RemoteFileName string
	ConfigServer   configProvider.ConfigServerType
	// SshConfig    *ssh.ClientConfig
	Error error
}

type FileSizeInfo struct {
	FileName string
	Server   string
	FileSize int64
}

var logHandler func(server, log string)
var statusHandler func(server, status string)

func StartTaskForEnv(config *configProvider.ConfigEnvironmentType,
	taskName tasks.TaskType,
	serversName,
	modifTime,
	cargo, cargo2 string,
	newLogHandler func(server, log string),
	newStatusHandler func(server, status string)) {

	logHandler = newLogHandler
	statusHandler = newStatusHandler

	numberOfServers := 0
	tasksChannel := make(chan tasks.ServerTask) //, numberOfServers)
	for _, serverConf := range config.Servers {
		if serverConf.IpAddresses != nil && (serversName == "" || serverConf.Name == serversName) {
			numberOfServers += len(serverConf.IpAddresses)
			for _, serverIp := range serverConf.IpAddresses {
				go func(serverIp string, serverConf configProvider.ConfigServerType) {
					task := tasks.ServerTask{Status: tasks.Planned,
						Type:          taskName,
						ModifTime:     modifTime,
						CommandCargo:  cargo,
						CommandCargo2: cargo2,
						ConfigServer:  serverConf,
						Server:        serverIp}
					tasksChannel <- task
				}(serverIp, serverConf)

			}
		}
	}
	if numberOfServers < 1 {
		panic("servers to connect is less than 1 - nothing to do")
	}
	count := 0
	for task := range tasksChannel {
		if task.Status == tasks.Finished || task.Status == tasks.Failed {
			count++
			//Print finished task:
			PrintTask(&task)
		} else {
			//print Inprogress task
			switch task.Type {
			case tasks.TypeGrepInLogs:
				go grepInLogs(task, tasksChannel)
			case tasks.TypeAwkInLogs:
				go awkInLogs(task, tasksChannel)
			case tasks.TypeExecuteCommand:
				go executeCommand(task, tasksChannel)
			case tasks.TypeArchiveLogs:
				go archiveLogs(task, tasksChannel)
			case tasks.TypeArchiveGzip:
				go archiveGzip(task, tasksChannel)
			case tasks.TypeDownloadArchive:
				go downloadFile(task, tasksChannel)
			case tasks.TypeUploadFile:
				go uploadFile(task, tasksChannel)
			default:
				panic(fmt.Sprintf("unknown task type task.Type=%v", task.Type))
			}
		}
		if count >= numberOfServers {
			break
		}
	}

}
func PrintTask(task *tasks.ServerTask) {
	statusHandler(task.Server, string(task.Status))
	logHandler(task.Server, fmt.Sprintf("SERVER: %s NAME: %s\n STATUS: %s", task.Server, task.Type, task.Status))
	if task.Status == tasks.Failed {
		logHandler(task.Server, task.ExecuteCmd)
		logHandler(task.Server, task.Error.Error())
	}
	logHandler(task.Server, task.RemoteFileName)
	logHandler(task.Server, task.Log)
	logHandler(task.Server, "↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
}
func taskForChannel(task *tasks.ServerTask, log string, err error, newStatus tasks.TaskStatus, nextTask *tasks.TaskType) *tasks.ServerTask {
	task.Log = log
	task.Error = err
	if err != nil {
		task.Status = tasks.Failed
	} else {
		task.Status = newStatus
		if nextTask != nil {
			task.Type = *nextTask
		}
	}
	return task
}

func fileNameFromServerIP(serverIp string) string {
	return strings.ReplaceAll(serverIp, ".", "_")
}
func printDownloadProgress(fileSizeInfo chan FileSizeInfo) {
	var fSize *FileSizeInfo
	f := <-fileSizeInfo
	fSize = &f
	for {
		fileStat, err := os.Stat(fSize.FileName)
		if err != nil {
			logHandler(fSize.Server, fmt.Sprintf("printDownloadProgress unable to get stat from file [%s]\nERROR:%v", fSize.FileName, err))
		}
		persent := math.Round(100*100*float64(fileStat.Size())/float64(fSize.FileSize)) / 100
		statusHandler(fSize.Server, fmt.Sprintf("downloaded ~%v%% [%s]", persent, fSize.FileName))
		logHandler(fSize.Server, fmt.Sprintf("SRV: [%s] ~%v%% downloaded of the file [%s]  \n", fSize.Server, persent, fSize.FileName))
		if persent > 95 {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func getFindExecCommad(logFolders []string, logFilePattern, mTime, commandToExecute string) string {
	// strGrep := fmt.Sprintf("grep --color=auto -H -A25 -B3 -i \"%s\" {}  \\;", task.CommandCargo)
	cmd := "cd ~"
	for _, folder := range logFolders {
		cmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" -mtime %s -exec %s \\;", cmd, folder, logFilePattern, mTime, commandToExecute)
	}
	return cmd
}
func getFindExecForTask(task tasks.ServerTask, commandToExecute string) string {
	return getFindExecCommad(task.ConfigServer.LogFolders, task.ConfigServer.LogFilePattern, task.ModifTime, commandToExecute)
}
