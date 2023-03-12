package main

import (
	"fmt"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"testing"
)

func TestEnvOpsManager_NothingToDownload(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	config.Servers[0].LogFilePattern = "long_name_which-cannotbefound.txt"
	StartTaskForEnv(&config,
		tasks.TypeArchiveLogs,
		"",
		"-0.2",
		t.TempDir(), "",
		func(server, log string) {
			fmt.Println(log)
		},
		func(server, status string) {
			fmt.Printf("NothingToDownload Server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`NothingToDownload failed`)
	}
}

func TestEnvOpsManager_Download(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	StartTaskForEnv(&config,
		tasks.TypeExecuteCommand,
		"",
		"-0.2",
		fmt.Sprintf("echo 'test_message' > %s/test.log", config.Servers[0].LogFolders[0]), "",
		func(server, log string) {},
		func(server, status string) {
			fmt.Printf("Execute command Server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`Command Execution Failed`)
	}
	StartTaskForEnv(&config,
		tasks.TypeArchiveLogs,
		"",
		"-0.2",
		t.TempDir(), "",
		func(server, log string) {
			fmt.Println(log)
		},
		func(server, status string) {
			fmt.Printf("Dowloading Server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`Dowloading Failed`)
	}
}
