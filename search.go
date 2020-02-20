package minidoc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Search struct {
	App             *SimpleApp
	SearchBar       *tview.Form
	ResultList      *tview.Table
	Detail          *tview.TextView
	Columns         *tview.Flex
	Layout          *tview.Flex
	IsResetted      bool
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
	s.ResetSearchBar()

	s.ResultList = tview.NewTable()
	s.ResultList.
		SetBorders(false).
		SetSeparator(' ').
		SetTitle("Results")
	s.ResultList.SetBorder(true)
	s.ResultList.SetBorderPadding(1, 1, 2, 2)

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

func (s *Search) ResetSearchBar() {
	log.Debug("resetting search bar")
	s.SearchBar.AddInputField("", "", 0, nil, nil)
	s.SearchBar.SetBorderPadding(0, 1, 0, 0)
	item := s.SearchBar.GetFormItem(0)
	input, ok := item.(*tview.InputField)

	if ok {
		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

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
				done := s.ShowSearchResult(text)
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
		})
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
}

const idColumnIndex = 1
const typeColumnIndex = 0
const fragmentsColumnIndex = 3
const selectedColumnIndex = 2

func (s *Search) ShowSearchResult(searchby string) bool {
	searchTerms := ""
	if len(searchby) > 0 {
		searchTerms = searchby
	}
	log.Debug("searching by " + searchTerms)

	if strings.HasPrefix(searchTerms, "@") {
		s.HandleCommand(searchTerms)
		return true
	}

	result := s.App.IndexHandler.Search(searchTerms)
	if result == nil {
		s.debug("result is empty")
		return false
	}

	s.UpdateResultList(result)
	return false
}

func (s *Search) UpdateResultList(result []MiniDoc) {
	s.ResultList.Clear()
	// doc type
	s.ResultList.InsertColumn(0)
	// id
	s.ResultList.InsertColumn(0)
	// matched
	s.ResultList.InsertColumn(0)
	// selected
	s.ResultList.InsertColumn(0)

	// Display search result
	for _, doc := range result {
		s.ResultList.InsertRow(0)
		s.UpdateSearchResultRow(0, doc)
	}

	s.ResultList.SetSelectable(false, false)
	s.ResultList.SetSeparator(' ')
	s.ResultList.SetSelectedStyle(tcell.ColorGray, tcell.ColorWhite, tcell.AttrNone)
	s.Detail.Clear()

	s.ResultList.SetInputCapture(s.GetResultListInputCaptureFunc())

	s.App.SetFocus(s.ResultList)
	s.App.Draw()
}

func (s *Search) UpdateSearchResultRow(rowIndex int, doc MiniDoc) {
	log.Debugf("updating row: id=%d row_index=%d minidoc=%v", doc.GetID(), rowIndex, doc)
	doctype := doc.GetType()
	doctype = strings.TrimSpace(doctype)
	i := 0
	s.ResultList.SetCell(rowIndex, i, NewCellWithBG(doctype, doc.GetIDString(), tcell.ColorWhite, tcell.ColorGray))
	i++
	s.ResultList.SetCell(rowIndex, i, NewCell(doc.GetID(), "", tcell.ColorWhite))
	i++
	s.ResultList.SetCell(rowIndex, i, NewCell(doc.IsSelected(), doc.IsSelectedString(), tcell.ColorWhite))
	i++
	// pad empty space to keep the result row width wider than few character wide
	log.Debugf("search fragments from doc %s", doc.GetSearchFragments())
	matched := doc.GetSearchFragments() + "                                                             "
	s.ResultList.SetCell(rowIndex, i, NewCell(matched, matched, tcell.ColorWhite))
}

func NewCellWithBG(reference interface{}, text string, color, bg tcell.Color) *tview.TableCell {
	return &tview.TableCell{
		Reference:       reference,
		Text:            text,
		Align:           tview.AlignLeft,
		Color:           color,
		BackgroundColor: bg,
	}
}

func NewCell(reference interface{}, text string, color tcell.Color) *tview.TableCell {
	return &tview.TableCell{
		Reference:       reference,
		Text:            text,
		Align:           tview.AlignLeft,
		Color:           color,
		BackgroundColor: tcell.ColorGray,
	}
}

