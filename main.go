package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"sebulk/pages"

	// _ "net/http/pprof"

	"github.com/rivo/tview"
)

var ServerLog, ServerStatus = map[string]string{"": ""}, map[string]string{"": ""}

func ServerTaskStatusHandler(server, status string) {
	ServerStatus[server] = status
	pages.DisplayServerTaskStatus(server, string(status))
}
func ServerLogHandler(server, logRecord string) {
	if val, ok := ServerLog[server]; ok {
		ServerLog[server] = fmt.Sprintf("%s\n%s", val, logRecord)
	} else {
		ServerLog[server] = logRecord
	}
	pages.DisplayServerLog(ServerLog[server])
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
	configPath := filepath.Dir(ex)
	// configPath = "./build"
	configPath = path.Join(configPath, "sebulk_config.json")

	config, err := configProvider.GetFileConfig(configPath)
	if err != nil {
		config = configProvider.GetDefaultConfig()
	}

	envs, err := configProvider.GetAnsibleEnvironmentsConfig(path.Join(filepath.Dir(ex), "./test/ansible-inventory", "CLASSIC"), "30.")
	if err != nil {
		envs, err = configProvider.GetAnsibleEnvironmentsConfig(path.Join(filepath.Dir(ex), "ansible-inventory", "CLASSIC"), "30.")
		if err != nil {
			envs, err = configProvider.GetAnsibleEnvironmentsConfig(path.Join(filepath.Dir(ex), "CLASSIC"), "30.")
		}
	}
	if err == nil {
		config.Environments = append(config.Environments, envs...)
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
		for server, log := range ServerLog {
			if len(server) > 0 {
				fileName := path.Join(config.DownloadFolder, fmt.Sprintf("%s.%s", fileNameFromServerIP(server), "txt"))
				if err := os.WriteFile(fileName, []byte(log), 0644); err != nil {
					panic(err)
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
	if err := app.SetRoot(pagesView, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
