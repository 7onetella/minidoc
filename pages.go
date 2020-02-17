package minidoc

import (
	"fmt"
	"github.com/rivo/tview"
	"strconv"
)

type SetAppAdapter struct{}

func (a SetAppAdapter) SetApp(app *SimpleApp) {
}

type PageItem interface {
	Page() (title string, content tview.Primitive)
	SetApp(*SimpleApp)
}

// Page object represent a page for Pages
type PageFunc func() (title string, content tview.Primitive)

// PagesHandler handles Pages
type PagesHandler struct {
	Pages         *tview.Pages
	PageIndex     map[string]string
	MenuBar       *tview.TextView
	CurrPageIndex int
	PageItems     []PageItem
}

// GotoPageByTitle goes to page with specified title
func (p *PagesHandler) GotoPageByTitle(title string) {
	//p.Debug("going to " + title)
	indexStr := p.PageIndex[title]
	index, _ := strconv.Atoi(indexStr)
	p.CurrPageIndex = index
	p.highlightAndSwitch()
	//testapp.Draw()
}

// GoToPrevPage goes to previous page
func (p *PagesHandler) GoToPrevPage() {
	p.CurrPageIndex = (p.CurrPageIndex - 1 + len(p.PageItems)) % len(p.PageItems)
	p.highlightAndSwitch()
}

// GoToNextPage goes to next page
func (p *PagesHandler) GoToNextPage() {
	p.CurrPageIndex = (p.CurrPageIndex + 1) % len(p.PageItems)
	p.highlightAndSwitch()
}

// LoadPages loads pages
func (p *PagesHandler) LoadPages(s *SimpleApp) {
	for index, pi := range p.PageItems {
		pi.SetApp(s)
		title, primitive := pi.Page()
		fmt.Fprintf(p.MenuBar, `["%d"][darkcyan]%s[white][""]  `, index, title)
		indexStr := strconv.Itoa(index)
		p.Pages.AddPage(indexStr, primitive, true, index == p.CurrPageIndex)
		p.PageIndex[title] = indexStr
	}
}

// UnloadPages unload pages
func (p *PagesHandler) UnloadPages() {
	for index := range p.PageItems {
		if index == 0 {
			continue
		}
		p.Pages.RemovePage(strconv.Itoa(index))
	}
}

// highlightAndSwitch goes to current page index
func (p *PagesHandler) highlightAndSwitch() {
	index := strconv.Itoa(p.CurrPageIndex)
	p.MenuBar.Highlight(index).ScrollToHighlight()
	p.Pages.SwitchToPage(index)
}
