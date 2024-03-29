package main

import (
	"fmt"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"testing"
)

// func TestEnvOpsManager_filterLogLines(t *testing.T) {
// 	// Aug 15, 2022 1:30:23
// 	RunTestForDir(t, "./test/logs/test1", "2022-08-15T01:30:22")
// }
// func TestEnvOpsManager_filterLogLines2(t *testing.T) {
// 	// Aug 15, 2022 1:30:23
// 	RunTestForDir(t, "./test/logs/test2", "2022-08-15T01:52:54")
// }
// func RunTestForDir(t *testing.T, logDir, strDateFrom string) {
// 	f, err := os.OpenFile("./test/test.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
// 	if err != nil {
// 		t.Fatalf("error opening file: %v", err)
// 	}
// 	defer f.Close()
// 	log.SetOutput(f)
// 	log.SetFlags(log.LstdFlags | log.Lshortfile)

// 	log.Printf("RunTestForDir [%s] date[%s]\n", logDir, strDateFrom)

// 	// Aug 15, 2022 12:30:33
// 	timeFrom, err := time.Parse("2006-01-02T15:04:05", strDateFrom)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

//		logDirParent, err := os.ReadDir(logDir)
//		if err == nil {
//			for _, f := range logDirParent {
//				if !f.IsDir() {
//					var file *os.File
//					file, err = os.Open(path.Join(logDir, f.Name()))
//					if err == nil {
//						bytes, _ := io.ReadAll(file)
//						filteredLines := filterLogLines("", string(bytes), timeFrom)
//						log.Println(filteredLines)
//						log.Println("/////////////////////////////////////////")
//					}
//					if err != nil {
//						t.Fatal(err)
//					}
//				}
//			}
//		}
//		if err != nil {
//			t.Fatal(err)
//		}
//	}
func TestEnvOpsManager_Awk(t *testing.T) {
	str := "./test/local_test.json"
	config := configProvider.GetEnvironemntConfig(&str)
	statResult := ""
	StartTaskForEnv(&config,
		tasks.TypeAwkInLogs,
		"",
		"-0.2",
		"var", "",
		func(server, serverGroup, log string) {
			fmt.Printf("AWK on server: %s-%s, Log:%s", serverGroup, server, log)
		},
		func(server, serverGroup, status string) {
			fmt.Printf("AWK on server: %s-%s, status: %s\n", serverGroup, server, status)
			statResult = status
		})
	if statResult != string(tasks.Finished) {
		t.Fatalf(`AWK Failed`)
	}
}
