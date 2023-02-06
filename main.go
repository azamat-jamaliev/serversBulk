package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"path/filepath"
	"serversBulk/modules/configProvider"
	"serversBulk/modules/tasks"
	"strconv"
	"strings"

	// _ "net/http/pprof"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var app *tview.Application
var serverLogView *tview.TextView
var serverStatusList *tview.List
var serversList, commandList *tview.List

var ServerLog map[string]string
var currentPageNmae string

func main() {

	// go http.ListenAndServe("localhost:8080", nil)

	// TRACE - have not found how to
	// file, _ := os.Create("./serversBulk_trace.out")
	// tracerWriter := bufio.NewWriter(file)
	// trace.Start(tracerWriter)
	// defer trace.Stop()

	currentPageNmae = "Main"
	ServerLog = map[string]string{"": ""}
	// ServerStatus = map[string]string{"": ""}
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	configPath := filepath.Dir(ex)
	// configPath = "./build"
	configPath = path.Join(configPath, "serversBulk_config.json")
	fmt.Println(configPath)
	config := configProvider.GetFileConfig(&configPath)

	app = tview.NewApplication()
	pages := tview.NewPages()
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPageNmae == "Reults" {
			if event.Key() == tcell.KeyCtrlC {
				app.Stop()
			}
		}
		return event
	})
	configDoneHandler := func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo string) {
		if currentPageNmae == "Main" {
			currentPageNmae = "Results"
			pages.SwitchToPage("Results")
			go StartTaskForEnv(config, taskName, "", mtime, cargo, ServerLogHandler, ServerTaskStatusHandler)
		}
	}
	envExitHandler := func() {
		if currentPageNmae == "EditEnvironment" {
			currentPageNmae = "Main"
			pages.SwitchToPage("Main")
			app.SetFocus(commandList)
			pages.RemovePage("EditEnvironment")
		}
	}
	configEditHandler := func(config *configProvider.ConfigEnvironmentType) {
		if currentPageNmae == "Main" {
			currentPageNmae = "EditEnvironment"
			pages.AddAndSwitchToPage("EditEnvironment", EditEnvPage(app, config, envExitHandler), true)
		}
	}

	pages.AddPage("Main", MainPage(app, &config, configDoneHandler, configEditHandler), true, true)
	pages.AddPage("Results", ResultsPage(app), true, false)
	// pages.AddPage("EditEnvironment", EditEnvPage(app), true, false)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
func ServerTaskStatusHandler(server, status string) {
	if serverItems := serverStatusList.FindItems(server, "", true, false); len(serverItems) > 0 {
		_, secondText := serverStatusList.GetItemText(serverItems[0])
		if secondText != string(tasks.Failed) && secondText != string(tasks.Finished) {
			serverStatusList.SetItemText(serverItems[0], server, status)
		}
		// tasks.Failed
	} else {
		serverStatusList.AddItem(server, status, 0, func() {
			serverLogView.SetText(ServerLog[server])
			app.SetFocus(serverLogView)
		})
		app.SetFocus(serverStatusList)
	}
}
func ServerLogHandler(server, log string) {
	val, ok := ServerLog[server]
	if ok {
		ServerLog[server] = fmt.Sprintf("%s\n%s", val, log)
	} else {
		ServerLog[server] = log
	}
	newText := fmt.Sprintf("%s\n%s", serverLogView.GetText(false), log)
	serverLogView.SetText(newText)
	app.Draw()
}

