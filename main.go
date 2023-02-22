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

// map[string]string
var currentPageNmae string

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
	fmt.Println(configPath)
	config, err := configProvider.GetFileConfig(configPath)
	if err != nil {
		f := -0.2
		config = configProvider.ConfigFileType{
			DownloadFolder: "~/",
			LogsMtime:      &f,
			Environments: []configProvider.ConfigEnvironmentType{{
				Name: "Example_Env_name",
				Servers: []configProvider.ConfigServerType{{
					Name:           "Server_Group",
					IpAddresses:    []string{"123.123.123.123"},
					LogFolders:     []string{"/var/logs"},
					LogFilePattern: "*.log",
					Login:          "userName",
				}},
			}},
		}
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
		// app.Draw()
	}
	configEditHandler := func(config *configProvider.ConfigEnvironmentType) {
		pagesView.AddAndSwitchToPage(pages.PageNameEditEnv, pages.EditEnvPage(app, config, envExitHandler, envSaveHandler), true)
	}
	configAddHandler := func() {
		envConf := configProvider.ConfigEnvironmentType{}
		config.Environments = append(config.Environments, envConf)
		pagesView.AddAndSwitchToPage(pages.PageNameEditEnv, pages.EditEnvPage(app, &config.Environments[len(config.Environments)-1], envExitHandler, envSaveHandler), true)
	}
	configDoneHandler := func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo string) {
		pagesView.SwitchToPage(pages.PageNameResults)
		resultPageController.SetDefaultFocus()
		go StartTaskForEnv(config, taskName, "", mtime, cargo, ServerLogHandler, ServerTaskStatusHandler)
	}

	resultsPage, resultPageController = pages.ResultsPage(app, GetServerLog)
	mainPage, mainPageController = pages.MainPage(app, &config, configDoneHandler, configEditHandler, configAddHandler)
	mainPageController.SetDefaultFocus()

	if executeWithParams(ServerLogHandler, ServerTaskStatusHandler) {
		pagesView.AddPage(pages.PageNameResults, resultsPage, true, false)
		pagesView.SwitchToPage(pages.PageNameResults)
		resultPageController.SetDefaultFocus()
	} else {
		pagesView.AddPage(pages.PageNameMain, mainPage, true, true)
		pagesView.AddPage(pages.PageNameResults, resultsPage, true, false)
	}
	if err := app.SetRoot(pagesView, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
