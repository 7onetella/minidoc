package minidoc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Search struct {
	App             *SimpleApp
	SearchBar       *tview.Form
	ResultList      *ResultList
	Detail          *tview.TextView
	Columns         *tview.Flex
	Layout          *tview.Flex
	IsEditMode      bool
	EditForm        *tview.Form
	CurrentRowIndex int
	debug           func(string)
}

func NewSearch() *Search {
	s := &Search{
		SearchBar: tview.NewForm(),
		Layout:    tview.NewFlex(),
	}
	return s
}

func (s *Search) SetApp(app *SimpleApp) {
	s.App = app
	s.debug = app.DebugView.Debug
}

// SearchPage returns search page
func (s *Search) Page() (title string, content tview.Primitive) {
	s.InitSearchBar()

	s.ResultList = NewResultList(s)

	s.Detail = tview.NewTextView()
	s.Detail.SetBorder(true)
	s.Detail.SetTitle("Preview")
	s.Detail.SetDynamicColors(true)

	s.Columns = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(s.ResultList, 0, 5, true).
		AddItem(s.Detail, 0, 5, false)

	rows := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(s.SearchBar, 2, 1, true).
		AddItem(s.Columns, 0, 10, false)

	s.Layout.AddItem(rows, 0, 1, true)
	s.Layout.SetBorder(true).SetBorderPadding(1, 1, 2, 2)

	return "Search", s.Layout
}

var words = []string{"@new", "@list", "@generate"}

func (s *Search) InitSearchBar() {
	//log.Debug("resetting search bar")
	s.SearchBar.AddInputField("", "", 0, nil, nil)
	s.SearchBar.SetBorderPadding(0, 1, 0, 0)
	item := s.SearchBar.GetFormItem(0)
	input, ok := item.(*tview.InputField)
	if ok {
		input.SetInputCapture(s.InputCapture(input))
	}

	input.SetAutocompleteFunc(func(currentText string) (entries []string) {
		if len(currentText) == 0 {
			return
		}
		for _, word := range words {
			if strings.HasPrefix(strings.ToLower(word), strings.ToLower(currentText)) {
				entries = append(entries, word)
			}
		}
		if len(entries) <= 1 {
			entries = nil
		}
		return
	})
}

func (s *Search) InputCapture(input *tview.InputField) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {

		text := input.GetText()
		var terms []string
		if len(text) > 0 {
			terms = strings.Split(text, " ")
		}
		if event.Key() == tcell.KeyEnter {
			// if term0 starts with @ and terms length is 1 then disregard enter
			if len(terms) == 0 && strings.HasPrefix(terms[0], "@") {
				return event
			}
			done := s.Search(text)
			if done {
				return nil
			}
			s.ResultList.ScrollToBeginning()
			s.SelectRow(0)
			s.App.SetFocus(s.SearchBar)
			defer s.App.Draw()
			return nil
		}

		if event.Key() == tcell.KeyTab {
			if len(terms) == 1 && strings.HasPrefix(terms[0], "@") {
				return event
			}
			log.Debug("tab pressed from search bar")
			s.GoToSearchResult()
			return nil
		}

		return event
	}
}

const typeColumnIndex = 0
const idColumnIndex = 1
const selectedColumnIndex = 2
const toggledColumnIndex = 3
const fragmentsColumnIndex = 4

func (s *Search) Search(searchby string) bool {
	searchTerms := ""
	if len(searchby) > 0 {
		searchTerms = searchby
	}
	log.Debug("searching by " + searchTerms)

	if strings.HasPrefix(searchTerms, "@") {
		s.HandleCommand(searchTerms)
		return true
	}

	for _, doctype := range doctypes {
		// e.g. url:10
		if strings.HasPrefix(searchTerms, doctype+":") {
			terms := strings.Split(searchTerms, " ")
			if len(terms) == 1 {
				idstr := searchTerms[strings.Index(searchTerms, ":")+1:]
				v, _ := strconv.Atoi(idstr)
				id := uint32(v)
				doc, err := s.App.BucketHandler.Read(id, doctype)
				// most likely record not found
				if err != nil {
					s.UpdateResult([]MiniDoc{})
					return false
				}
				doc.SetSearchFragments(doc.GetTitle())
				s.UpdateResult([]MiniDoc{doc})
				return false
			}
		}

		// e.g. url is typed
		if searchTerms == doctype {
			s.HandleCommand("@list " + searchTerms)
			return true
		}
	}

	result := s.App.IndexHandler.Search(searchTerms)
	if result == nil {
		s.debug("result is empty")
		return false
	}

	s.UpdateResult(result)
	return false
}

