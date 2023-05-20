package pages

import (
	"fmt"
	"sebulk/modules/tasks"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var serverLogView *tview.TextView
var serverStatusList *tview.List
var getServerLog func(server string) string

func DisplayServerTaskStatus(server, status string) {
	if serverStatusList != nil {
		if serverItems := serverStatusList.FindItems(server, "", true, false); len(serverItems) > 0 {
			_, secondText := serverStatusList.GetItemText(serverItems[0])
			if secondText != string(tasks.Failed) && secondText != string(tasks.Finished) {
				serverStatusList.SetItemText(serverItems[0], server, status)
			}
		} else {
			serverStatusList.AddItem(server, status, 0, nil)
		}
	}
}
func DisplayServerLog(newText string) {
	setServerLogTest(newText)
	app.Draw()
}
func setServerLogTest(newText string) {
	maxSize := 2500
	if len(newText) <= maxSize {
		serverLogView.SetText(newText)
	} else {
		warning := "The content is too big - Please use [CRTL+S] to save full content into files"
		shortnewTest := newText[len(newText)-maxSize:]
		serverLogView.SetText(fmt.Sprintf(">>>>%s \n%s\n>>>>%s", warning, shortnewTest, warning))
	}
}
func ResultsPage(version string, appObj *tview.Application, getServerLogFunc func(server string) string, exitHandlerFunc func(), saveLogsHandlerFunc func()) (tview.Primitive, *PageController) {
	ctrl, page, grid := NewMainPageController(version, appObj, func() {})

	serverLogView = tview.NewTextView()
	serverStatusList = tview.NewList()
	getServerLog = getServerLogFunc
	serverStatusList.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		setServerLogTest(getServerLog(mainText))
	})
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if exitHandlerFunc != nil && event.Key() == tcell.KeyEsc {
			serverStatusList.Clear()
			serverLogView.Clear()
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
