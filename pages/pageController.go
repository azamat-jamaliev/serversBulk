package pages

import (
	"math"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type pageController struct {
	focusOrder []tview.Primitive
	app        *tview.Application
	header     tview.Primitive
	footer     tview.Primitive
	// lastItemExitHandler func()
}
type FocusChangeDirection string

func (pCtrl *pageController) addFocus(primitive tview.Primitive) tview.Primitive {
	pCtrl.focusOrder = append(pCtrl.focusOrder, primitive)
	return primitive
}

func (pCtrl *pageController) setNewFocus(event *tcell.EventKey) {
	d := 0
	process := true
	curAppFocus := pCtrl.app.GetFocus()
	if strings.Trim(reflect.ValueOf(curAppFocus).Type().String(), " ") == "*tview.List" {
		process = false
	}
	// fmt.Println("reflect.ValueOf(curAppFocus).Type().String() = [", reflect.ValueOf(curAppFocus).Type().String(), "]")
	if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyRight ||
		(event.Key() == tcell.KeyUp && process) {
		d = -1
	} else if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyLeft ||
		(event.Key() == tcell.KeyDown && process) {
		d = 1
	}
	if d != 0 {
		for i, item := range pCtrl.focusOrder {
			if item == curAppFocus {
				if i+d >= len(pCtrl.focusOrder) {
					// doneHandlerFunc()
					// return focusOrder[0]
				}
				result := int(math.Abs(float64(i+d))) % len(pCtrl.focusOrder)
				pCtrl.app.SetFocus(pCtrl.focusOrder[result])
			}
		}
	}
}

func (pCtrl *pageController) newPrimitive(text string) tview.Primitive {
	return tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText(text)
}

func NewPageController(appObj *tview.Application, lastItemExitHandlerFunc func()) *pageController {
	p := &pageController{
		focusOrder: []tview.Primitive{},
		app:        appObj,
	}
	return p
}
func NewMainPageController(appObj *tview.Application, lastItemSelectedHandlerFunc func()) (*pageController, *tview.Flex, *tview.Grid) {
	controller := NewPageController(appObj, lastItemSelectedHandlerFunc)
	controller.header = controller.newPrimitive("!!! SERVERS BULK !!!\nworkd when Grafana or Ansible is not available")
	controller.footer = controller.newPrimitive("[ESC]=go back   [Ctrl+C]=to exit")
	// controller.lastItemExitHandler = lastItemSelectedHandlerFunc
	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(controller.header, 0, 0, 1, 2, 0, 0, false)
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(grid, 0, 1, true).
		AddItem(controller.footer, 1, 0, true)

	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		controller.setNewFocus(event)
		return event
	})
	return controller, page, grid
}
