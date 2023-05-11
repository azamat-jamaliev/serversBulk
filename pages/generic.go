package pages

import "github.com/rivo/tview"

// var app *tview.Application

const (
	PageNameEditEnv string = "PageNameEditEnv"
	PageNameMain    string = "PageNameMain"
	PageNameResults string = "PageNameResults"
)

func newPrimitive(text string) *tview.TextView {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

var app *tview.Application
