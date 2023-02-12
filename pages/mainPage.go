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

	app = appObj

	var searchField, commandField, mtimeField *tview.InputField
	var taskName tasks.TaskType
	var focusOrder []tview.Primitive
	var getNewFocusPrimitive func(direction int) tview.Primitive

	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)

	serversList := tview.NewList()
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

	commandList := tview.NewList().ShowSecondaryText(false).
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
		AddItem("Quit", "", 'q', func() {
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
					doneHandlerFunc(&config.Environments[serversList.GetCurrentItem()],
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
			editHandlerFunc(&config.Environments[serversList.GetCurrentItem()])
		}
		return event
	})

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// if currentPageNmae == "Main" {
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
		// }
		return event
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[Enter]=move to next field  [ESC]=move back  [Ctrl+E]=edit env."), 1, 0, true)

	return page
}