func (s *Search) UpdateResult(result []MiniDoc) {
	s.ResultList.Clear()
	// doc type
	s.ResultList.InsertColumns(5)

	// Display search result
	for _, doc := range result {
		s.ResultList.InsertRow(0)
		s.ResultList.UpdateRow(0, doc)
	}

	s.ResultList.SetSelectable(false, false)
	s.ResultList.SetSeparator(' ')
	s.Detail.Clear()

	s.App.SetFocus(s.ResultList)
	s.App.Draw()
}

func EditWithVim(app *SimpleApp, doc MiniDoc) MiniDoc {
	json := JsonMapFrom(doc)
	jh := NewJsonMapWrapper(json)

	for _, fieldName := range doc.GetViEditFields() {
		inputFile := "/tmp/.minidoc_input.tmp"
		WriteToFile(inputFile, jh.string(fieldName))

		OpenVim(app, inputFile)

		content, err := ReadFromFile(inputFile)
		log.Debugf("new content from input file: %s", content)
		if err != nil {
			log.Errorf("error reading: %v", err)
		}
		content = strings.TrimSpace(content)
		jh.set(fieldName, content)
		log.Debugf("json.description: %s", jh.string(fieldName))
	}
	doc, err := MiniDocFrom(json)
	if err != nil {
		log.Errorf("error converting: %v", err)
	}

	return doc
}

const (
	DIRECTION_NONE = iota
	DOWN
	UP
)

func (s *Search) UpdateCurrRowIndexFromSelectedRow(direction int) {
	rowIndex, _ := s.ResultList.GetSelection()
	log.Debugf("UpdateCurrRowIndexFromSelectedRow row index before: %d", rowIndex)
	switch direction {
	case DOWN:
		if rowIndex < (s.ResultList.GetRowCount() - 1) {
			rowIndex += 1
		}
	case UP:
		if rowIndex > 0 {
			rowIndex -= 1
		}
	}
	log.Debugf("UpdateCurrRowIndexFromSelectedRow row index: after %d", rowIndex)
	s.CurrentRowIndex = rowIndex
}

func (s *Search) SelectRow(row int) {
	if row == 0 || row < s.ResultList.GetRowCount() {
		s.CurrentRowIndex = row
		s.ResultList.Select(s.CurrentRowIndex, 0)
	}
}

func (s *Search) GoToSearchResult() {
	s.App.SetFocus(s.ResultList)
	s.ResultList.SetSelectable(true, false)
	s.Preview(DIRECTION_NONE)
	s.App.Draw()
}

func (s *Search) GoToSearchBar() {
	s.ResultList.SetSelectable(false, false)
	//s.SearchBar.Clear(true)
	//s.InitSearchBar()
	s.App.Draw()
	s.App.SetFocus(s.SearchBar)
}

func (s *Search) DelegateEventHandlingMiniDoc(event *tcell.EventKey) *tcell.EventKey {

	// in search result

	// get current row
	s.UpdateCurrRowIndexFromSelectedRow(DIRECTION_NONE)
	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return nil
	}

	// so far open browser for url
	doc.HandleEvent(event)

	// write any change from event handling
	s.App.DataHandler.Write(doc)

	// update the view
	s.Preview(DIRECTION_NONE)

	return nil
}

