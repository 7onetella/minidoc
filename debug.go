package minidoc

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os"
	"time"
)

type DebugView struct {
	App    *SimpleApp
	Layout *tview.Flex
	Rows   *tview.Table
	IsDebugOn bool
}

func NewDebugView(app *SimpleApp) *DebugView {

	rows := NewLoggingRows()

	layout := tview.NewFlex()
	layout.AddItem(rows, 0, 1, false)
	layout.SetBorder(true).SetBorderPadding(0, 0, 0, 0)
	layout.SetTitle("Debug")

	debugView := &DebugView{
		App:    app,
		Layout: layout,
		Rows:   rows,
	}
	return debugView
}

func NewLoggingRows() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		InsertColumn(0).
		InsertRow(0).
		InsertColumn(0).
		InsertColumn(0).
		SetSelectable(false, false).
		SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
	return table
}

func (d *DebugView) Debug(message string) {
	lastRow := 0
	if d.App == nil {
		os.Exit(1)
	}
	app := d.App
	rows := d.Rows
	rows.InsertRow(lastRow)
	rows.SetCell(lastRow, 0, &tview.TableCell{
		Text:            time.Now().Format(time.RFC3339),
		Align:           tview.AlignLeft,
		Color:           tcell.ColorDarkCyan,
		BackgroundColor: tcell.ColorDefault,
	})
	rows.SetCellSimple(lastRow, 1, message)
	rows.ScrollToBeginning()
	app.Draw()
}
