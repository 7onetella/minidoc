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
    [white:darkcyan][List of shortcut keys for the following scope[][white]

    [black:darkcyan][Entire App[][white]

       Ctrl-c  <-  Exit
       Ctrl-n  <-  New note
       Ctrl-h  <-  Navigate to left menu item
       Ctrl-l  <-  Navigate to right menu item
       Ctrl-o  <-  Show debug view
       Ctrl-d  <-  Go to debug view and back
       Ctrl-e  <-  Clear debug view
       Tab     <-  Switch between various views

    [black:darkcyan][Search Bar[][white]

       Ctrl-n      <-  New note
       Ctrl-Space  <-  Delete the entire line.
       Ctrl-u      <-  Delete the entire line.
	   Ctrl-a      <-  Move to the beginning of the line.
	   Ctrl-e      <-  Move to the end of the line.
	   Alt-left, Alt-b   <-  Move left by one word.
	   Alt-right, Alt-f  <-  Move right by one word.
	   Ctrl-k      <-  Delete from the cursor to the end of the line.
	   Ctrl-w      <-  Delete the last word before the cursor.

    [black:darkcyan][Search Result Rows[][white]

       Tab         <-  Go back to search bar
	   j           <-  Move down
       k           <-  Move up
       i           <-  Load currently selected row in the edit view
       e           <-  Edit vim editable fields, e.g. note
	   t           <-  Toggle toggle-able field
       spacebar    <-  Select row
       Ctrl-j      <-  Move row down
       Ctrl-k      <-  Move row up
       Ctrl-d      <-  Batch delete selected rows
       Ctrl-a      <-  Select all / Deselect all
       Ctrl-t      <-  Toggle all / Detoggle all
`)
	return "Help", h.Content
}
