package minidoc

import (
	"fmt"
	"github.com/rivo/tview"
	"strconv"
	"strings"
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
	PrevMenuText  string
}

// GotoPageByTitle goes to page with specified title
func (p *PagesHandler) GotoPageByTitle(title string) {
	log.Debug("go to page by title: " + title)
	indexStr := p.PageIndex[title]
	index, _ := strconv.Atoi(indexStr)
	p.CurrPageIndex = index
	log.Debugf("go to page by title index %d", index)
	p.highlightAndSwitch()
}

func (p *PagesHandler) HasPage(title string) bool {
	//p.Debug("going to " + title)
	_, hasPage := p.PageIndex[title]
	return hasPage
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

func (p *PagesHandler) AddPage(s *SimpleApp, pi PageItem) {
	p.PrevMenuText = strings.Replace(p.MenuBar.GetText(false), "\n", "", -1)
	index := len(s.PagesHandler.PageItems)
	p.PageItems = append(p.PageItems, pi)
	pi.SetApp(s)
	title, primitive := pi.Page()
	log.Debugf("add::previous menu text: %s", p.PrevMenuText)
	text := fmt.Sprintf(`["%d"][darkcyan]%s[white][""]  `, index, title)
	log.Debugf("add::new menu text: %s", text)
	//fmt.Fprintf(p.MenuBar, text)
	newtext := p.PrevMenuText + text
	log.Debugf("add::new menu new text: %s", newtext)
	p.MenuBar.SetText(newtext)
	indexStr := strconv.Itoa(index)
	p.Pages.AddPage(indexStr, primitive, true, true)
	p.PageIndex[title] = indexStr
}

func (p *PagesHandler) RemoveLastPage(s *SimpleApp) {
	index := len(s.PagesHandler.PageItems) - 1
	p.MenuBar.SetText(p.PrevMenuText)
	log.Debugf("remove::previous menu text: %s", p.PrevMenuText)
	indexStr := strconv.Itoa(index)
	log.Debugf("remove page index %d", index)
	pi := p.PageItems[index]
	title, _ := pi.Page()
	p.Pages.RemovePage(indexStr)
	p.PageItems = p.PageItems[:index]
	p.CurrPageIndex = index - 1
	log.Debugf("current index %d", p.CurrPageIndex)
	delete(p.PageIndex, title)
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
	log.Debugf("switching to index %d", p.CurrPageIndex)
	p.MenuBar.Highlight(index).ScrollToHighlight()
	p.Pages.SwitchToPage(index)
}
