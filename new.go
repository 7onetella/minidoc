package minidoc

import (
	"github.com/rivo/tview"
)

type New struct {
	App           *SimpleApp
	Layout        *tview.Flex
	Form          *tview.Form
	debug         func(string)
	IndexHandler  *IndexHandler
	BucketHandler *BucketHandler
}

func NewNewPage(doctype string) *New {
	n := &New{
		Layout: tview.NewFlex(),
	}

	var json interface{}
	var doc MiniDoc
	switch doctype {
	case "url":
		doc = &URLDoc{}
		doc.SetType("url")
		json = Jsonize(doc)
	case "note":
		doc := &NoteDoc{}
		doc.SetType("note")
		json = Jsonize(doc)
	case "todo":
		doc := &ToDoDoc{}
		doc.SetType("todo")
		json = Jsonize(doc)
	}

	n.Form = NewEditorForm(json)
	n.Form.SetBorder(false)
	n.Form.AddButton("Cancel", n.CancelAction)

	return n
}

func (n *New) SetApp(app *SimpleApp) {
	n.App = app
	n.debug = app.DebugView.Debug
}

func (n *New) Page() (title string, content tview.Primitive) {

	n.Layout.AddItem(n.Form, 0, 1, true)
	n.Layout.SetBorder(true).SetBorderPadding(0, 1, 1, 1)

	return "New", n.Layout
}

func (n *New) CreateAction() {
}

func (n *New) Reset() {
	f := n.Form
	f.Clear(true)
	f = tview.NewForm().AddInputField("Title:", "", 60, nil, nil).
		AddInputField("Description:", "", 60, nil, nil).
		AddInputField("Tags:", "", 60, nil, nil)

	//if n.CreateAction != nil {
	//	f.AddButton("Create", n.CreateAction)
	//}
	f.SetBorderPadding(1, 1, 2, 2)
	f.SetBorder(false)
	n.Layout.RemoveItem(n.Form)

	n.Form = f
	n.Layout.AddItem(n.Form, 0, 1, true)
}

func (n *New) CancelAction() {
	n.App.PagesHandler.RemoveLastPage(n.App)
	n.App.PagesHandler.GotoPageByTitle("Search")
	n.App.Draw()
}
