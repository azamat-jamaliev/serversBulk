package main

import (
	"fmt"

	"sebulk/modules/configProvider"
	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

func executeOnServer(serverConf *configProvider.ConfigServerType, server, cmd string) (string, error) {
	statusHandler(server, "CONNECTING...")
	logHandler(server, fmt.Sprintf("connecting to server: [%s] to execute: [%s]", server, cmd))
	sshAdv, err := sshHelper.OpenSshAdvanced(serverConf, server)
	if err != nil {
		logHandler(server, fmt.Sprintf("executeOnServer - OpenSshAdvanced cannot open connection to server:[%s]:\nERROR:%v", server, err))
		return "", err
	}
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
