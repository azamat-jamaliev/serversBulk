package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strings"
	"time"

	"serversBulk/modules/configProvider"
	"serversBulk/modules/logHelper"
	"serversBulk/modules/sshHelper"
	"serversBulk/modules/tasks"

	"github.com/fatih/color"
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

func mainOld() {
	logHelper.LogPrintln("********************************************************************************")
	logHelper.LogPrintln("*                           ServersBulk                                        *")
	logHelper.LogPrintln("********************************************************************************")
	logHelper.LogPrintln("see source code in: https://github.com/azamat-jamaliev/serversBulk")

	var taskName tasks.TaskType
	configFileName := flag.String("c", "./config/serversBulk_config.json", "path to environment configuration file")
	serversName := flag.String("servers", "", "to search/download only from the servers with NAME='servers', \n\tfor example if you need to download from SERVER_GROUP_NAME\n\tserevers you can use parameter: `--servers SERVER_GROUP_NAME` ")
	modifTime := flag.String("mtime", "-0.2", "same as mtime for 'find'")
	grepFor := flag.String("s", "", "search string like in:\ngrep --color=auto --mtime -0.2 -H -A2 -B4  \"search string\"")
	executeCmd := flag.String("e", "", "execute given command:\nserversBulk --servers SERVER_GROUP_NAME -e \"curl -v -g http://localhost:28080/api/v1/monitoring/health\"\n\tto get SERVER_GROUP_NAME health from all SERVER_GROUP_NAME nodes")
	localDir := flag.String("d", "", "folder where log files should be downloaded")
	uploadLocalFile := flag.String("u", "", "File which will be uploaded to /var/tmp to the target servers")
	logFilePattern := flag.String("f", "", "log File pattern: i.e. *.log the value will overwrite value in config")
	flag.Parse()

	color.New(color.FgYellow).Println("NOTE: the files are filtered by mTime by default. \nCurrent mTime:%s\n", *modifTime)

	switch {
	case *grepFor != "":
		taskName = tasks.TypeGrepInLogs
	case *executeCmd != "":
		taskName = tasks.TypeExecuteCommand
	case *localDir != "":
		taskName = tasks.TypeArchiveLogs
	case *uploadLocalFile != "":
		taskName = tasks.TypeUploadFile
	default:
		logHelper.ErrFatalln(errors.New(""), "Operation is not defined - please use --help ")
	}

	config := configProvider.GetEnvironemntConfig(configFileName)

	numberOfServers := 0
	tasksChannel := make(chan tasks.ServerTask) //, numberOfServers)
	for _, serverConf := range config.Servers {
		if serverConf.IpAddresses != nil && (*serversName == "" || serverConf.Name == *serversName) {
			numberOfServers += len(serverConf.IpAddresses)
			if *logFilePattern != "" {
				serverConf.LogFilePattern = *logFilePattern
			}
			for _, serverIp := range serverConf.IpAddresses {
				go func(serverIp string, serverConf configProvider.ConfigServerType) {
					task := tasks.ServerTask{Status: tasks.Planned,
						Type:           taskName,
						ConfigFileName: *configFileName,
						ServersName:    *serversName,
						ModifTime:      *modifTime,
						GrepFor:        *grepFor,
						ExecuteCmd:     *executeCmd,
						LocalDir:       *localDir,
						LocalFile:      *uploadLocalFile,
						LogFilePattern: *logFilePattern,
						ConfigServer:   serverConf,
						Server:         serverIp}
					tasksChannel <- task
				}(serverIp, serverConf)

			}
		}
	}
	logHelper.LogPrintf("calculated number of servers to connect=[%v]\n", numberOfServers)
	if numberOfServers < 1 {
		logHelper.ErrFatal(errors.New("servers to connect is less than 1 - nothing to do"))
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
				logHelper.ErrFatalf(errors.New(""), "unknown task type task.Type=%v", task.Type)
			}
		}
		if count >= numberOfServers {
			break
		}
	}

}
func PrintTask(task *tasks.ServerTask) {
	var clr *color.Color
	if task.Status == tasks.Failed {
		clr = color.New(color.FgRed)
		clr = clr.Add(color.Bold)
	} else if task.Status == tasks.Finished {
		clr = color.New(color.FgGreen)
		clr = clr.Add(color.Bold)
	} else {
		clr = color.New(color.BgWhite)
	}
	yellColor := color.New(color.FgYellow)
	yellColor = yellColor.Add(color.Bold)
	clr.Println("")
	clr.Println("↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓")
	yellColor.Printf("SERVER: %s\n", task.Server)
	clr.Printf("NAME: %s\n", task.Type)
	clr.Printf("STATUS: %s\n", task.Status)
	if task.Status == tasks.Failed {
		clr.Println(task.ExecuteCmd)
		clr.Println(task.Error)
	}
	clr.Println(task.RemoteFileName)
	clr.Println(task.Log)
	clr.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
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
	logHelper.LogPrintf("executeOnServer server: %s", server)
	// strOutut = make(chan string)
	sshAdv := sshHelper.OpenSshAdvanced(serverConf, server)
	defer sshAdv.Close()
	logHelper.LogPrintf("execute command:[%s] on ssh server:[%s]", cmd, server)
	buff, e := sshAdv.NewSession().Output(cmd)
	if e != nil {
		logHelper.ErrLogf(e, "error while executing cmd:[%s] os server [%s], cmd_output[%s]", cmd, server, buff)
	}
	str := string(buff)
	return str, e
}

