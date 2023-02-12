package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"serversBulk/modules/configProvider"
	"serversBulk/modules/tasks"
	"serversBulk/pages"

	// _ "net/http/pprof"

	"github.com/gdamore/tcell/v2"
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
	pages.DisplayServerLog(server, ServerLog[server])
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
	// file, _ := os.Create("./serversBulk_trace.out")
	// tracerWriter := bufio.NewWriter(file)
	// trace.Start(tracerWriter)
	// defer trace.Stop()

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	configPath := filepath.Dir(ex)
	// configPath = "./build"
	configPath = path.Join(configPath, "serversBulk_config.json")
	fmt.Println(configPath)
	config, err := configProvider.GetFileConfig(configPath)
	if err != nil {
		panic(err)
	}

	app := tview.NewApplication()
	pagesView := tview.NewPages()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
		}
		return event
	})

	envExitHandler := func() {
		pagesView.SwitchToPage(pages.PageNameMain)
		pagesView.RemovePage(pages.PageNameEditEnv)
	}
	configEditHandler := func(config *configProvider.ConfigEnvironmentType) {
		pagesView.AddAndSwitchToPage(pages.PageNameEditEnv, pages.EditEnvPage(app, config, envExitHandler), true)
	}
	configDoneHandler := func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo string) {
		pagesView.SwitchToPage(pages.PageNameResults)
		go StartTaskForEnv(config, taskName, "", mtime, cargo, ServerLogHandler, ServerTaskStatusHandler)
	}

	if executeWithParams(ServerLogHandler, ServerTaskStatusHandler) {
		pagesView.AddPage(pages.PageNameResults, pages.ResultsPage(app, GetServerLog), true, false)
		pagesView.SwitchToPage(pages.PageNameResults)
	} else {
		pagesView.AddPage(pages.PageNameMain, pages.MainPage(app, &config, configDoneHandler, configEditHandler), true, true)
		pagesView.AddPage(pages.PageNameResults, pages.ResultsPage(app, GetServerLog), true, false)
	}
	if err := app.SetRoot(pagesView, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
