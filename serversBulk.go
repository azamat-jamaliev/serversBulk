package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"serversBulk/modules/configProvider"
	"serversBulk/modules/logHelper"
	"serversBulk/modules/sshHelper"

	"github.com/fatih/color"
)

type CurrentFileForDownloading struct {
	Server       string
	FileName     string
	ConfigServer configProvider.ConfigServerType
	// SshConfig    *ssh.ClientConfig
	Error error
}

type TaskStatus string

const (
	Planned    TaskStatus = "Planned"
	InProgress TaskStatus = "InProgress"
	Finished   TaskStatus = "Finished"
	Failed     TaskStatus = "Failed"
)

type TaskType string

const (
	TypeUploadFile      TaskType = "UploadFile"
	TypeArchiveLogs     TaskType = "ArchiveLogs"
	TypeArchiveGzip     TaskType = "TypeArchiveGzip"
	TypeDownloadArchive TaskType = "DownloadLogs"
	TypeGrepInLogs      TaskType = "GrepInLogs"
	TypeExecuteCommand  TaskType = "ExecuteCommand"
)

type ServerTask struct {
	Type            TaskType
	Status          TaskStatus
	FileName        string
	Log             string
	JsonFileName    string
	ServersName     string
	ModifTime       string
	GrepFor         string
	ExecuteCmd      string
	LogDir          string
	UploadLocalFile string
	LogFilePattern  string
	ConfigServer    configProvider.ConfigServerType
	Server          string
	Error           error
}

