// Demo code for the List primitive.
package main

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {

	app := tview.NewApplication()
	pages := tview.NewPages()
	configDoneHandler := func() {
		pages.SwitchToPage("Results")
	}
	pages.AddPage("Main", MainPage(app, configDoneHandler), true, true)
	pages.AddPage("Results", ResultsPage(app), true, false)

	if err := app.SetRoot(pages, true).EnableMouse(false).Run(); err != nil {
		panic(err)
	}
}

func MainPage(app *tview.Application, doneHandler func()) tview.Primitive {
	var searchField, commandField, mtimeField *tview.InputField
	var serversList, commandList *tview.List
	var focusOrder []tview.Primitive
	var getNewFocusPrimitive func(direction int) tview.Primitive
	var updateCommandFields func(commandLabel, commandPlaceholder string)
	//tview.NewFlex() - Add / remove from flex item in case of different options are selected

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

	searchField = tview.NewInputField().
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
		SetFieldWidth(60)
	mtimeField = tview.NewInputField().
		SetLabel("Modified Time: ").
		SetPlaceholder("").
		SetAcceptanceFunc(tview.InputFieldFloat).
		SetFieldWidth(5).SetText("-0.2")

	commandList = tview.NewList().ShowSecondaryText(false).
		AddItem("Download logs", "", 'd', func() {
			updateCommandFields("download to: ", "C:\\temp or ~/Downloads")
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Execute command", "", 'e', func() {
			updateCommandFields("command: ", "to execute on servers")
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Search in logs", "", 's', func() {
			updateCommandFields("search: ", "text to grep on server")
			app.SetFocus(getNewFocusPrimitive(1))
		}).
		AddItem("Quite", "", 'q', func() {
			app.Stop()
		})

	main := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(commandField, 0, 1, true).
			AddItem(mtimeField, 0, 1, true), 1, 1, true).
		AddItem(searchField, 1, 1, true).
		AddItem(serversList, 0, 1, true)

	grid.AddItem(commandList, 1, 0, 1, 1, 0, 100, true).
		AddItem(main, 1, 1, 1, 1, 0, 100, true)

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
				if i == len(focusOrder)-1 {
					doneHandler()
					return focusOrder[i]
				}
				result := int(math.Abs(float64(i+direction))) % len(focusOrder)
				return focusOrder[result]
			}
		}
		return focusOrder[0]
	}
	updateCommandFields = func(commandLabel, commandPlaceholder string) {
		commandField.SetLabel(commandLabel)
		commandField.SetPlaceholder(commandPlaceholder)
	}

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			app.SetFocus(getNewFocusPrimitive(-1))
		} else if event.Key() == tcell.KeyEnter && app.GetFocus() != commandList {
			app.SetFocus(getNewFocusPrimitive(1))
		}
		return event
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[Enter]=move to next field     [ESC]=move back"), 1, 0, true)

	return page
}
func ResultsPage(app *tview.Application) tview.Primitive {
	//tview.NewFlex() - Add / remove from flex item in case of different options are selected

	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)
		// .AddItem(newPrimitive("Footer"), 2, 0, 1, 3, 0, 0, false)

	main := newPrimitive("Running logs")
	result := newPrimitive("Status")

	grid.AddItem(result, 1, 0, 1, 1, 0, 100, true).
		AddItem(main, 1, 1, 1, 1, 0, 100, true)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// if event.Key() == tcell.KeyEsc {
		// 	// app.SetFocus(getNewFocusPrimitive(-1))
		// }
		return event
	})

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[ESC]=move back"), 1, 0, true)

	return page
}
