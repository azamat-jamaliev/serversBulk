package main

import (
	"fmt"

	"sebulk/modules/sshHelper"
	"sebulk/modules/tasks"
)

// NOTE: output result is in the channel
func grepInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	statusHandler(task.Server, task.ConfigServer.Name, "CONNECTING...")
	logHandler(task.Server, task.ConfigServer.Name, fmt.Sprintf("connecting to server: [%s] to grep", task.Server))
	strOutput := ""
	sshAdv, err := sshHelper.OpenSshAdvanced(&task.ConfigServer, task.Server)
	if err == nil {
		defer sshAdv.Close()
		if err == nil {
			strGrep := fmt.Sprintf("grep --color=auto -H -A25 -B3 -i \"%s\" {} ", task.CommandCargo)
			task.ExecuteCmd = getFindExecForTask(task, strGrep)
			strOutput, err = executeWithConnection(sshAdv, task.Server, task.ConfigServer.Name, task.ExecuteCmd)
		}
	}
	output <- *taskForChannel(&task, strOutput, err, tasks.Finished, nil)
}
