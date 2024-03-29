package main

import (
	"fmt"

	"sebulk/modules/configProvider"
	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

func executeOnServer(serverConf *configProvider.ConfigServerType, server, serverGroup, cmd string) (string, error) {
	statusHandler(server, serverGroup, "CONNECTING...")
	logHandler(server, serverGroup, fmt.Sprintf("connecting to server: [%s] to execute: [%s]", server, cmd))
	sshAdv, err := sshHelper.OpenSshAdvanced(serverConf, server)
	if err != nil {
		logHandler(server, serverGroup, fmt.Sprintf("executeOnServer - OpenSshAdvanced cannot open connection to server:[%s]:\nERROR:%v", server, err))
		return "", err
	}
	defer sshAdv.Close()
	str, e := executeWithConnection(sshAdv, server, serverGroup, cmd)
	return str, e
}
func executeWithConnection(sshAdv *sshHelper.SshAdvanced, server, serverGroup, cmd string) (string, error) {
	statusHandler(server, serverGroup, fmt.Sprintf("EXECUTING.. command:[%s]", cmd))
	logHandler(server, serverGroup, fmt.Sprintf("executing command:[%s] on ssh server:[%s]", cmd, server))
	sess := sshAdv.NewSession()
	buff, e := sess.CombinedOutput(cmd)
	if e != nil {
		logHandler(server, serverGroup, fmt.Sprintf("error while executing cmd:[%s] os server [%s], cmd_output[%s]\nERROR:%v", cmd, server, buff, e))
	}
	str := string(buff)
	return str, e
}
func executeCommand(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ConfigServer.Name, task.CommandCargo)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}
