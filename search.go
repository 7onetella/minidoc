package minidoc

import (
	"fmt"
	"github.com/0xAX/notificator"
	"github.com/atotto/clipboard"
	"strconv"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/google/uuid"
	"github.com/rivo/tview"
)

type Search struct {
	App             *SimpleApp
	SearchBar       *tview.Form
	Rows            *tview.Flex
	ResultList      *ResultList
	Detail          *tview.TextView
	Columns         *tview.Flex
	Layout          *tview.Flex
	IsEditMode      bool
	EditForm        *tview.Form
	CurrentRowIndex int
	debug           func(string)
	RegionID        int
	RegionCount     int
	RegionDocIDs    map[int]string
	Referenced      *tview.TextView
}

func NewSearch() *Search {
	s := &Search{
		SearchBar:    tview.NewForm(),
		Layout:       tview.NewFlex(),
		RegionDocIDs: map[int]string{},
	}
	return s
}

func (s *Search) SetApp(app *SimpleApp) {
	s.App = app
	s.debug = app.DebugView.Debug
}

// SearchPage returns search page
func (s *Search) Page() (title string, content tview.Primitive) {
	s.InitSearchBar("search")

	s.ResultList = NewResultList(s)

	s.Detail = tview.NewTextView()
	s.Detail.SetBorder(true)
	s.Detail.SetTitle("Preview")
	s.Detail.SetDynamicColors(true)
	s.Detail.SetRegions(true)
	s.Detail.SetBorderPadding(0, 1, 2, 2)
	s.Detail.SetInputCapture(s.PreviewInputCapture())

	s.Referenced = tview.NewTextView()
	s.Referenced.SetBorder(true)
	s.Referenced.SetTitle("Referenced")
	s.Referenced.SetDynamicColors(true)
	s.Referenced.SetRegions(true)
	s.Referenced.SetBorderPadding(0, 1, 2, 2)

	s.Columns = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(s.ResultList, 0, 5, true).
		AddItem(s.Detail, 0, 5, false)

	s.Rows = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(s.SearchBar, 2, 1, true).
		AddItem(s.Columns, 0, 10, false)

	s.Layout.AddItem(s.Rows, 0, 1, true)
	s.Layout.SetBorder(true).SetBorderPadding(1, 1, 2, 2)

	return "Search", s.Layout
}

func (s *Search) GetInstance() interface{} {
	return s
}

var words = []string{"@new", "@generate", "@tag", "@untag", "@export", "@import"}

func (s *Search) InitSearchBar(placeholder string) {
	//log.Debug("resetting search bar")

	input := tview.NewInputField().SetFieldWidth(0).SetPlaceholder(placeholder)
	s.SearchBar.AddFormItem(input)
	s.SearchBar.SetBorderPadding(0, 1, 0, 0)
	s.SearchBar.SetFieldTextColor(tcell.ColorYellow)
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
			if terms == nil {
				return event
			}

			// if term0 starts with @ and terms length is 1 then disregard enter
			if len(terms) == 1 && strings.HasPrefix(terms[0], "@") {
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
			log.Debug("tab pressed from search bar")
			if len(terms) == 1 && strings.HasPrefix(terms[0], "@") {
				return event
			}

			if s.ResultList.GetRowCount() > 0 {
				s.GoToSearchResult()
				return nil
			}
		}

		if event.Key() == tcell.KeyCtrlSpace {
			s.GoToSearchBar(true, "")
			return nil
		}

		return event
	}
}

