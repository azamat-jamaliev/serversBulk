package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"sebulk/modules/configProvider"
	"sebulk/modules/sshHelper"
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
						Type:         taskName,
						ModifTime:    modifTime,
						CommandCargo: cargo,
						ConfigServer: serverConf,
						Server:       serverIp}
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
func executeOnServer(serverConf *configProvider.ConfigServerType, server, cmd string) (string, error) {
	statusHandler(server, "CONNECTING...")
	logHandler(server, fmt.Sprintf("connecting to server: [%s] to execute: [%s]", server, cmd))
	sshAdv := sshHelper.OpenSshAdvanced(serverConf, server)
	defer sshAdv.Close()
	statusHandler(server, fmt.Sprintf("executing command:[%s]", cmd))
	logHandler(server, fmt.Sprintf("executing command:[%s] on ssh server:[%s]", cmd, server))
	buff, e := sshAdv.NewSession().Output(cmd)
	if e != nil {
		logHandler(server, fmt.Sprintf("error while executing cmd:[%s] os server [%s], cmd_output[%s]\nERROR:%v", cmd, server, buff, e))
	}
	str := string(buff)
	return str, e
}
func executeCommand(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.CommandCargo)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}

// NOTE: output result is in the channel
func grepInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	task.ExecuteCmd = fmt.Sprintf("cd %s", "")
	for _, folder := range task.ConfigServer.LogFolders {
		task.ExecuteCmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" -mtime %s -exec grep --color=auto -H -A15 -B15 -i \"%s\" {}  \\;", task.ExecuteCmd, folder, task.ConfigServer.LogFilePattern, task.ModifTime, task.CommandCargo)
	}
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}
func fileNameFromServerIP(serverIp string) string {
	return strings.ReplaceAll(serverIp, ".", "_")
}
func archiveLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	tarNamefile := fmt.Sprintf("~/%s.%s", fileNameFromServerIP(task.Server), "tar")
	cmd := fmt.Sprintf("cd %s", task.ConfigServer.LogFolders[0])
	for _, folder := range task.ConfigServer.LogFolders {
		cmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" -mtime %s -exec tar rvf %s {} \\;", cmd, folder, task.ConfigServer.LogFilePattern, task.ModifTime, tarNamefile)
	}
	str, e := executeOnServer(&task.ConfigServer, task.Server, cmd)

	nextTask := tasks.TypeArchiveGzip
	task.RemoteFileName = tarNamefile
	output <- *taskForChannel(&task, str, e, tasks.InProgress, &nextTask)
}
func archiveGzip(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	tarGzNamefile := fmt.Sprintf("%s.gz", task.RemoteFileName)
	task.ExecuteCmd = fmt.Sprintf("if [ -e %s ]; then cd %s; tar cvzf %s %s ; rm %s;echo 'true';fi", task.RemoteFileName, path.Dir(tarGzNamefile), path.Base(tarGzNamefile), path.Base(task.RemoteFileName), task.RemoteFileName)
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	logHandler(task.Server, fmt.Sprintf("RESPONSE str [%s]\n", str))

	nextTask := tasks.TypeDownloadArchive
	task.RemoteFileName = tarGzNamefile
	task.LocalFile = path.Join(task.CommandCargo, path.Base(task.RemoteFileName))
	output <- *taskForChannel(&task, str, e, tasks.InProgress, &nextTask)
}
func downloadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	fileProgress := make(chan FileSizeInfo)
	defer close(fileProgress)

	logHandler(task.Server, fmt.Sprintf("Open file [%s] on server [%s]\n", task.RemoteFileName, task.Server))
	sftpClient.RemoveDirectory(path.Dir(task.RemoteFileName))
	srcFile, err := sftpClient.OpenFile(path.Base(task.RemoteFileName), (os.O_RDONLY))
	if err != nil {
		output <- *taskForChannel(&task, "downloadFile - Unable to open remote file", err, tasks.Finished, nil)
		return
	}
	fileInfo, _ := srcFile.Stat()

	logHandler(task.Server, fmt.Sprintf("Create file [%s]\n", task.LocalFile))
	dstFile, err := os.Create(task.LocalFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("downloadFile - Unable to create file [%s]", task.LocalFile), err, tasks.Finished, nil)
		return
	}
	defer dstFile.Close()

	logHandler(task.Server, fmt.Sprintf("DOWNLOADING file[%s] Srv[%s] to[%s]\n", task.RemoteFileName, task.Server, task.LocalFile))
	go printDownloadProgress(fileProgress)
	fileProgress <- FileSizeInfo{FileName: task.LocalFile, Server: task.Server, FileSize: fileInfo.Size()}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to copy file [%s] to [%s]", task.RemoteFileName, task.LocalFile), err, tasks.Finished, nil)
		return
	}

	err = sftpClient.Remove(path.Base(task.RemoteFileName))
	output <- *taskForChannel(&task, "", err, tasks.Finished, nil)
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
func uploadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	destFilePath := path.Join("/var/tmp/", path.Base(task.CommandCargo))

	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	dstFile, err := sftpClient.Create(destFilePath)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return
	}

	srcFile, err := os.Open(task.CommandCargo)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return
	}
	defer srcFile.Close()

	logHandler(task.Server, fmt.Sprintf("UPLOADING file [%s] to [%s] on server [%s]\n", task.CommandCargo, task.RemoteFileName, task.Server))
	_, err = io.Copy(dstFile, srcFile)
	output <- *taskForChannel(&task, fmt.Sprintf("File on remote server[%s]", destFilePath), err, tasks.Finished, nil)
}