func newPrimitive(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func MainPage(app *tview.Application, config *configProvider.ConfigFileType,
	doneHandler func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo string),
	editHandler func(config *configProvider.ConfigEnvironmentType)) tview.Primitive {
	var searchField, commandField, mtimeField *tview.InputField
	var taskName tasks.TaskType
	var focusOrder []tview.Primitive
	var getNewFocusPrimitive func(direction int) tview.Primitive

	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)

	serversList = tview.NewList()
	for _, env := range config.Environments {
		serverLine := ""
		for _, server := range env.Servers {
			ipLine := ""
			for _, ipAddr := range server.IpAddresses {
				ipLine = fmt.Sprintf("%s %s", ipLine, ipAddr)
			}
			serverLine = fmt.Sprintf("%s %s=[%s]", serverLine, server.Name, ipLine)
		}
		serversList.AddItem(env.Name, serverLine, 0, nil)
	}

	searchField = tview.NewInputField().
		SetLabel("Quick search: ").
		SetPlaceholder("environment name or server IP").
		SetFieldWidth(40).
		SetChangedFunc(func(text string) {
			if found := serversList.FindItems(text, text, false, true); len(found) > 0 {
				serversList.SetCurrentItem(found[0])
			}
		})

	commandField = tview.NewInputField().
		SetLabel("command: ").
		SetPlaceholder("").
		SetFieldWidth(30)
	mtimeInfoLabel := tview.NewTextView()
	mtimeField = tview.NewInputField().
		SetLabel("less than: ").
		SetPlaceholder("").
		SetAcceptanceFunc(tview.InputFieldFloat).
		SetFieldWidth(5).SetText("-0.2").
		SetChangedFunc(func(text string) {
			if value, err := strconv.ParseFloat(text, 64); err == nil {
				mtimeInfoLabel.SetText(fmt.Sprintf("~%vh. ago", math.Round(math.Abs(24*float64(value)))))
			}
		})
	if config.LogsMtime != nil {
		mtimeField.SetText(fmt.Sprintf("%v", *config.LogsMtime))
	}

	commandList = tview.NewList().ShowSecondaryText(false).
		AddItem("Download logs", "", 'd', func() {
			commandField.SetLabel("download to: ").SetPlaceholder("C:\\temp or ~/Downloads").
				SetText(config.DownloadFolder)
			taskName = tasks.TypeArchiveLogs
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Execute command", "", 'e', func() {
			commandField.SetLabel("command: ").SetPlaceholder("to execute on servers")
			taskName = tasks.TypeExecuteCommand
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Search in logs", "", 's', func() {
			commandField.SetLabel("search: ").SetPlaceholder("text to grep on server")
			taskName = tasks.TypeGrepInLogs
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Upload file", "", 's', func() {
			commandField.SetLabel("file to upload: ").SetPlaceholder("C:\\temp\\1.zip or ~/Downloads/1.zip")
			taskName = tasks.TypeUploadFile
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Quite", "", 'q', func() {
			app.Stop()
		})
		// mtimeInfoLabel
	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(commandField, 1, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(mtimeField, 0, 1, true).
			AddItem(mtimeInfoLabel, 0, 1, true), 1, 1, true).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignLeft).
			SetText(" Select Environment:"), 1, 1, true).
		AddItem(searchField, 1, 1, true).
		AddItem(serversList, 0, 1, true)

	grid.AddItem(commandList, 1, 0, 1, 1, 0, 20, true).
		AddItem(main, 1, 1, 1, 1, 0, 60, true)

	focusOrder = []tview.Primitive{commandList,
		commandField,
		mtimeField,
		searchField,
		serversList,
	}
	getNewFocusPrimitive = func(direction int) tview.Primitive {
		curAppFocus := app.GetFocus()
		for i := 0; i < len(focusOrder); i++ {
			if focusOrder[i] == curAppFocus {
				if i+direction >= len(focusOrder) {
					doneHandler(&config.Environments[serversList.GetCurrentItem()],
						taskName,
						mtimeField.GetText(),
						commandField.GetText())

					return focusOrder[0]
				}
				result := int(math.Abs(float64(i+direction))) % len(focusOrder)
				return focusOrder[result]
			}
		}
		return focusOrder[0]
	}
	serversList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlE {
			editHandler(&config.Environments[serversList.GetCurrentItem()])
		}
		return event
	})

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if currentPageNmae == "Main" {
			if event.Key() == tcell.KeyEsc {
				app.SetFocus(getNewFocusPrimitive(-1))
			} else if event.Key() == tcell.KeyEnter && app.GetFocus() != commandList {
				app.SetFocus(getNewFocusPrimitive(1))
			} else if event.Key() == tcell.KeyDown {
				if app.GetFocus() == searchField || app.GetFocus() == commandField || app.GetFocus() == mtimeField {
					app.SetFocus(getNewFocusPrimitive(1))
				}
			} else if event.Key() == tcell.KeyUp {
				if app.GetFocus() == searchField || app.GetFocus() == mtimeField {
					app.SetFocus(getNewFocusPrimitive(-1))
				}
			}
		}
		return event
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[Enter]=move to next field  [ESC]=move back  [Ctrl+E]=edit env."), 1, 0, true)

	return page
}
func ResultsPage(app *tview.Application) tview.Primitive {
	//tview.NewFlex() - Add / remove from flex item in case of different options are selected

	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)
		// .AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	serverLogView = tview.NewTextView()
	serverStatusList = tview.NewList()

	grid.AddItem(serverStatusList, 1, 0, 1, 1, 0, 20, true).
		AddItem(serverLogView, 1, 1, 1, 1, 0, 60, true)

	serverLogView.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(serverStatusList)
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[ESC]=go back   [Ctrl+C]=to exit"), 1, 0, true)

	return page
}

var editEnvForm *tview.Form

func EditEnvPage(app *tview.Application, env *configProvider.ConfigEnvironmentType, exitHandler func()) tview.Primitive {
	editEnvForm = tview.NewForm()
	editEnvForm.SetBorder(true).SetTitle("Environment Information")

	modifyEnv := env
	if env == nil {
		modifyEnv = &configProvider.ConfigEnvironmentType{}
	}
	editEnvForm.AddInputField("Environment name:", modifyEnv.Name, 20, nil, nil)
	for _, srv := range modifyEnv.Servers {
		editEnvForm.AddInputField("Server Name: ", srv.Name, 20, nil, nil).
			AddTextArea("Log Folders:", strings.Join(srv.LogFolders[:], "\n"), 0, 3, 0, nil).
			AddInputField("Log File Pattern: ", srv.LogFilePattern, 7, nil, nil).
			AddInputField("ssh login: ", srv.Login, 20, nil, nil).
			AddPasswordField("ssh password: ", srv.Passowrd, 20, '*', nil).
			AddInputField("ssh identity file: ", srv.IdentityFile, 60, nil, nil).
			AddTextArea("Servers IPs/Names:", strings.Join(srv.IpAddresses[:], "\n"), 0, 3, 0, nil).
			AddCheckbox("Use Bastion (ssh tunnel):", srv.BastionServer != "", nil)
		if srv.BastionServer != "" {
			editEnvForm.AddInputField("Bastion server IP/name: ", srv.BastionServer, 20, nil, nil).
				AddInputField("Bastion ssh login: ", srv.BastionLogin, 20, nil, nil).
				AddPasswordField("Bastion ssh password: ", srv.BastionPassword, 20, '*', nil).
				AddInputField("Bastion Identity File: ", srv.BastionIdentityFile, 60, nil, nil)
		}
	}
	editEnvForm.AddButton(" -      Add Servers        - ", nil).
		AddButton("Save", nil).
		AddButton("Cancel", func() {
			exitHandler()
		})

	return editEnvForm
}