func (s *Search) PreviewInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		eventKey := event.Key()

		switch eventKey {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'n':
				if s.RegionCount > 0 {
					s.RegionID++
					if s.RegionID%s.RegionCount == 0 {
						s.RegionID = 0
					}
					regionID := fmt.Sprintf("%d", s.RegionID)
					s.Detail.Highlight(regionID)
					s.Referenced.Clear()

					s.ShowNextReferenced()
				}
				//s.App.StatusBar.SetText(fmt.Sprintf("region count: %d, region id: %d", s.RegionCount, s.RegionID))
			case 'e':
				doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
				if err != nil {
					log.Debugf("error getting json from curr row: %v", err)
					return event
				}
				doc, changed := EditWithVim(s.App, doc)
				if changed {
					_, err = s.App.DataHandler.Write(doc)
					if err != nil {
						log.Errorf("error writing: %v", err)
					}
				}
				s.Preview(DIRECTION_NONE)
				return nil
			case 'i':
				s.Edit()
			default:
				if s.RegionCount > 0 {
					//regionText := s.Detail.GetRegionText(fmt.Sprintf("%d", s.RegionID))

					docid := s.RegionDocIDs[s.RegionID]
					sslice := strings.Split(docid, ":")
					doctype := sslice[0]
					id := sslice[1]

					doc, _ := s.App.DataHandler.BucketHandler.Read(toUnit32FromString(id), doctype)
					if doc != nil {
						doc.HandleEvent(event)
						//s.App.StatusBar.SetText(doc.GetTitle())
					}
				}
				return nil
			}

		case tcell.KeyTab:
			s.GoToSearchBar(false, "")
			s.HideReferenced()
			return nil
		case tcell.KeyBacktab:
			s.GoToSearchResult()
			s.HideReferenced()
			return nil
		default:
			return event
		}

		return event
	}
}

func (s *Search) ShowNextReferenced() {
	docid := s.RegionDocIDs[s.RegionID]
	sslice := strings.Split(docid, ":")
	if len(sslice) != 2 {
		return
	}
	doctype := sslice[0]
	id := sslice[1]

	doc, _ := s.App.DataHandler.BucketHandler.Read(toUnit32FromString(id), doctype)
	if doc != nil {
		json := JsonMapFrom(doc)
		jh := NewJsonMapWrapper(json)

		content := ""
		s.Referenced.SetTitle(doc.GetIDString())
		for _, fieldName := range doc.GetDisplayFields() {
			if fieldName == "type" || fieldName == "id" {
				continue
			}

			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			v := jh.string(fieldName)
			content += "\n"
			content += fmt.Sprintf("[white]%s:[white] ", fieldNameCleaned)
			content += "[darkcyan]"
			content += Transpose(v, s)
			content += "[darkcyan]"
			content += "\n"
		}
		fmt.Fprintf(s.Referenced, content)
		s.Rows.RemoveItem(s.Referenced)
		s.Rows.AddItem(s.Referenced, 0, 3, false)
	}
}

func (s *Search) HideReferenced() {
	s.Rows.RemoveItem(s.Referenced)
}

const typeColumnIndex = 0
const selectedColumnIndex = 1
const toggledColumnIndex = 2
const fragmentsColumnIndex = 3
const idColumnIndex = 4

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

	result, stat := s.App.IndexHandler.Search(searchTerms)
	if result == nil {
		s.debug("result is empty")
		return false
	}
	s.App.SetStatus("[white:darkcyan] " + stat + "[white]")

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

