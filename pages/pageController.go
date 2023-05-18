package pages

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PageController struct {
	focusOrder          []tview.Primitive
	primitiveHint       map[tview.Primitive]string
	app                 *tview.Application
	header              tview.Primitive
	footer              *tview.TextView
	defaultFooterText   string
	lastItemExitHandler func()
	ReloadList          func()
}
type FocusChangeDirection string

func (pCtrl *PageController) addFocus(primitive tview.Primitive) tview.Primitive {
	pCtrl.focusOrder = append(pCtrl.focusOrder, primitive)
	return primitive
}
func (pCtrl *PageController) clearFocus() {
	pCtrl.focusOrder = []tview.Primitive{}
}
func (pCtrl *PageController) setDefaultFooterText() {
	pCtrl.footer.SetText(pCtrl.defaultFooterText)
}
func (pCtrl *PageController) setFooterText(newFootertext string) {
	pCtrl.footer.SetText(fmt.Sprintf("%s; %s", pCtrl.defaultFooterText, newFootertext))
}

func (pCtrl *PageController) setNewFocus(event *tcell.EventKey) {
	d := 0
	processUpDown := true
	processEnter := true
	processLeftRigh := true
	// *TextView
	curAppFocus := pCtrl.app.GetFocus()
	curFocusName := strings.Trim(reflect.ValueOf(curAppFocus).Type().String(), " ")
	if curFocusName == "*tview.List" {
		processLeftRigh = true
		processUpDown = false
	} else if curFocusName == "*tview.TextView" {
		processUpDown = false
		processEnter = false
	} else if curFocusName == "*tview.InputField" {
		processLeftRigh = false
	}
	if event.Key() == tcell.KeyEsc ||
		(event.Key() == tcell.KeyLeft && processLeftRigh) ||
		(event.Key() == tcell.KeyUp && processUpDown) ||
		(event.Key() == tcell.KeyTab && event.Modifiers()&tcell.ModShift != 0) {
		d = -1
	} else if (event.Key() == tcell.KeyEnter && processEnter) ||
		(event.Key() == tcell.KeyRight && processLeftRigh) ||
		(event.Key() == tcell.KeyDown && processUpDown) ||
		event.Key() == tcell.KeyTab {
		d = 1
	}
	if d != 0 {
		for i, item := range pCtrl.focusOrder {
			if item == curAppFocus {
				// Go next page/action only by Enter button
				newFocusNum := i + d
				if newFocusNum >= len(pCtrl.focusOrder) && event.Key() == tcell.KeyEnter && pCtrl.lastItemExitHandler != nil {
					pCtrl.lastItemExitHandler()
				} else if newFocusNum >= 0 && newFocusNum < len(pCtrl.focusOrder) {
					pCtrl.app.SetFocus(pCtrl.focusOrder[newFocusNum])
					if hint, hasHint := pCtrl.primitiveHint[pCtrl.focusOrder[newFocusNum]]; hasHint {
						pCtrl.setFooterText(hint)
					} else {
						pCtrl.setDefaultFooterText()
					}
				}
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
		focusOrder:    []tview.Primitive{},
		app:           appObj,
		primitiveHint: make(map[tview.Primitive]string),
	}
	app = appObj
	return p
}
func NewMainPageController(appObj *tview.Application, lastItemSelectedHandlerFunc func()) (*PageController, *tview.Flex, *tview.Grid) {
	controller := NewPageController(appObj, lastItemSelectedHandlerFunc)
	controller.header = newPrimitive("!!! SeBulk v1.0.4 !!! \nworks when GrayLog or Ansible is not available")
	controller.lastItemExitHandler = lastItemSelectedHandlerFunc
	grid := tview.NewGrid().
		SetRows(2, 0).
		SetColumns(30, 0).
		SetBorders(true).
		AddItem(controller.header, 0, 0, 1, 2, 0, 0, false)

	controller.defaultFooterText = "[ESC]=go back [Ctrl+C]=to exit"
	page, f := NewPageWithFooter(grid, controller.defaultFooterText)
	controller.footer = f

	return controller, page, grid
}
func NewPageWithFooter(mainpart tview.Primitive, footer string) (*tview.Flex, *tview.TextView) {
	f := newPrimitive(footer)
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainpart, 0, 1, true).
		AddItem(f, 1, 0, true)

	return page, f
}
