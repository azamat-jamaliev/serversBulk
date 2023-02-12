package pages

import "github.com/rivo/tview"

var app *tview.Application

const (
	PageNameEditEnv string = "PageNameEditEnv"
	PageNameMain    string = "PageNameMain"
	PageNameResults string = "PageNameResults"
)

func newPrimitive(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func newMainPageWithGrid() (*tview.Flex, *tview.Grid) {
	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available"), 0, 0, 1, 2, 0, 0, false)
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(newPrimitive("[ESC]=go back   [Ctrl+C]=to exit"), 1, 0, true)
	return page, grid
}
