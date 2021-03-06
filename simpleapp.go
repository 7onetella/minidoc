package minidoc

import (
	"fmt"
	"github.com/7onetella/minidoc/config"
	l "github.com/7onetella/minidoc/log"
	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

func init() {
	minidocHome := CreateMinidocHomeIfNotFound()
	CreateGeneratedDirIfNotFound()
	log = l.GetNewLogrusLogger(minidocHome)
}

var log *logrus.Logger

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

	app.Draw()
	app.SetStatus("[white:darkcyan] Ctrl-h <- navigate left | Ctrl-l <- navigate right[white]" +
		"                                                                                                     ")

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
			NewDocFlow("note", app)
			defer app.Draw()
			return nil
		case tcell.KeyCtrlU:
			NewDocFlow("url", app)
			defer app.Draw()
			return nil
		case tcell.KeyCtrlT:
			NewDocFlow("todo", app)
			defer app.Draw()
			return nil
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

func (app *SimpleApp) SetStatus(text string, rightsides ...string) {
	_, _, width, _ := app.Layout.GetRect()

	rightsidesLength := 0
	if len(rightsides) > 0 {
		rightsidesLength = len(strings.Join(rightsides, " "))
	}
	app.StatusBar.SetText(text)
	rawText := app.StatusBar.GetText(true)
	rawText = strings.ReplaceAll(rawText, "\n", "")

	textLength := len(rawText)
	paddingWidth := width - textLength - rightsidesLength
	padding := fmt.Sprintf("%-*s", paddingWidth, "")
	fmt.Fprint(app.StatusBar, padding)

	for _, rightside := range rightsides {
		fmt.Fprint(app.StatusBar, rightside)
	}
}

func newTextViewBar() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
}

func CreateMinidocHomeIfNotFound() string {
	homedir, _ := homedir.Dir() // return path with slash at the end
	minidocHome := homedir + "/.minidoc"

	if contains(os.Args, "--dev") {
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err.Error())
		}

		minidocHome = pwd + "/.minidoc"
	}

	//fmt.Println("creating " + minidocHome)
	os.MkdirAll(minidocHome, os.ModePerm)
	return minidocHome
}

func CreateGeneratedDirIfNotFound() string {
	cfg := config.Config()
	generatedDir := cfg.GetString("generated_doc_path")
	homedir, _ := homedir.Dir() // return path with slash at the end
	minidocGenDir := homedir + generatedDir

	os.MkdirAll(minidocGenDir, os.ModePerm)
	return minidocGenDir
}
