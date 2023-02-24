package pages

import (
	"sebulk/modules/tasks"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var serverLogView *tview.TextView
var serverStatusList *tview.List
var getServerLog func(server string) string

func DisplayServerTaskStatus(server, status string) {
	if serverItems := serverStatusList.FindItems(server, "", true, false); len(serverItems) > 0 {
		_, secondText := serverStatusList.GetItemText(serverItems[0])
		if secondText != string(tasks.Failed) && secondText != string(tasks.Finished) {
			serverStatusList.SetItemText(serverItems[0], server, status)
		}
	} else {
		serverStatusList.AddItem(server, status, 0, nil)
	}
}
func DisplayServerLog(newText string) {
	setServerLogTest(newText)
	app.Draw()
}
func setServerLogTest(newText string) {
	if len(newText) < 5000 {
		serverLogView.SetText(newText)
	} else {
		serverLogView.SetText("The content is too big - Please use [CRTL+S] to seve it into the file")
	}
}
func ResultsPage(appObj *tview.Application, getServerLogFunc func(server string) string, exitHandlerFunc func(), saveLogsHandlerFunc func()) (tview.Primitive, *PageController) {
	ctrl, page, grid := NewMainPageController(appObj, func() {})

	serverLogView = tview.NewTextView()
	serverStatusList = tview.NewList()
	getServerLog = getServerLogFunc
	serverStatusList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		setServerLogTest(getServerLog(mainText))
	})
	serverStatusList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if exitHandlerFunc != nil && event.Key() == tcell.KeyEsc {
			exitHandlerFunc()
		} else if saveLogsHandlerFunc != nil && event.Key() == tcell.KeyCtrlS {
			saveLogsHandlerFunc()
		}
		return event
	})

	grid.AddItem(ctrl.addFocus(serverStatusList), 1, 0, 1, 1, 0, 20, true).
		AddItem(ctrl.addFocus(serverLogView), 1, 1, 1, 1, 0, 60, true)

	return page, ctrl
}
