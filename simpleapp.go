package minidoc

import (
	l "github.com/7onetella/minidoc/log"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"os"
	"strconv"
)

var log = l.GetNewLogrusLogger()

// SimpleApp comes with the menu and debug pane
type SimpleApp struct {
	*tview.Application
	MenuBar           *tview.TextView
	StatusBar         *tview.TextView
	Layout            *tview.Flex
	Rows              *tview.Flex
	PagesHandler      *PagesHandler
	DebugView         *DebugView
	PrevFocused       tview.Primitive
	delegatedKeyEvent func(event *tcell.EventKey) *tcell.EventKey
	confirmExit       bool
	IndexHandler      *IndexHandler
	BucketHandler     *BucketHandler
	dataFolderPath    string
	DataHandler       *DataHandler
	docsReindexed     bool
}

type SimpleAppOption func(*SimpleApp)

func WithSimpleAppConfirmExit(confirm bool) SimpleAppOption {
	return func(app *SimpleApp) {
		app.confirmExit = confirm
	}
}

func WithSimpleAppPages(pages []PageItem) SimpleAppOption {
	return func(app *SimpleApp) {
		app.PagesHandler.PageItems = pages
	}
}

func WithSimpleAppDebugOn() SimpleAppOption {
	return func(app *SimpleApp) {
		app.ToggleDebug()
	}
}

func WithSimpleAppDelegateKeyEvent(keyEvent func(event *tcell.EventKey) *tcell.EventKey) SimpleAppOption {
	return func(app *SimpleApp) {
		app.delegatedKeyEvent = keyEvent
	}
}

func WithSimpleAppDataFolderPath(path string) SimpleAppOption {
	return func(app *SimpleApp) {
		app.dataFolderPath = path
	}
}

func WithSimpleAppDocsReindexed(reindex bool) SimpleAppOption {
	return func(app *SimpleApp) {
		app.docsReindexed = reindex
	}
}

func NewSimpleApp(opts ...SimpleAppOption) *SimpleApp {

	menu := newTextViewBar()
	status := newTextViewBar()
	layout := tview.NewFlex()

	pageHandler := &PagesHandler{
		Pages:         tview.NewPages(),
		PageIndex:     map[string]string{},
		MenuBar:       menu,
		CurrPageIndex: 0,
		PageItems:     nil,
	}

	app := &SimpleApp{
		tview.NewApplication(),
		menu,
		status,
		layout,
		nil,
		pageHandler,
		nil,
		nil,
		nil,
		false,
		nil,
		nil,
		".",
		nil,
		false,
	}

	status.SetText("Ctrl-L for help")
	app.DebugView = NewDebugView(app)

	app.Rows = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(menu, 1, 1, false).
		AddItem(pageHandler.Pages, 0, 9, true).
		AddItem(status, 1, 1, false)

	for _, opt := range opts {
		opt(app)
	}

	app.BucketHandler = NewBucketHandler(
		WithBucketHandlerDebug(app.DebugView.Debug),
		WithBucketHandlerDBPath(app.dataFolderPath+"/store.db"),
	)
	app.IndexHandler = NewIndexHandler(
		WithIndexHandlerDebug(app.DebugView.Debug),
		WithIndexHandlerIndexPath(app.dataFolderPath+"/index"),
	)
	app.DataHandler = &DataHandler{
		app.BucketHandler,
		app.IndexHandler,
	}

	if app.docsReindexed {
		Reindex(app.IndexHandler, app.BucketHandler)
	}

	app.MenuBar.Highlight(strconv.Itoa(0))

	pageHandler.LoadPages(app)

	layout.AddItem(app.Rows, 0, 1, true)
	app.Layout = layout

	app.SetInputCapture(app.GetInputCaptureFunc())

	return app
}

func Reindex(ih *IndexHandler, db *BucketHandler) {
	buckets := doctypes
	for _, bucket := range buckets {
		docs, _ := db.ReadAll(bucket)
		ih.debug(bucket + ":retrieved:" + strconv.Itoa(len(docs)))
		for _, doc := range docs {
			err := ih.Index(doc)
			if err != nil {
				log.Errorf("indexing %v failed: %v", doc, err)
				return
			}
		}
	}
}

func (app *SimpleApp) GetInputCaptureFunc() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		pagesHandler := app.PagesHandler

		switch event.Key() {
		case tcell.KeyCtrlE:
			app.ClearDebug()
		case tcell.KeyCtrlL:
			pagesHandler.GoToNextPage()
			return nil
		case tcell.KeyCtrlH:
			pagesHandler.GoToPrevPage()
			return nil
		case tcell.KeyCtrlO:
			app.ToggleDebug()
		case tcell.KeyCtrlD:
			app.GoToDebugView()
		case tcell.KeyCtrlN:
			if err := NewDocFlow("note", app); err != nil {
				return nil
			}
			defer app.Draw()
		case tcell.KeyCtrlC:
			app.Exit()
		default:
			//if app.delegatedKeyEvent != nil {
			//	delegated := app.delegatedKeyEvent(event)
			//	return delegated
			//}
		}

		return event
	}
}

func (app *SimpleApp) Exit() {
	modal := tview.NewModal().
		SetText("Do you want to quit the application?").
		AddButtons([]string{"Quit", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Quit" {
				app.Stop()
				l.Logfile.Sync()
				l.Logfile.Close()
				os.Exit(0)
			} else {
				if err := app.SetRoot(app.Layout, true).Run(); err != nil {
					panic(err)
				}
			}
		})

	if app.confirmExit {
		if err := app.SetRoot(modal, false).Run(); err != nil {
			panic(err)
		}
	}
	app.Stop()
	l.Logfile.Sync()
	l.Logfile.Close()
	os.Exit(0)
}

func (app *SimpleApp) GoToDebugView() {
	loggingRows := app.DebugView.Rows
	if !loggingRows.HasFocus() {
		app.PrevFocused = app.GetFocus()
		app.SetFocus(loggingRows)
		loggingRows.SetSelectable(true, false)
	} else {
		loggingRows.SetSelectable(false, false)
		app.SetFocus(app.PrevFocused)
	}
	app.Draw()
}

func (app *SimpleApp) ToggleDebug() {
	debugView := app.DebugView
	if debugView.IsDebugOn {
		debugView.IsDebugOn = false
		app.Rows.RemoveItem(app.DebugView.Layout)
	} else {
		debugView.IsDebugOn = true
		app.Rows.AddItem(app.DebugView.Layout, 0, 5, false)
	}
	app.Draw()
}

func (app *SimpleApp) ClearDebug() {
	app.DebugView.Rows.Clear()
	app.Draw()
}

func (app *SimpleApp) ClearMenu() {
	app.MenuBar.Clear()
	app.Draw()
}

func newTextViewBar() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
}