func (s *Search) GetResultListInputCaptureFunc() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		//log.Debug("EventKey: " + event.Name())
		eventKey := event.Key()

		switch eventKey {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'i':
				s.LoadEdit()
			case 'j':
				// if the current mode is edit then remove edit and load detail
				s.LoadPreview(DOWN)
			case 'k':
				s.LoadPreview(UP)
			case 'e':
				key, done := s.EditCurrentFieldRowWithVi(event)
				if done {
					return key
				}
			case ' ':
				log.Debug("spacebar pressed")
				s.SelectRecordForCurrentRow()
			default:
				return s.DelegateAction(event)
			}

		case tcell.KeyTab:
			s.GoToSearchBar()
			return nil
		case tcell.KeyEnter:
			s.LoadPreview(DIRECTION_NONE)
			return nil
		case tcell.KeyCtrlD:
			s.DeleteSelectedRows()
		default:
			return s.DelegateAction(event)
		}

		return event
	}
}

func (s *Search) EditCurrentFieldRowWithVi(event *tcell.EventKey) (*tcell.EventKey, bool) {
	doc, err := s.GetMiniDocFromRow(s.CurrentRowIndex)
	if err != nil {
		log.Debugf("error getting json from curr row: %v", err)
		return event, true
	}
	json := Jsonize(doc)
	jh := NewJSONHandler(json)

	for _, fieldName := range doc.GetViEditFields() {
		inputFile := "/tmp/.minidoc_input.tmp"
		WriteToFile(inputFile, jh.string(fieldName))

		OpenVim(s.App, inputFile)

		content, err := ReadFromFile(inputFile)
		log.Debugf("new content from input file: %s", content)
		if err != nil {
			log.Errorf("error reading: %v", err)
		}
		content = strings.TrimSpace(content)
		jh.set(fieldName, content)
		log.Debugf("json.description: %s", jh.string(fieldName))
		doc, err = MiniDocFrom(json)
		if err != nil {
			log.Errorf("error converting: %v", err)
		}
		err = s.App.DataHandler.Write(doc)
		if err != nil {
			log.Errorf("error writing: %v", err)
		}
	}
	s.LoadPreview(DIRECTION_NONE)
	return nil, false
}

const (
	DIRECTION_NONE = iota
	DOWN
	UP
)

func (s *Search) SetCurrentRowIndex(direction int) {
	rowIndex, _ := s.ResultList.GetSelection()
	log.Debugf("SetCurrentRowIndex row index before: %d", rowIndex)
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
	log.Debugf("SetCurrentRowIndex row index: after %d", rowIndex)
	s.CurrentRowIndex = rowIndex
}

func (s *Search) SelectRow(rowIndex int) {
	if rowIndex == 0 || rowIndex < s.ResultList.GetRowCount() {
		s.CurrentRowIndex = rowIndex
		s.ResultList.Select(s.CurrentRowIndex, 0)
	}
}

func (s *Search) GoToSearchResult() {
	s.App.SetFocus(s.ResultList)
	s.ResultList.SetSelectable(true, false)
	s.LoadPreview(DIRECTION_NONE)
	s.App.Draw()
}

func (s *Search) GoToSearchBar() {
	s.ResultList.SetSelectable(false, false)
	s.SearchBar.Clear(true)
	s.ResetSearchBar()
	s.IsResetted = true
	s.App.Draw()
	s.App.SetFocus(s.SearchBar)
}

func (s *Search) DelegateAction(event *tcell.EventKey) *tcell.EventKey {

	// in search result

	// get current row
	s.SetCurrentRowIndex(DIRECTION_NONE)
	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.GetMiniDocFromRow(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return nil
	}

	// call doc.HandleEvent(key)
	doc.HandleEvent(event)
	s.App.DataHandler.Write(doc)
	s.LoadPreview(DIRECTION_NONE)
	// e.g.
	// call receiver if key o = open for url, if key d = mark done for todo task
	// return nil

	return nil
}

