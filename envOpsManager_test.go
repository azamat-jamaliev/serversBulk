package main

import (
	"fmt"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
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

func TestEnvOpsManager_Upload(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	StartTaskForEnv(&config,
		tasks.TypeUploadFile,
		"",
		"-0.2",
		"./test/local_test.json", "/var/tmp/",
		func(server, log string) {
			fmt.Printf("Upload file to server: %s, Log:%s", server, log)
		},
		func(server, status string) {
			fmt.Printf("Upload file to server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`Uploading Failed`)
	}
}
func TestEnvOpsManager_Grep(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	StartTaskForEnv(&config,
		tasks.TypeGrepInLogs,
		"",
		"-0.2",
		"var", "",
		func(server, log string) {
			fmt.Printf("GREP on server: %s, Log:%s", server, log)
		},
		func(server, status string) {
			fmt.Printf("GREP on server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`GREP Failed`)
	}
}

func TestEnvOpsManager_Awk(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	StartTaskForEnv(&config,
		tasks.TypeAwkInLogs,
		"",
		"-0.2",
		"var", "",
		func(server, log string) {
			fmt.Printf("AWK on server: %s, Log:%s", server, log)
		},
		func(server, status string) {
			fmt.Printf("AWK on server: %s, status: %s\n", server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`AWK Failed`)
	}
}
