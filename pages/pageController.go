package pages

import (
	"math"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PageController struct {
	focusOrder          []tview.Primitive
	app                 *tview.Application
	header              tview.Primitive
	footer              tview.Primitive
	lastItemExitHandler func()
	ReloadList          func()
}
type FocusChangeDirection string

func (pCtrl *PageController) addFocus(primitive tview.Primitive) tview.Primitive {
	pCtrl.focusOrder = append(pCtrl.focusOrder, primitive)
	return primitive
}

func (pCtrl *PageController) setNewFocus(event *tcell.EventKey) {
	d := 0
	processUpDown := true
	processEnter := true
	// *TextView
	curAppFocus := pCtrl.app.GetFocus()
	curFocusName := strings.Trim(reflect.ValueOf(curAppFocus).Type().String(), " ")
	if curFocusName == "*tview.List" {
		processUpDown = false
	} else if curFocusName == "*tview.TextView" {
		processUpDown = false
		processEnter = false
	}
	// fmt.Println("reflect.ValueOf(curAppFocus).Type().String() = [", reflect.ValueOf(curAppFocus).Type().String(), "]")
	if event.Key() == tcell.KeyEsc || event.Key() == tcell.KeyRight ||
		(event.Key() == tcell.KeyUp && processUpDown) {
		d = -1
	} else if (event.Key() == tcell.KeyEnter && processEnter) || event.Key() == tcell.KeyLeft ||
		(event.Key() == tcell.KeyDown && processUpDown) || event.Key() == tcell.KeyTab {
		d = 1
	}
	if d != 0 {
		for i, item := range pCtrl.focusOrder {
			if item == curAppFocus {
				// Go next page/action only by Enter button
				if i+d >= len(pCtrl.focusOrder) && event.Key() == tcell.KeyEnter {
					if pCtrl.lastItemExitHandler != nil {
						pCtrl.lastItemExitHandler()
					}
				}
				result := int(math.Abs(float64(i+d))) % len(pCtrl.focusOrder)
				pCtrl.app.SetFocus(pCtrl.focusOrder[result])
				return
			}
		}
		pCtrl.SetDefaultFocus()
	}
}
func (pCtrl *PageController) SetDefaultFocus() {
	if pCtrl.app != nil {
		if pCtrl.focusOrder != nil && len(pCtrl.focusOrder) > 0 {
			pCtrl.app.SetFocus(pCtrl.focusOrder[0])
		}
	}
	pCtrl.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		pCtrl.setNewFocus(event)
		return event
	})
}

func NewPageController(appObj *tview.Application, lastItemExitHandlerFunc func()) *PageController {
	p := &PageController{
		focusOrder: []tview.Primitive{},
		app:        appObj,
	}
	app = appObj
	return p
}
func NewMainPageController(appObj *tview.Application, lastItemSelectedHandlerFunc func()) (*PageController, *tview.Flex, *tview.Grid) {
	controller := NewPageController(appObj, lastItemSelectedHandlerFunc)
	controller.header = newPrimitive("!!! SeBulk !!! \nworks when GrayLog or Ansible is not available")
	controller.lastItemExitHandler = lastItemSelectedHandlerFunc
	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(controller.header, 0, 0, 1, 2, 0, 0, false)
	page, f := NewPageWithFooter(grid, "[ESC]=go back   [Ctrl+C]=to exit")
	controller.footer = f

	return controller, page, grid
}
func NewPageWithFooter(mainpart tview.Primitive, footer string) (*tview.Flex, tview.Primitive) {
	f := newPrimitive(footer)
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainpart, 0, 1, true).
		AddItem(f, 1, 0, true)

	return page, f
}
