package minidoc

import (
	"fmt"
	"github.com/rivo/tview"
)

type Help struct {
	App     *SimpleApp
	debug   func(string)
	Content *tview.TextView
}

func NewHelp() *Help {
	h := &Help{
		Content: tview.NewTextView(),
	}
	h.Content.SetDynamicColors(true).SetBorder(true)
	return h
}

func (h *Help) SetApp(app *SimpleApp) {
	h.App = app
	h.debug = app.DebugView.Debug
}

func (h *Help) Page() (title string, content tview.Primitive) {
	fmt.Fprintf(h.Content, `

    [yellow]List of shortcut keys for the following scope[white]

    [darkcyan]Entire App[white]

       Ctrl-C  <-  Exit
       Ctrl-H  <-  Navigate to left menu item
       Ctrl-L  <-  Navigate to right menu item
       Ctrl-O  <-  Show debug view
       Ctrl-D  <-  Go to debug view and back
       Ctrl-E  <-  Clear debug view
       Tab     <-  Switch between various views

    [darkcyan]Search Result Rows[white]

	   j           <-  Move down
       k           <-  Move up
       i           <-  Load currently selected row in the edit view
       spacebar    <-  Select row
       d           <-  Batch delete selected rows
       Tab         <-  Go back to search bar

`)
	return "Help", h.Content
}