func EditWithVim(app *SimpleApp, doc MiniDoc) (MiniDoc, bool) {
	json := JsonMapFrom(doc)
	jh := NewJsonMapWrapper(json)

	changed := false
	for _, fieldName := range doc.GetViEditFields() {
		UUID := uuid.New().String()
		file := fmt.Sprintf("/tmp/%s", UUID)
		// write field value to the file
		WriteToFile(file, jh.string(fieldName))
		// let user edit
		OpenVim(app, file)
		// read what was entered
		content, err := ReadFromFile(file)
		// delete now since content has been read
		DeleteFile(file)

		content = strings.TrimSpace(content)
		if len(content) > 0 {
			changed = true
		}
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

	return doc, changed
}

const (
	DIRECTION_NONE = iota
	DOWN
	UP
)

func (s *Search) UpdateCurrRowIndexFromSelectedRow(direction int) {
	rowIndex, _ := s.ResultList.GetSelection()
	//log.Debugf("UpdateCurrRowIndexFromSelectedRow row index before: %d", rowIndex)
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
	//log.Debugf("UpdateCurrRowIndexFromSelectedRow row index: after %d", rowIndex)
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

func (s *Search) GoToSearchBar(clear bool, placeholder string) {
	s.ResultList.SetSelectable(false, false)
	if clear {
		s.SearchBar.Clear(true)
		s.InitSearchBar(placeholder)
	}
	s.App.SetStatus("[white:darkcyan] Ctrl-h <- navigate left | Ctrl-l <- navigate right[white]")
	s.App.Draw()
	s.App.SetFocus(s.SearchBar)
}

func (s *Search) GoToPreview() {
	s.App.SetFocus(s.Detail)
	s.Detail.Highlight("0")
	s.ShowNextReferenced()
	s.App.Draw()
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
	s.App.SetStatus(fmt.Sprintf("[white:darkcyan] spacebar <- select | %s", doc.GetAvailableActions()), fmt.Sprintf("row %d ", s.CurrentRowIndex))

	json := JsonMapFrom(doc)
	jh := NewJsonMapWrapper(json)

	content := ""
	s.Detail.SetTitle(doc.GetIDString())
	s.RegionID = 0
	s.RegionCount = 0
	for _, fieldName := range doc.GetDisplayFields() {
		if fieldName == "type" || fieldName == "id" {
			continue
		}

		fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
		//s.debug("preview field for " + fieldNameCleaned)
		v := jh.string(fieldName)
		content += "\n"
		content += fmt.Sprintf("[white]%s:[white] ", fieldNameCleaned)

		lines := strings.Split(v, "\n")
		if len(lines) > 1 {
			for _, line := range lines {
				content += "\n"
				content += "[darkcyan]"
				content += Transpose(line, s)
				content += "[darkcyan]"
			}
		} else {
			content += "[darkcyan]"
			content += Transpose(v, s)
			content += "[darkcyan]"
		}
		content += "\n"
	}

	s.Detail.Clear()
	fmt.Fprintf(s.Detail, content)
}

func Transpose(line string, s *Search) string {
	tokens := strings.Split(line, " ")
	n := len(tokens)
	content := make([]string, n)
	for i, token := range tokens {
		if strings.HasPrefix(token, "[") &&
			strings.HasSuffix(token, "]") {
			docid := token[1 : len(token)-1]
			sslice := strings.Split(docid, ":")
			doctype := sslice[0]
			id := sslice[1]
			doc, _ := s.App.DataHandler.BucketHandler.Read(toUnit32FromString(id), doctype)
			if doc != nil {
				content[i] = fmt.Sprintf(`["%d"][yellow]%s[darkcyan][""]`, s.RegionCount, doc.GetTitle())
				s.RegionDocIDs[s.RegionCount] = docid
			}
			s.RegionCount++
		} else {
			content[i] = fmt.Sprintf("%s", token)
		}
	}
	out := ""
	for i := 0; i < n; i++ {
		out += content[i] + " "
	}
	return out
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

func (s *Search) SelectAllRows() {
	for i := 0; i < s.ResultList.GetRowCount(); i++ {
		log.Debugf("current row %d", s.CurrentRowIndex)
		doc, err := s.LoadMiniDocFromDB(i)
		if err != nil {
			log.Errorf("minidoc from failed: %v", err)
			return
		}

		if !doc.IsSelected() {
			doc.SetIsSelected(true)
		} else {
			doc.SetIsSelected(false)
		}
		s.ResultList.UpdateRow(i, doc)
		s.App.ForceDraw()
	}
}

func (s *Search) ToggleAllRows() {
	for i := 0; i < s.ResultList.GetRowCount(); i++ {
		log.Debugf("current row %d", s.CurrentRowIndex)
		doc, err := s.LoadMiniDocFromDB(i)
		if err != nil {
			log.Errorf("minidoc from failed: %v", err)
			return
		}

		if !doc.GetToggle() {
			doc.SetToggle(true)
		} else {
			doc.SetToggle(false)
		}
		_, err = s.App.DataHandler.Write(doc)
		if err == nil {
			s.ResultList.UpdateRow(i, doc)
			s.App.ForceDraw()
		}
	}
}

func (s *Search) ClipboardCopy() {
	content := ""
	for i := 0; i < s.ResultList.GetRowCount(); i++ {
		log.Debugf("current row %d", s.CurrentRowIndex)
		doc, err := s.LoadMiniDocFromDB(i)
		if err != nil {
			log.Errorf("minidoc from failed: %v", err)
			return
		}
		content += "[" + doc.GetIDString() + "]\n"
		content += "\n"
	}
	clipboard.WriteAll(content)
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

func (s *Search) Notify(title, text string) {
	notify := notificator.New(notificator.Options{
		DefaultIcon: "icon/default.png",
		AppName:     "Minidoc",
	})

	notify.Push(title, text, "/home/user/icon.png", notificator.UR_CRITICAL)
}
