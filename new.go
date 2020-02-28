package minidoc

import (
	"fmt"
	"github.com/rivo/tview"
)

type New struct {
	App           *SimpleApp
	Layout        *tview.Flex
	Form          *tview.Form
	json          interface{}
	IndexHandler  *IndexHandler
	BucketHandler *BucketHandler
}

func NewNewPage(doc MiniDoc) *New {
	n := &New{
		Layout: tview.NewFlex(),
	}

	log.Debugf("new doctype: %s", doc.GetType())

	n.Form = NewFormWithFields(doc)
	n.Form.SetBorder(false)
	n.Form.AddButton("Create", n.CreateAction)
	n.Form.AddButton("Cancel", n.CancelAction)
	n.json = JsonMapFrom(doc)

	return n
}

func (n *New) SetApp(app *SimpleApp) {
	n.App = app
}

func (n *New) Page() (title string, content tview.Primitive) {

	n.Layout.AddItem(n.Form, 0, 1, true)
	n.Layout.SetBorder(true).SetBorderPadding(0, 1, 1, 1)

	return "New", n.Layout
}

func (n *New) CreateAction() {
	f := n.Form
	jh := NewJsonMapWrapper(n.json)

	ExtractFieldValues(jh, f)

	doc, err := MiniDocFrom(n.json)
	if err != nil {
		log.Errorf("MiniDocFrom failed: %v", err)
		return
	}
	log.Debugf("minidoc from json: %v", n.json)

	id, err := n.App.DataHandler.Write(doc)
	if err != nil {
		log.Errorf("updating %v failed: %v", doc, err)
		return
	}

	n.App.PagesHandler.RemoveLastPage(n.App)
	n.App.PagesHandler.GotoPageByTitle("Search")
	n.App.SetStatus(fmt.Sprintf("[white]%s:%d created[white]", doc.GetType(), id))
	n.App.Draw()
}

func (n *New) CancelAction() {
	n.App.PagesHandler.RemoveLastPage(n.App)
	n.App.PagesHandler.GotoPageByTitle("Search")
	n.App.Draw()
}
