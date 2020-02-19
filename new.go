package minidoc

import (
	"github.com/rivo/tview"
)

type New struct {
	App           *SimpleApp
	Layout        *tview.Flex
	Form          *tview.Form
	json          interface{}
	debug         func(string)
	IndexHandler  *IndexHandler
	BucketHandler *BucketHandler
}

func NewNewPage(doctype string, app *SimpleApp) *New {
	n := &New{
		Layout: tview.NewFlex(),
	}

	var doc MiniDoc
	switch doctype {
	case "url":
		doc = &URLDoc{}
		doc.SetType("url")
	case "note":
		doc := &NoteDoc{}
		doc.SetType("note")
	case "todo":
		doc := &ToDoDoc{}
		doc.SetType("todo")
	}

	n.Form = NewEditorForm(doc)
	n.Form.SetBorder(false)
	n.Form.AddButton("Create", n.CreateAction)
	n.Form.AddButton("Cancel", n.CancelAction)
	var json interface{}
	json = Jsonize(doc)
	n.json = json

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
	f := n.Form
	jh := NewJSONHandler(n.json)

	ExtractFieldValues(jh, f)

	doc, err := MiniDocFrom(n.json)
	if err != nil {
		log.Errorf("MiniDocFrom failed: %v", err)
		return
	}
	log.Debugf("minidoc from json: %v", n.json)

	err = n.Save(doc)
	if err != nil {
		log.Errorf("updating %v failed: %v", doc, err)
		return
	}

	n.App.PagesHandler.RemoveLastPage(n.App)
	n.App.PagesHandler.GotoPageByTitle("Search")
	n.App.Draw()
}

func (n *New) Save(doc MiniDoc) error {
	_, err := n.App.BucketHandler.Write(doc)
	if err != nil {
		return err
	}
	err = n.App.IndexHandler.Index(doc)
	return err
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
