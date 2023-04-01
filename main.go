package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"sebulk/pages"
	"sync"

	// _ "net/http/pprof"

	"github.com/rivo/tview"
)

var ServerLog, ServerStatus = map[string]string{"": ""}, map[string]string{"": ""}
var muStatus sync.Mutex
var muLog sync.Mutex

func ServerTaskStatusHandler(server, status string) {
	muStatus.Lock()
	ServerStatus[server] = status
	pages.DisplayServerTaskStatus(server, string(status))
	muStatus.Unlock()
}
func ServerLogHandler(server, logRecord string) {
	muLog.Lock()
	if val, ok := ServerLog[server]; ok {
		ServerLog[server] = fmt.Sprintf("%s\n%s", val, logRecord)
	} else {
		ServerLog[server] = logRecord
	}
	pages.DisplayServerLog(ServerLog[server])
	muLog.Unlock()
}
func GetServerLog(server string) string {
	if val, ok := ServerLog[server]; ok {
		return val
	}
	return ""
}

func main() {

	// go http.ListenAndServe("localhost:8080", nil)

	// TRACE - have not found how to
	// file, _ := os.Create("./sebulk_trace.out")
	// tracerWriter := bufio.NewWriter(file)
	// trace.Start(tracerWriter)
	// defer trace.Stop()
	var resultsPage, mainPage tview.Primitive
	var mainPageController, resultPageController *pages.PageController
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)

	// os.O_APPEND
	f, err := os.OpenFile(path.Join(exePath, "sebulk.log"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println(">>>>>START<<<<<<")
	configPath := path.Join(exePath, "sebulk_config.json")
	config, err := configProvider.GetFileConfig(configPath)
	if err != nil {
		log.Printf("[WARNING] cannot open config file [%s] ERROR:[%s]\n", configPath, err)
		config = configProvider.GetDefaultConfig()
	}

	envs, err := configProvider.GetAnsibleEnvironmentsConfig(path.Join(exePath, "./test/ansible-inventory", "CLASSIC"), "30.")
	if err != nil {
		log.Printf("[INFO] cannot load Ansible config [./test/ansible-inventory] ERROR:[%s]\n", err)
		envs, err = configProvider.GetAnsibleEnvironmentsConfig(path.Join(exePath, "ansible-inventory", "CLASSIC"), "30.")
		if err != nil {
			log.Printf("[WARNING] cannot load Ansible config [./ansible-inventory] ERROR:[%s]\n", err)
			envs, err = configProvider.GetAnsibleEnvironmentsConfig(path.Join(exePath, "CLASSIC"), "30.")
		}
	}
	if err == nil {
		config.Environments = append(config.Environments, envs...)
	} else {
		log.Printf("[WARNING] cannot load Ansible config [./ansible-inventory] ERROR:[%s]\n", err)
	}

	app := tview.NewApplication()
	pagesView := tview.NewPages()

	envExitHandler := func() {
		pagesView.SwitchToPage(pages.PageNameMain)
		pagesView.RemovePage(pages.PageNameEditEnv)
		mainPageController.SetDefaultFocus()
	}
	envSaveHandler := func() {
		pagesView.SwitchToPage(pages.PageNameMain)
		pagesView.RemovePage(pages.PageNameEditEnv)
		mainPageController.SetDefaultFocus()
		configProvider.SaveFileConfig(&configPath, config)
		mainPageController.ReloadList()
	}
	configEditHandler := func(config *configProvider.ConfigEnvironmentType) {
		pagesView.AddAndSwitchToPage(pages.PageNameEditEnv, pages.EditEnvPage(app, config, envExitHandler, envSaveHandler), true)
	}
	configAddHandler := func() {
		envConf := configProvider.ConfigEnvironmentType{}
		config.Environments = append(config.Environments, envConf)
		pagesView.AddAndSwitchToPage(pages.PageNameEditEnv, pages.EditEnvPage(app, &config.Environments[len(config.Environments)-1], envExitHandler, envSaveHandler), true)
	}
	configDoneHandler := func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo, cargo2 string) {
		pagesView.SwitchToPage(pages.PageNameResults)
		resultPageController.SetDefaultFocus()
		go StartTaskForEnv(config, taskName, "", mtime, cargo, cargo2, ServerLogHandler, ServerTaskStatusHandler)
	}
	saveServerLogHandler := func() {
		for server, srvLog := range ServerLog {
			if len(server) > 0 {
				fileName := path.Join(config.DownloadFolder, fmt.Sprintf("%s.%s", fileNameFromServerIP(server), "txt"))
				if err := os.WriteFile(fileName, []byte(srvLog), 0644); err != nil {
					log.Panicf("[ERROR] cannot save ServerLog file to [%s] ERROR:[%s]\n", fileName, err)
				}
			}
		}
	}

	mainPage, mainPageController = pages.MainPage(app, &config, configDoneHandler, configEditHandler, configAddHandler)
	mainPageController.SetDefaultFocus()

	if mainExecWithParams(ServerLogHandler, ServerTaskStatusHandler) {
		resultsPage, resultPageController = pages.ResultsPage(app, GetServerLog, nil, saveServerLogHandler)
		pagesView.AddPage(pages.PageNameResults, resultsPage, true, true)
		resultPageController.SetDefaultFocus()
	} else {
		resultsPage, resultPageController = pages.ResultsPage(app, GetServerLog, envExitHandler, saveServerLogHandler)
		pagesView.AddPage(pages.PageNameMain, mainPage, true, true)
		pagesView.AddPage(pages.PageNameResults, resultsPage, true, false)
	}
	if err := app.SetRoot(pagesView, true).EnableMouse(true).Run(); err != nil {
		log.Panicf("[ERROR] app.SetRoot failed with ERROR:[%s]\n", err)
	}
}
