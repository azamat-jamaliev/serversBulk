package main

import (
	"flag"
	"fmt"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
)

func mainExecWithParams(newLogHandler func(server, log string),
	newStatusHandler func(server, status string)) bool {
	var taskName tasks.TaskType
	configFileName := flag.String("c", "./config/sebulk_config.json", "path to environment configuration file")
	serversName := flag.String("servers", "", "to search/download only from the servers with NAME='servers', \n\tfor example if you need to download from SERVER_GROUP_NAME\n\tserevers you can use parameter: `--servers SERVER_GROUP_NAME` ")
	modifTime := flag.String("mtime", "-0.2", "same as mtime for 'find'")
	grepFor := flag.String("s", "", "search string like in:\ngrep --color=auto --mtime -0.2 -H -A2 -B4  \"search string\"")
	executeCmd := flag.String("e", "", "execute given command:\nsebulk --servers SERVER_GROUP_NAME -e \"curl -v -g http://localhost:28080/api/v1/monitoring/health\"\n\tto get SERVER_GROUP_NAME health from all SERVER_GROUP_NAME nodes")
	localDir := flag.String("d", "", "folder where log files should be downloaded")
	uploadLocalFile := flag.String("u", "", "File which will be uploaded to /var/tmp to the target servers")
	// logFilePattern := flag.String("f", "", "log File pattern: i.e. *.log the value will overwrite value in config")
	flag.Parse()

	fmt.Printf("!! NOTE: the files are filtered by mTime by default. \nCurrent mTime:%s\n", *modifTime)

	cargo := ""
	cargo2 := "/var/log/"
	switch {
	case *grepFor != "":
		taskName = tasks.TypeGrepInLogs
		cargo = *grepFor
	case *executeCmd != "":
		taskName = tasks.TypeExecuteCommand
		cargo = *executeCmd
	case *localDir != "":
		taskName = tasks.TypeArchiveLogs
		cargo = *localDir
	case *uploadLocalFile != "":
		taskName = tasks.TypeUploadFile
	default:
		return false
	}

	config := configProvider.GetEnvironemntConfig(configFileName)
	go StartTaskForEnv(&config,
		taskName,
		*serversName,
		*modifTime,
		cargo, cargo2, newLogHandler, newStatusHandler)

	return true
}
