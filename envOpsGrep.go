package main

import (
	"fmt"

	"sebulk/modules/tasks"
)

// NOTE: output result is in the channel
func grepInLogs(task tasks.ServerTask, output chan<- tasks.ServerTask) {
	// TODO: improve serach by showing only data for specified period:
	// get from Linux: date "+%Y-%m-%d %H:%M:%S" !!!
	task.ExecuteCmd = fmt.Sprintf("cd %s", "")
	for _, folder := range task.ConfigServer.LogFolders {
		task.ExecuteCmd = fmt.Sprintf("%s; find %s ! -readable -prune -o -type f -iname \"%s\" -mtime %s -exec grep --color=auto -H -A15 -B15 -i \"%s\" {}  \\;", task.ExecuteCmd, folder, task.ConfigServer.LogFilePattern, task.ModifTime, task.CommandCargo)
	}
	str, e := executeOnServer(&task.ConfigServer, task.Server, task.ExecuteCmd)
	output <- *taskForChannel(&task, str, e, tasks.Finished, nil)
}
