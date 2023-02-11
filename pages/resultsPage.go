package pages

import (
	"serversBulk/modules/tasks"

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
		// tasks.Failed
	} else {
		serverStatusList.AddItem(server, status, 0, func() {
			serverLogView.SetText(getServerLog(server))
			app.SetFocus(serverLogView)
		})
		app.SetFocus(serverStatusList)
	}
}
func DisplayServerLog(server, newText string) {
	serverLogView.SetText(newText)
	app.Draw()
}
func ResultsPage(appObj *tview.Application, getServerLogFunc func(server string) string) tview.Primitive {
	app = appObj
	page, grid := newMainPageWithGrid()

	serverLogView = tview.NewTextView()
	serverStatusList = tview.NewList()
	getServerLog = getServerLogFunc

	grid.AddItem(serverStatusList, 1, 0, 1, 1, 0, 20, true).
		AddItem(serverLogView, 1, 1, 1, 1, 0, 60, true)

	serverLogView.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(serverStatusList)
	})

	return page
}
