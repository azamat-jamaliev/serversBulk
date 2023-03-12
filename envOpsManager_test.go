package main

import (
	"fmt"
	"log"
	"os"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func setTestLogFile(t *testing.T) {
	f, err := os.OpenFile("./test/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		t.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
