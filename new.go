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

func NewNew() *New {
	//n := &New{
	//	Layout: tview.NewFlex(),
	//}
	//
	////minidoc := &BaseDoc{}
	//n.Form := nil //NewEdit(minidoc, nil, nil, nil, n.CreateAction).Form
	//n.Form.SetBorder(false)
	return nil
}

func (n *New) SetApp(app *SimpleApp) {
	n.App = app
	n.debug = app.DebugView.Debug
}

// SearchPage returns search page
func (n *New) Page() (title string, content tview.Primitive) {

	//rows := tview.NewFlex().SetDirection(tview.FlexRow).
	//	AddItem(n.Form, 0, 1, true)

	n.Layout.AddItem(n.Form, 0, 1, true)
	n.Layout.SetBorder(true).SetBorderPadding(0, 1, 1, 1)

	return "New", n.Layout

}

func (n *New) CreateAction() {
	f := n.Form
	title := GetInputValue(f, "Title:")
	description := GetInputValue(f, "Description:")
	tags := GetInputValue(f, "Tags:")

	input := &BaseDoc{
		Title:       title,
		Type:        "url",
		Description: description,
		Tags:        tags,
	}

	err := n.UpdateMinidoc(input)
	if err != nil {
		n.debug("error while updating: " + err.Error())
	}

	n.Reset()
	n.App.PagesHandler.GotoPageByTitle("Search")
}

func (n *New) UpdateMinidoc(minidoc *BaseDoc) error {
	key, err := n.App.BucketHandler.Write(minidoc)
	if err != nil {
		return err
	}
	minidoc.ID = key
	err = n.App.IndexHandler.Index(minidoc)
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