func main() {
	logHelper.LogPrintln("********************************************************************************")
	logHelper.LogPrintln("*                           ServersBulk                                        *")
	logHelper.LogPrintln("********************************************************************************")
	logHelper.LogPrintln("see source code in: https://github.com/azamat-jamaliev/serversBulk")

	var taskName TaskType
	jsonFileName := flag.String("c", "./config/serversBulk_config.json", "path to environment configuration file")
	serversName := flag.String("servers", "", "to search/download only from the servers with NAME='servers', \n\tfor example if you need to download from SERVER_GROUP_NAME\n\tserevers you can use parameter: `--servers SERVER_GROUP_NAME` ")
	modifTime := flag.String("mtime", "-0.2", "same as mtime for 'find'")
	grepFor := flag.String("s", "", "search string like in:\ngrep --color=auto --mtime -0.2 -H -A2 -B4  \"search string\"")
	executeCmd := flag.String("e", "", "execute given command:\nserversBulk --servers SERVER_GROUP_NAME -e \"curl -v -g http://localhost:28080/api/v1/monitoring/health\"\n\tto get SERVER_GROUP_NAME health from all SERVER_GROUP_NAME nodes")
	logDir := flag.String("d", "", "folder where log files should be downloaded")
	uploadLocalFile := flag.String("u", "", "File which will be uploaded to /var/tmp to the target servers")
	logFilePattern := flag.String("f", "", "log File pattern: i.e. *.log the value will overwrite value in config")
	flag.Parse()

	color.New(color.FgYellow).Println("NOTE: the files are filtered by mTime by default. \nCurrent mTime:%s\n", *modifTime)

	switch {
	case *grepFor != "":
		taskName = TypeGrepInLogs
	case *executeCmd != "":
		taskName = TypeExecuteCommand
	case *logDir != "":
		taskName = TypeArchiveLogs
	case *uploadLocalFile != "":
		taskName = TypeUploadFile
	default:
		logHelper.ErrFatalWithMessage("Operation is not defined - please use --help ", errors.New(""))
	}

	config := configProvider.GetConfig(jsonFileName)

	numberOfServers := 0
	tasksChannel := make(chan ServerTask) //, numberOfServers)
	for _, serverConf := range config.Servers {
		if serverConf.IpAddresses != nil && (*serversName == "" || serverConf.Name == *serversName) {
			numberOfServers += len(serverConf.IpAddresses)
			if *logFilePattern != "" {
				serverConf.LogFilePattern = *logFilePattern
			}
			for _, serverIp := range serverConf.IpAddresses {
				go func(serverIp string, serverConf configProvider.ConfigServerType) {
					task := ServerTask{Status: Planned,
						Type:            taskName,
						JsonFileName:    *jsonFileName,
						ServersName:     *serversName,
						ModifTime:       *modifTime,
						GrepFor:         *grepFor,
						ExecuteCmd:      *executeCmd,
						LogDir:          *logDir,
						UploadLocalFile: *uploadLocalFile,
						LogFilePattern:  *logFilePattern,
						ConfigServer:    serverConf,
						Server:          serverIp}
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
		if task.Status == Finished || task.Status == Failed {
			count++
			//Print finished task:
			PrintTask(&task)
		} else {
			//print Inprogress task
			switch task.Type {
			case TypeGrepInLogs:
				go grepInLogs(task, tasksChannel)
			case TypeExecuteCommand:
				go executeCommand(task, tasksChannel)
			case TypeArchiveLogs:
				go archiveLogs(task, tasksChannel)
			case TypeArchiveGzip:
				go archiveGzip(task, tasksChannel)
			case TypeDownloadArchive:
				go downloadZipLog(task, tasksChannel)
			case TypeUploadFile:
				go uploadFile(task, tasksChannel)
			default:
				logHelper.ErrFatalWithMessage(fmt.Sprintf("unknown task type task.Type=%v", task.Type), errors.New(""))
			}
		}
		if count >= numberOfServers {
			break
		}
	}

}
func PrintTask(task *ServerTask) {
	var clr *color.Color
	if task.Status == Failed {
		clr = color.New(color.FgRed)
		clr = clr.Add(color.Bold)
	} else if task.Status == Finished {
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
	if task.Status == Failed {
		clr.Println(task.ExecuteCmd)
		clr.Println(task.Error)
	}
	clr.Println(task.FileName)
	clr.Println(task.Log)
	clr.Println("↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑")
}

func taskForChannel(task *ServerTask, log string, err error, newStatus TaskStatus, nextTask *TaskType) *ServerTask {
	task.Log = log
	task.Error = err
	if err != nil {
		task.Status = Failed
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
		logHelper.ErrLogWinMessage(fmt.Sprintf("error while executing cmd:[%s] os server [%s], cmd_output[%s]", cmd, server, buff), e)
	}
	str := string(buff)
	return str, e
}

func executeCommand(task ServerTask, output chan<- ServerTask) {
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, Finished, nil)
}

// NOTE: output result is in the channel
func grepInLogs(task ServerTask, output chan<- ServerTask) {
	task.ExecuteCmd = fmt.Sprintf("find %s -type f -iname \"%s\" -mtime %s -exec grep --color=auto -H -A15 -B15 \"%s\" {}  \\;", task.ConfigServer.LogFolder, task.ConfigServer.LogFilePattern, task.ModifTime, task.GrepFor)
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, Finished, nil)
}
func archiveLogs(task ServerTask, output chan<- ServerTask) {
	tarNamefile := fmt.Sprintf("%s/%s.%s", task.ConfigServer.LogFolder, strings.ReplaceAll(task.Server, ".", "_"), "tar")
	cmd := fmt.Sprintf("cd %s; find ./ -type f -iname \"%s\" -mtime %s -exec tar rvf %s {} \\;", task.ConfigServer.LogFolder, task.ConfigServer.LogFilePattern, task.ModifTime, tarNamefile)
	str, e := executeOnServer(&task.ConfigServer, task.Server, cmd)

	nextTask := TypeArchiveGzip
	task.FileName = tarNamefile
	output <- *taskForChannel(&task, str, e, InProgress, &nextTask)
}
func archiveGzip(task ServerTask, output chan<- ServerTask) {
	tarGzNamefile := fmt.Sprintf("%s.gz", task.FileName)
	task.ExecuteCmd = fmt.Sprintf("if test -f '%s'; then cd %s; tar cvzf %s %s ; rm %s;echo 'true';fi", task.FileName, path.Dir(tarGzNamefile), path.Base(tarGzNamefile), path.Base(task.FileName), task.FileName)
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	logHelper.LogPrintf("RESPONSE str [%s]\n", str)

	nextTask := TypeDownloadArchive
	task.FileName = tarGzNamefile
	output <- *taskForChannel(&task, str, e, InProgress, &nextTask)
}
func downloadZipLog(task ServerTask, output chan<- ServerTask) {
	localZipFileName := path.Join(task.LogDir, path.Base(task.FileName))

	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()

	logHelper.LogPrintf("Open file [%s] on server [%s]\n", task.FileName, task.Server)
	srcFile, err := sftpClient.OpenFile(task.FileName, (os.O_RDONLY))
	if err != nil {
		output <- *taskForChannel(&task, "Unable to download file", err, Finished, nil)
		return
	}
	logHelper.LogPrintf("Create file [%s]\n", task.FileName, localZipFileName)
	dstFile, err := os.Create(localZipFileName)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file [%s]", localZipFileName), err, Finished, nil)
		return
	}
	defer dstFile.Close()

	logHelper.LogPrintf("DOWNLOADING file [%s] to [%s]\n", task.FileName, localZipFileName)
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to copy file [%s] to [%s]", task.FileName, localZipFileName), err, Finished, nil)
		return
	}

	err = sftpClient.Remove(task.FileName)
	output <- *taskForChannel(&task, "", err, Finished, nil)
}
func uploadFile(task ServerTask, output chan<- ServerTask) {
	destFilePath := path.Join("/var/tmp/", path.Base(task.UploadLocalFile))

	sshAdv := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	dstFile, err := sftpClient.Create(destFilePath)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, Finished, nil)
		return
	}

	srcFile, err := os.Open(task.UploadLocalFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, Finished, nil)
		return
	}
	defer srcFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	output <- *taskForChannel(&task, fmt.Sprintf("File on remote server[%s]", destFilePath), err, Finished, nil)
}
