package main

import (
	"fmt"
	"log"
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
	FileName    string
	Server      string
	ServerGroup string
	FileSize    int64
}

var logHandler func(server, serverGroup, log string)
var statusHandler func(server, serverGroup, status string)

func StartTaskForEnv(config *configProvider.ConfigEnvironmentType,
	taskName tasks.TaskType,
	serversName,
	modifTime,
	cargo, cargo2 string,
	newLogHandler func(server, serverGroup, log string),
	newStatusHandler func(server, serverGroup, status string)) {

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
	statusHandler(task.Server, task.ConfigServer.Name, string(task.Status))
	logHandler(task.Server, task.ConfigServer.Name, fmt.Sprintf("SERVER: %s NAME: %s\n STATUS: %s", task.Server, task.Type, task.Status))

	logHandler(task.Server, task.ConfigServer.Name, "↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	if task.Status == tasks.Failed {
		logHandler(task.Server, task.ConfigServer.Name, task.ExecuteCmd)
		logHandler(task.Server, task.ConfigServer.Name, task.Error.Error())
	}
	logHandler(task.Server, task.ConfigServer.Name, task.RemoteFileName)
	logHandler(task.Server, task.ConfigServer.Name, task.Log)
	logHandler(task.Server, task.ConfigServer.Name, "↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
}
func taskForChannel(task *tasks.ServerTask, srvLog string, err error, newStatus tasks.TaskStatus, nextTask *tasks.TaskType) *tasks.ServerTask {
	log.Printf("[DEBUG] server=[%s] task.Type=[%s], task.CommandCargo=[%s], task.CommandCargo2[%s]", task.Server, task.Type, task.CommandCargo, task.CommandCargo2)
	task.Log = srvLog
	task.Error = err
	if err != nil {
		task.Status = tasks.Failed
		log.Printf("[ERROR] task ERROR:[%s] LOG:[%s]", err, srvLog)
	} else {
		task.Status = newStatus
		if nextTask != nil {
			task.Type = *nextTask
		}
	}
	return task
}

func fileNameFromServer(serverIp, serverGroup string) string {
	return fmt.Sprintf("%s-%s", serverGroup, strings.ReplaceAll(serverIp, ".", "_"))
}
func printDownloadProgress(fSize FileSizeInfo) {
	for {
		fileStat, err := os.Stat(fSize.FileName)
		if err != nil {
			logHandler(fSize.Server, fSize.ServerGroup, fmt.Sprintf("printDownloadProgress unable to get stat from file [%s]\nERROR:%v", fSize.FileName, err))
		}
		persent := math.Round(100*100*float64(fileStat.Size())/float64(fSize.FileSize)) / 100
		statusHandler(fSize.Server, fSize.ServerGroup, fmt.Sprintf("downloaded ~%v%% [%s]", persent, fSize.FileName))
		logHandler(fSize.Server, fSize.ServerGroup, fmt.Sprintf("SRV: [%s] ~%v%% downloaded of the file [%s]  \n", fSize.Server, persent, fSize.FileName))
		if persent > 95 {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func getFindExecCommad(logFolders []string, logFilePattern, mTime, commandToExecute, homeFolder string) string {
	// strGrep := fmt.Sprintf("grep --color=auto -H -A25 -B3 -i \"%s\" {}  \\;", task.CommandCargo)
	cmd := fmt.Sprintf("cd %s", homeFolder)
	for _, folder := range logFolders {
		cmd = fmt.Sprintf("%s; find %s -type f -iname \"%s\" -mtime %s -exec %s \\;", cmd, folder, logFilePattern, mTime, commandToExecute)
	}
	return cmd
}
func getFindExecForTask(task tasks.ServerTask, commandToExecute, homeFolder string) string {
	return getFindExecCommad(task.ConfigServer.LogFolders, task.ConfigServer.LogFilePattern, task.ModifTime, commandToExecute, homeFolder)
}