func (s *Search) LoadPreview(direction int) {
	if s.IsEditMode {
		s.Columns.RemoveItem(s.EditForm)
		s.Columns.AddItem(s.Detail, 0, 5, true)
		s.IsEditMode = false
	}

	s.SetCurrentRowIndex(direction)

	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.GetMiniDocFromRow(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return
	}

	// move keys like j and k controls the selection
	// result list select only makes sense for shifting the focus over and selecting
	s.App.StatusBar.SetText(fmt.Sprintf("%d | %s", s.CurrentRowIndex, doc.GetAvailableActions()))

	json := Jsonize(doc)
	jh := NewJSONHandler(json)

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

func (s *Search) LoadEdit() {
	s.ResultList.SetSelectable(false, false)

	doc, err := s.GetMiniDocFromRow(s.CurrentRowIndex)
	if err != nil {
		log.Debugf("error getting json from curr row: %v", err)
		return
	}

	s.EditForm = NewEdit(s, doc).Form

	if !s.IsEditMode {
		s.Columns.RemoveItem(s.Detail)
		s.Columns.AddItem(s.EditForm, 0, 5, true)
		s.IsEditMode = true
		s.App.SetFocus(s.EditForm)
	}
	s.App.Draw()
}

func (s *Search) DeleteSelectedRows() {
	ConfirmationModal(s.App, "Batch delete selected rows?", s.BatchDelete)
}

func (s *Search) BatchDelete() {
	for i := 0; i < s.ResultList.GetRowCount(); i++ {
		log.Debugf("current row %d", s.CurrentRowIndex)
		doc, err := s.GetMiniDocFromRow(i)
		if err != nil {
			log.Errorf("minidoc from failed: %v", err)
			return
		}

		if !doc.IsSelected() {
			log.Debugf("row %d not selected skipping", i)
			continue
		}

		log.Debugf("deleting %v", doc)
		s.DeleteFunc(doc)
		if err != nil {
			log.Errorf("deleting %v failed: %v", doc, err)
			return
		}
		//log.Debugf("removing row %v", i)
		//s.ResultList.RemoveRow(i)
	}
}

func (s *Search) SelectRecordForCurrentRow() {
	log.Debugf("current row %d", s.CurrentRowIndex)
	doc, err := s.GetMiniDocFromRow(s.CurrentRowIndex)
	if err != nil {
		log.Errorf("minidoc from failed: %v", err)
		return
	}

	if doc.IsSelected() {
		doc.SetIsSelected(false)
	} else {
		doc.SetIsSelected(true)
	}

	s.UpdateSearchResultRow(s.CurrentRowIndex, doc)
}

func (s *Search) DeleteFunc(doc MiniDoc) error {
	err := s.App.BucketHandler.Delete(doc)
	if err != nil {
		return err
	}
	err = s.App.IndexHandler.Delete(doc)
	return err
}

func (s *Search) GetMiniDocFromRow(rowIndex int) (MiniDoc, error) {
	ref := s.ResultList.GetCell(rowIndex, idColumnIndex).GetReference()
	id, ok := ref.(uint32)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] id column not uint32 but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	ref = s.ResultList.GetCell(rowIndex, typeColumnIndex).GetReference()
	doctype, ok := ref.(string)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] type column not string but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	ref = s.ResultList.GetCell(rowIndex, fragmentsColumnIndex).GetReference()
	fragments, ok := ref.(string)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] type column not string but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}
	log.Debugf("read fragments from current row %s", fragments)
	ref = s.ResultList.GetCell(rowIndex, selectedColumnIndex).GetReference()
	isSelected, ok := ref.(bool)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] type column not bool but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	json, err := s.App.BucketHandler.Read(id, doctype)
	if err != nil {
		log.Debugf("read error: %v", err)
		return nil, err
	}

	// json unmarshaller will exclude empty value fields
	// jh.set("fragments", fragments) will throw error
	// convert to minidoc to set these two values
	doc, _ := MiniDocFrom(json)
	doc.SetIsSelected(isSelected)
	doc.SetSearchFragments(fragments)
	return doc, nil
}

//type RefHandler struct{
//	err error
//}
//
//func NewRefHandler() *RefHandler {
//	return &RefHandler{}
//}
//
//func (rh *RefHandler) uint32(ref interface{}) uint32 {
//	v, ok := ref.(uint32)
//	if !ok {
//		rh.err = fmt.Errorf(fmt.Sprintf("uint32 is expected but got %v", reflect.TypeOf(ref)))
//		log.Errorf()
//	}
//	return v
//}

func (s *Search) UnLoadEdit() {
	s.GoToSearchResult()
	s.LoadPreview(DIRECTION_NONE)
}
