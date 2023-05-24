package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

func uploadFile(task tasks.ServerTask, output chan<- tasks.ServerTask) error {
	destFilePath := path.Join(task.CommandCargo2, path.Base(task.CommandCargo))

	sshAdv, err := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	if err != nil {
		output <- *taskForChannel(&task, "uploadFile - OpenSshAdvanced cannot open connection", err, tasks.Failed, nil)
		return err
	}
	defer sshAdv.Close()
	sftpClient := sshAdv.NewSftpClient()
	dstFile, err := sftpClient.Create(destFilePath)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return err
	}

	srcFile, err := os.Open(task.CommandCargo)
	if err != nil {
		output <- *taskForChannel(&task, fmt.Sprintf("Unable to create file[%s]", destFilePath), err, tasks.Finished, nil)
		return err
	}
	defer srcFile.Close()

	logHandler(task.Server, task.ConfigServer.Name, fmt.Sprintf("UPLOADING file [%s] to [%s] on server [%s]\n", task.CommandCargo, task.RemoteFileName, task.Server))
	_, err = io.Copy(dstFile, srcFile)
	output <- *taskForChannel(&task, fmt.Sprintf("File on remote server[%s]", destFilePath), err, tasks.Finished, nil)
	return err
}
