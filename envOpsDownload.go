package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

func archiveLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) error {
	var sshAdv *sshHelper.SshAdvanced
	var err error
	str := ""

	if len(task.ConfigServer.LogFolders) > 0 {
		statusHandler(task.Server, "CONNECTING...")
		logHandler(task.Server, fmt.Sprintf("archiveLogs - connecting to server: [%s] ", task.Server))
		sshAdv, err = sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
		if err == nil {
			defer sshAdv.Close()

			tarNamefile := fmt.Sprintf("~/%s.%s", fileNameFromServerIP(task.Server), "tar")

			cmd := fmt.Sprintf("cd %s", task.ConfigServer.LogFolders[0])
			for _, folder := range task.ConfigServer.LogFolders {
				cmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" -mtime %s -exec tar rvf %s {} \\;", cmd, folder, task.ConfigServer.LogFilePattern, task.ModifTime, tarNamefile)
			}
			str, err = executeWithConnection(sshAdv, task.Server, cmd)
			logHandler(task.Server, fmt.Sprintf("archiveLogs - cmd: [%s] - result: [%s] ", cmd, str))

			if err == nil {

				cmd = fmt.Sprintf(`if [ ! -f %s ]; then
						echo "NOT_FOUND"
					fi`, tarNamefile)
				str, err = executeWithConnection(sshAdv, task.Server, cmd)
				str = strings.TrimSpace(str)
				logHandler(task.Server, fmt.Sprintf("archiveLogs - cmd: [%s] - result: [%s] ", cmd, str))
				if str != "NOT_FOUND" {
					nextTask := tasks.TypeArchiveGzip
					task.RemoteFileName = tarNamefile
					output <- *taskForChannel(&task, str, err, tasks.InProgress, &nextTask)
					return nil
				} else {
					logHandler(task.Server, "No logs found for the filter")
				}
			}
		}
	} else {
		err = fmt.Errorf("ERROR: 'LogFolders' is empty for [%s] server=[%s]", task.ConfigServer.Name, task.Server)
	}

	output <- *taskForChannel(&task, str, err, tasks.Finished, nil)
	return err
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
func downloadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) error {
	sshAdv, err := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	if err != nil {
		output <- *taskForChannel(&task, "downloadFile - OpenSshAdvanced cannot open connection", err, tasks.Failed, nil)
		return err
	}
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	fileProgress := make(chan FileSizeInfo)
	defer close(fileProgress)

	logHandler(task.Server, fmt.Sprintf("Open file [%s] on server [%s]\n", task.RemoteFileName, task.Server))
	sftpClient.RemoveDirectory(path.Dir(task.RemoteFileName))
	srcFile, err := sftpClient.OpenFile(path.Base(task.RemoteFileName), (os.O_RDONLY))
	if err != nil {
		output <- *taskForChannel(&task, "downloadFile - Unable to open remote file", err, tasks.Failed, nil)
		return err
	}
	fileInfo, _ := srcFile.Stat()

	logHandler(task.Server, fmt.Sprintf("Create file [%s]\n", task.LocalFile))
	dstFile, err := os.Create(task.LocalFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("downloadFile - Unable to create file [%s]", task.LocalFile), err, tasks.Failed, nil)
		return err
	}
	defer dstFile.Close()

	logHandler(task.Server, fmt.Sprintf("DOWNLOADING file[%s] Srv[%s] to[%s]\n", task.RemoteFileName, task.Server, task.LocalFile))
	go printDownloadProgress(fileProgress)
	fileProgress <- FileSizeInfo{FileName: task.LocalFile, Server: task.Server, FileSize: fileInfo.Size()}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to copy file [%s] to [%s]", task.RemoteFileName, task.LocalFile), err, tasks.Failed, nil)
		return err
	}

	err = sftpClient.Remove(path.Base(task.RemoteFileName))
	output <- *taskForChannel(&task, "", err, tasks.Finished, nil)
	return err
}