func (s *Search) Preview(direction int) {
	if s.IsEditMode {
		s.Columns.RemoveItem(s.EditForm)
		s.Columns.AddItem(s.Detail, 0, 5, true)
		s.IsEditMode = false
	}

	s.UpdateCurrRowIndexFromSelectedRow(direction)

	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return
	}

	// move keys like j and k controls the selection
	// result list select only makes sense for shifting the focus over and selecting
	s.App.StatusBar.SetText(fmt.Sprintf("%d | %s", s.CurrentRowIndex, doc.GetAvailableActions()))

	json := JsonMapFrom(doc)
	jh := NewJsonMapWrapper(json)

	content := ""
	s.Detail.SetTitle(doc.GetIDString())

	for _, fieldName := range doc.GetDisplayFields() {
		if fieldName == "type" || fieldName == "id" {
			continue
		}

		fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
		//s.debug("preview field for " + fieldNameCleaned)
		v := jh.string(fieldName)
		content += "\n"
		content += fmt.Sprintf("  [white]%s:[white]", fieldNameCleaned)
		lines := strings.Split(v, "\n")
		if len(lines) > 1 {
			content += "\n"
			for _, line := range lines {
				content += fmt.Sprintf("    [darkcyan]%s[darkcyan] \n", line)
			}
		} else {
			content += fmt.Sprintf(" [darkcyan]%s[darkcyan] \n", v)
		}
	}

	s.Detail.Clear()
	fmt.Fprintf(s.Detail, "%s", content)
}

func (s *Search) Edit() {
	s.ResultList.SetSelectable(false, false)

	doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
	if err != nil {
		log.Debugf("error getting json from curr row: %v", err)
		return
	}

	s.EditForm = NewEditForm(s, doc)

	if !s.IsEditMode {
		s.Columns.RemoveItem(s.Detail)
		s.Columns.AddItem(s.EditForm, 0, 5, true)
		s.IsEditMode = true
		s.App.SetFocus(s.EditForm)
	}
	s.App.Draw()
}

func (s *Search) BatchDeleteConfirmation() {
	ConfirmationModal(s.App, "Batch delete selected rows?", s.BatchDeleteActionFunc)
}

func ConfirmationModal(app *SimpleApp, message string, action func()) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				action()
			}
			if err := app.SetRoot(app.Layout, true).Run(); err != nil {
				panic(err)
			}
		})

	if err := app.SetRoot(modal, false).Run(); err != nil {
		panic(err)
	}
}

func (s *Search) BatchDeleteActionFunc() {
	for i := 0; i < s.ResultList.GetRowCount(); i++ {
		log.Debugf("current row %d", s.CurrentRowIndex)
		doc, err := s.LoadMiniDocFromDB(i)
		if err != nil {
			log.Errorf("minidoc from failed: %v", err)
			return
		}

		if !doc.IsSelected() {
			log.Debugf("row %d not selected skipping", i)
			continue
		}

		log.Debugf("deleting %v", doc)
		s.App.DataHandler.Delete(doc)
		if err != nil {
			log.Errorf("deleting %v failed: %v", doc, err)
			return
		}
		//log.Debugf("removing row %v", i)
		//s.ResultList.RemoveRow(i)
	}
}

func (s *Search) ToggleSelected() {
	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return
	}

	if doc.IsSelected() {
		doc.SetIsSelected(false)
	} else {
		doc.SetIsSelected(true)
	}

	s.ResultList.UpdateRow(s.CurrentRowIndex, doc)
}

func (s *Search) ToggleTogglable() {
	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return
	}

	if doc.GetToggle() {
		doc.SetToggle(false)
	} else {
		doc.SetToggle(true)
	}

	s.App.DataHandler.Write(doc)

	s.ResultList.UpdateRow(s.CurrentRowIndex, doc)
}

func (s *Search) LoadMiniDocFromDB(row int) (MiniDoc, error) {
	return s.ResultList.LoadMiniDocFromDB(row)
}

func (s *Search) UnLoadEdit() {
	s.GoToSearchResult()
	s.Preview(DIRECTION_NONE)
}
