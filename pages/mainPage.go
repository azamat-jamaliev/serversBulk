package pages

import (
	"fmt"
	"math"
	"sebulk/modules/configProvider"
	"sebulk/modules/tasks"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// *tview.Application, getServerLogFunc func(server string) string)
func MainPage(appObj *tview.Application, config *configProvider.ConfigFileType,
	doneHandlerFunc func(config *configProvider.ConfigEnvironmentType, taskName tasks.TaskType, mtime, cargo, cargo2 string),
	editHandlerFunc func(config *configProvider.ConfigEnvironmentType),
	addHandlerFunc func()) (tview.Primitive, *PageController) {

	var searchField, commandField, mtimeField, uploadToField *tview.InputField
	var taskName tasks.TaskType
	serversList := tview.NewList()

	lastItemSelectedHandler := func() {
		doneHandlerFunc(&config.Environments[serversList.GetCurrentItem()],
			taskName,
			mtimeField.GetText(),
			commandField.GetText(),
			uploadToField.GetText())
	}

	ctrl, page, grid := NewMainPageController(appObj, lastItemSelectedHandler)

	ctrl.ReloadList = func() {
		serversList.Clear()
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
	}
	ctrl.ReloadList()

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
		SetFieldWidth(80)
	uploadToField = tview.NewInputField().
		SetLabel("upload to: ").
		SetPlaceholder("folder to upload i.e. /var/tmp").
		SetFieldWidth(80).SetText(config.UploadFolder)
	mtimeInfoLabel := tview.NewTextView().SetTextAlign(tview.AlignLeft)
	mtimeField = tview.NewInputField().
		SetLabel("less than: ").
		SetPlaceholder("").
		SetAcceptanceFunc(tview.InputFieldFloat).
		SetFieldWidth(5).SetText("-0.2").
		SetChangedFunc(func(text string) {
			if value, err := strconv.ParseFloat(text, 64); err == nil {
				mtimeInfoLabel.SetText(fmt.Sprintf("~%vh. ago", math.Round(math.Abs(24*float64(value)))))
				if value > 0 {
					mtimeField.SetLabel("more than: ")
				} else {
					mtimeField.SetLabel("less than: ")
				}
			}
		})
	if config.LogsMtime != nil {
		mtimeField.SetText(fmt.Sprintf("%v", *config.LogsMtime))
	}

	commandList := tview.NewList().ShowSecondaryText(false).
		AddItem("Download logs", "", 'd', nil).
		AddItem("Execute command", "", 'e', nil).
		AddItem("Search in logs", "", 's', nil).
		// AddItem("Quick check errors", "", 'q', nil).
		AddItem("Upload file", "", 'u', nil)

	main := tview.NewFlex()
	clearFields := func() {
		ctrl.clearFocus()
		main.Clear()
		ctrl.addFocus(commandList)
	}
	addServersList := func() {
		main.AddItem(tview.NewTextView().
			SetTextAlign(tview.AlignLeft).
			SetText(" Select Environment:"), 1, 1, true).
			AddItem(ctrl.addFocus(searchField), 1, 1, true).
			AddItem(ctrl.addFocus(serversList), 0, 1, true)
	}
	listHandlers := []func(){
		func() {
			clearFields()
			main.SetDirection(tview.FlexRow).
				AddItem(ctrl.addFocus(commandField), 1, 1, true).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(ctrl.addFocus(mtimeField), 19, 1, true).
					AddItem(mtimeInfoLabel, 0, 1, true), 1, 1, true)
			addServersList()
			commandField.SetLabel("download to: ").SetPlaceholder("C:\\temp or ~/Downloads").
				SetText(config.DownloadFolder)
			taskName = tasks.TypeArchiveLogs
		},
		func() {
			clearFields()
			main.SetDirection(tview.FlexRow).
				AddItem(ctrl.addFocus(commandField), 1, 1, true)
			commandField.SetLabel("command: ").SetPlaceholder("to execute on servers").
				SetText("")
			addServersList()
			taskName = tasks.TypeExecuteCommand
		},
		func() {
			clearFields()
			main.SetDirection(tview.FlexRow).
				AddItem(ctrl.addFocus(commandField), 1, 1, true).
				AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
					AddItem(ctrl.addFocus(mtimeField), 19, 1, true).
					AddItem(mtimeInfoLabel, 0, 1, true), 1, 1, true)
			addServersList()
			commandField.SetLabel("search: ").SetPlaceholder("text to grep on server").
				SetText("")
			taskName = tasks.TypeGrepInLogs
		},
		// func() {
		// 	clearFields()
		// 	main.SetDirection(tview.FlexRow).
		// 		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
		// 			AddItem(ctrl.addFocus(mtimeField), 19, 1, true).
		// 			AddItem(mtimeInfoLabel, 0, 1, true), 1, 1, true)
		// 	addServersList()
		// 	taskName = tasks.TypeAwkInLogs
		// },
		func() {
			clearFields()
			main.SetDirection(tview.FlexRow).
				AddItem(ctrl.addFocus(commandField), 1, 1, true).
				AddItem(ctrl.addFocus(uploadToField), 1, 1, true)
			addServersList()
			commandField.SetLabel("file to upload: ").SetPlaceholder("C:\\temp\\1.zip or ~/Downloads/1.zip").
				SetText("")
			taskName = tasks.TypeUploadFile
		}}
	commandList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		listHandlers[index]()
	})
	listHandlers[0]()

	grid.AddItem(commandList, 1, 0, 1, 1, 0, 20, true).
		AddItem(main, 1, 1, 1, 1, 0, 60, true)

	serversList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlE {
			editHandlerFunc(&config.Environments[serversList.GetCurrentItem()])
		} else if event.Key() == tcell.KeyCtrlA {
			addHandlerFunc()
		}
		return event
	})

	return page, ctrl
}