func executeCommand(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}

// NOTE: output result is in the channel
func grepInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	task.ExecuteCmd = fmt.Sprintf("find %s -type f -iname \"%s\" -mtime %s -exec grep --color=auto -H -A15 -B15 -i \"%s\" {}  \\;", task.ConfigServer.LogFolder, task.ConfigServer.LogFilePattern, task.ModifTime, task.GrepFor)
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}
func archiveLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	tarNamefile := fmt.Sprintf("~/%s.%s", strings.ReplaceAll(task.Server, ".", "_"), "tar")
	cmd := fmt.Sprintf("cd %s; find ./ -type f -iname \"%s\" -mtime %s -exec tar rvf %s {} \\;", task.ConfigServer.LogFolder, task.ConfigServer.LogFilePattern, task.ModifTime, tarNamefile)
	str, e := executeOnServer(&task.ConfigServer, task.Server, cmd)

	nextTask := tasks.TypeArchiveGzip
	task.RemoteFileName = tarNamefile
	output <- *taskForChannel(&task, str, e, tasks.InProgress, &nextTask)
}
func archiveGzip(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	tarGzNamefile := fmt.Sprintf("%s.gz", task.RemoteFileName)
	task.ExecuteCmd = fmt.Sprintf("if [ -e %s ]; then cd %s; tar cvzf %s %s ; rm %s;echo 'true';fi", task.RemoteFileName, path.Dir(tarGzNamefile), path.Base(tarGzNamefile), path.Base(task.RemoteFileName), task.RemoteFileName)
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	logHelper.LogPrintf("RESPONSE str [%s]\n", str)

	nextTask := tasks.TypeDownloadArchive
	task.RemoteFileName = tarGzNamefile
	task.LocalFile = path.Join(task.LocalDir, path.Base(task.RemoteFileName))
	output <- *taskForChannel(&task, str, e, tasks.InProgress, &nextTask)
}
func downloadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	fileProgress := make(chan FileSizeInfo)
	defer close(fileProgress)

	logHelper.LogPrintf("Open file [%s] on server [%s]\n", task.RemoteFileName, task.Server)
	sftpClient.RemoveDirectory(path.Dir(task.RemoteFileName))
	srcFile, err := sftpClient.OpenFile(path.Base(task.RemoteFileName), (os.O_RDONLY))
	if err != nil {
		output <- *taskForChannel(&task, "downloadFile - Unable to open remote file", err, tasks.Finished, nil)
		return
	}
	fileInfo, _ := srcFile.Stat()

	logHelper.LogPrintf("Create file [%s]\n", task.LocalFile)
	dstFile, err := os.Create(task.LocalFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("downloadFile - Unable to create file [%s]", task.LocalFile), err, tasks.Finished, nil)
		return
	}
	defer dstFile.Close()

	logHelper.LogPrintf("DOWNLOADING file[%s] Srv[%s] to[%s]\n", task.RemoteFileName, task.Server, task.LocalFile)
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
	// logHelper.LogPrintf("Create file [%s]\n", task.RemoteFileName, task.LocalFile)
	// os.ReadDir(path.Dir(task.LocalFile))
	logHelper.LogPrintln("in printDownloadProgress")
	var fSize *FileSizeInfo
	for {
		select {
		case f, ok := <-fileSizeInfo:
			if !ok {
				break
			}
			fSize = &f
		case <-time.After(1 * time.Second):
			if fSize != nil {
				fileStat, err := os.Stat(fSize.FileName)
				if err != nil {
					logHelper.ErrLogf(err, "printDownloadProgress unable to get stat from file [%s]", fSize.FileName)
				}
				logHelper.LogPrintf("SRV: [%s] ~%v%% downloaded of the file [%s]  \n", fSize.Server, math.Round(100*100*float64(fileStat.Size())/float64(fSize.FileSize))/100, fSize.FileName)
			}
		}
	}
}
func uploadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	destFilePath := path.Join("/var/tmp/", path.Base(task.LocalFile))

	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	dstFile, err := sftpClient.Create(destFilePath)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return
	}

	srcFile, err := os.Open(task.LocalFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return
	}
	defer srcFile.Close()

	logHelper.LogPrintf("UPLOADING file [%s] to [%s] on server [%s]\n", task.LocalFile, task.RemoteFileName, task.Server)
	_, err = io.Copy(dstFile, srcFile)
	output <- *taskForChannel(&task, fmt.Sprintf("File on remote server[%s]", destFilePath), err, tasks.Finished, nil)
}
