// Demo code for the List primitive.
package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	var serversSearch, commandField, mtimeField *tview.InputField
	var serversList *tview.List
	var commandList *tview.List
	//tview.NewFlex() - Add / remove from flex item in case of different options are selected

	app := tview.NewApplication()

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	// newPrimitive("Main content")

	// sideBar := newPrimitive("Side Bar")

	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)
		// .AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	serversList = tview.NewList().AddItem("Server1", "12.123.123.123", 0, nil).
		AddItem("Server2", "12.123.123.123", 0, nil).
		AddItem("Server3", "12.123.123.123", 0, nil).
		AddItem("Server4", "12.123.123.123", 0, nil).
		AddItem("Server5", "12.123.123.123", 0, nil).
		AddItem("Server6", "12.123.123.123", 0, nil).
		AddItem("Server7", "12.123.123.123", 0, nil).
		AddItem("Server8", "12.123.123.123", 0, nil).
		AddItem("Server9", "12.123.123.123", 0, nil).
		AddItem("Server10", "12.123.123.123", 0, nil).
		AddItem("Server11", "12.123.123.123", 0, nil).
		AddItem("Server12", "12.123.123.123", 0, nil).
		AddItem("Server13", "12.123.123.123", 0, nil).
		AddItem("Server14", "12.123.123.123", 0, nil).
		AddItem("Server15", "12.123.123.123", 0, nil).
		AddItem("Server16", "12.123.123.123", 0, nil).
		AddItem("Server17", "12.123.123.123", 0, nil).
		AddItem("Server18", "12.123.123.123", 0, nil).
		AddItem("Server19", "12.123.123.123", 0, nil).
		AddItem("Server20", "12.123.123.123", 0, nil).
		AddItem("Server21", "12.123.123.123", 0, nil).
		AddItem("Server221", "12.123.123.123", 0, nil).
		AddItem("Server231", "12.123.123.123", 0, nil).
		AddItem("Server241", "12.123.123.123", 0, nil).
		AddItem("Server251", "12.123.123.123", 0, nil).
		AddItem("Server261", "12.123.123.123", 0, nil).
		AddItem("DB1", "321.321.321.321", 0, nil)

	serversSearch = tview.NewInputField().
		SetLabel("Quick search: ").
		SetPlaceholder("server name or IP").
		SetFieldWidth(70).
		SetChangedFunc(func(text string) {
			if found := serversList.FindItems(text, text, false, true); len(found) > 0 {
				serversList.SetCurrentItem(found[0])
			}
		})
	commandField = tview.NewInputField().
		SetLabel("command: ").
		SetPlaceholder("").
		SetFieldWidth(50)
	mtimeField = tview.NewInputField().
		SetLabel("Modified Time: ").
		SetPlaceholder("").
		SetFieldWidth(5).SetText("-0.2")

	commandList = tview.NewList().ShowSecondaryText(false).
		AddItem("Download logs", "", 'd', func() {
			// app.SetFocus(serversSearch)
		}).
		AddItem("Execute command", "", 'e', nil).
		AddItem("Search in logs", "", 's', nil).
		AddItem("Quite", "", 'q', func() {
			app.Stop()
		})

	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(commandField, 0, 1, true).
			AddItem(mtimeField, 0, 1, true), 1, 1, true).
		AddItem(serversSearch, 1, 1, true).
		AddItem(serversList, 0, 1, true)
	// Layout for screens wider than 100 cells.
	grid.AddItem(commandList, 1, 0, 1, 1, 0, 100, true).
		AddItem(main, 1, 1, 1, 1, 0, 100, true)

	findNewFocus := func(focusElements []tview.Primitive, currentFocus tview.Primitive, direction int) int {
		for i := 0; i < len(focusElements); i++ {
			if focusElements[i] == currentFocus {
				result := i + direction
				if result >= len(focusElements) {
					return 0
				} else if result < 0 {
					return len(focusElements) - 1
				}
				return result
			}
		}
		return 0
	}
	focusOrder := []tview.Primitive{commandList,
		commandField,
		mtimeField,
		serversSearch,
		serversList,
	}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetFocus(focusOrder[findNewFocus(focusOrder, app.GetFocus(), -1)])
		} else if event.Key() == tcell.KeyEnter {
			app.SetFocus(focusOrder[findNewFocus(focusOrder, app.GetFocus(), 1)])
		}
		return event
	})
	if err := app.SetRoot(grid, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}
