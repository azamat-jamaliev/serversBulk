package pages

import (
	"fmt"
	"math"
	"serversBulk/modules/configProvider"
	"serversBulk/modules/tasks"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// *tview.Application, getServerLogFunc func(server string) string)
func MainPage(appObj *tview.Application, config *configProvider.ConfigFileType,
	doneHandlerFunc func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo string),
	editHandlerFunc func(config *configProvider.ConfigEnvironmentType)) tview.Primitive {

	var searchField, commandField, mtimeField *tview.InputField
	var taskName tasks.TaskType
	serversList := tview.NewList()

	lastItemSelectedHandler := func() {
		doneHandlerFunc(&config.Environments[serversList.GetCurrentItem()],
			taskName,
			mtimeField.GetText(),
			commandField.GetText())
	}

	ctrl, page, grid := NewMainPageController(appObj, lastItemSelectedHandler)

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
	mtimeInfoLabel := tview.NewTextView().SetTextAlign(tview.AlignLeft)
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

	commandList := tview.NewList().ShowSecondaryText(false).
		AddItem("Download logs", "", 'd', nil).
		AddItem("Execute command", "", 'e', nil).
		AddItem("Search in logs", "", 's', nil).
		AddItem("Upload file", "", 's', nil)
	listHandlers := []func(){
		func() {
			commandField.SetLabel("download to: ").SetPlaceholder("C:\\temp or ~/Downloads").
				SetText(config.DownloadFolder)
			taskName = tasks.TypeArchiveLogs
		},
		func() {
			commandField.SetLabel("command: ").SetPlaceholder("to execute on servers").
				SetText("")
			taskName = tasks.TypeExecuteCommand
		},
		func() {
			commandField.SetLabel("search: ").SetPlaceholder("text to grep on server").
				SetText("")
			taskName = tasks.TypeGrepInLogs
		},
		func() {
			commandField.SetLabel("file to upload: ").SetPlaceholder("C:\\temp\\1.zip or ~/Downloads/1.zip").
				SetText("")
			taskName = tasks.TypeUploadFile
		}}
	commandList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		listHandlers[index]()
	})
	ctrl.addFocus(commandList)
	listHandlers[0]()

	// mtimeInfoLabel
	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ctrl.addFocus(commandField), 1, 1, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(ctrl.addFocus(mtimeField), 0, 1, true).
			AddItem(mtimeInfoLabel, 0, 0, true), 1, 1, true).
		AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignLeft).
			SetText(" Select Environment:"), 1, 1, true).
		AddItem(ctrl.addFocus(searchField), 1, 1, true).
		AddItem(ctrl.addFocus(serversList), 0, 1, true)

	grid.AddItem(commandList, 1, 0, 1, 1, 0, 20, true).
		AddItem(main, 1, 1, 1, 1, 0, 60, true)

	serversList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlE {
			editHandlerFunc(&config.Environments[serversList.GetCurrentItem()])
		}
		return event
	})

	return page
}
