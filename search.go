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
		AddItem(s.Columns, 0, 10, true)

	s.Layout.AddItem(rows, 0, 1, true)
	s.Layout.SetBorder(true).SetBorderPadding(1, 1, 2, 2)

	return "Search", s.Layout
}

func (s *Search) ResetSearchBar() {
	log.Debug("resetting search bar")
	s.SearchBar.AddInputField("", "", 0, nil, nil)
	s.SearchBar.SetBorderPadding(0, 1, 0, 0)
	item := s.SearchBar.GetFormItem(0)
	input, ok := item.(*tview.InputField)

	if ok {
		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

			if event.Key() == tcell.KeyCtrlN {
				s.LoadEdit()
				s.App.SetFocus(s.SearchBar)
				defer s.App.Draw()
				return nil
			}

			if event.Key() == tcell.KeyEnter {
				s.ShowSearchResult(input.GetText())
				s.App.SetFocus(s.SearchBar)
				defer s.App.Draw()
				return nil
			}

			if event.Key() == tcell.KeyTab {
				log.Debug("tab pressed from search bar")
				s.GoToSearchResult()
				return nil
			}

			return event
		})
	}
}

const idColumnIndex = 1
const typeColumnIndex = 0

func (s *Search) ShowSearchResult(searchby string) {
	searchTerms := ""
	if len(searchby) > 0 {
		searchTerms = searchby
	}
	log.Debug("searching by " + searchTerms)

	result := s.App.IndexHandler.Search(searchTerms)
	if result == nil {
		s.debug("result is empty")
		return
	}

	//debug("result size is " + strconv.Itoa(len(result.Items)))
	s.ResultList.Clear()
	s.ResultList.InsertColumn(0)
	s.ResultList.InsertColumn(0)
	s.ResultList.InsertColumn(0)
	//s.ResultList.InsertColumn(0)

	// Display search result
	for _, doc := range result {
		s.ResultList.InsertRow(0)
		s.UpdateSearchResultRow(0, doc)
	}

	//headerCell := func(val string) *tview.TableCell {
	//	return &tview.TableCell{
	//		Color:         tcell.ColorYellow,
	//		Align:         tview.AlignCenter,
	//		Text:          val,
	//		NotSelectable: true,
	//	}
	//}

	//s.ResultList.InsertRow(0)
	//i := 0
	//s.ResultList.SetCell(0, i, headerCell(""))
	//i++
	//s.ResultList.SetCell(0, i, headerCell(""))
	//i++
	//s.ResultList.SetCell(0, i, headerCell("Title"))
	//i++
	//s.ResultList.SetCell(0, i, headerCell(""))

	s.ResultList.SetSelectable(false, false)
	s.ResultList.SetSeparator(' ')
	s.ResultList.SetSelectedStyle(tcell.ColorGray, tcell.ColorWhite, tcell.AttrNone)
	s.Detail.Clear()

	s.ResultList.SetInputCapture(s.GetInputCaptureFunc())

	s.App.SetFocus(s.ResultList)
	s.App.Draw()
}

func (s *Search) UpdateSearchResultRow(rowIndex int, doc MiniDoc) {
	i := 0
	log.Debugf("updating row: id=%d row_index=%d minidoc=%v", doc.GetID(), rowIndex, doc)

	doctype := doc.GetType()
	doctype = strings.TrimSpace(doctype)
	s.ResultList.SetCell(rowIndex, i, NewCellWithBG(doctype, doc.GetIDString(), tcell.ColorWhite, tcell.ColorGray))
	i++
	s.ResultList.SetCell(rowIndex, i, NewCell(doc.GetID(), "", tcell.ColorWhite))
	i++
	//s.ResultList.SetCell(rowIndex, i, NewCell(doc.GetTitle(), doc.GetTitle(), tcell.ColorDarkCyan))
	//i++
	matched := ""
	// if the result is only matched on doc type, don't show type shown in fragment
	// don't make sense showing type and type as fragments
	if doc.GetType() != doc.GetSearchFragments() {
		matched = doc.GetSearchFragments() + "                                                         "
	}
	s.ResultList.SetCell(rowIndex, i, NewCell(matched, matched, tcell.ColorWhite))
	//s.ResultList.SetCellSimple(rowIndex, i, matched)
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

func (s *Search) GetInputCaptureFunc() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		//debug("EventKey: " + event.Name())

		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'i':
				s.LoadEdit()
			case 'j':
				// if the current mode is edit then remove edit and load detail
				s.LoadPreview(DOWN)
			case 'k':
				s.LoadPreview(UP)
			}

		case tcell.KeyTab:
			s.GoToSearchBar()
			return nil

		case tcell.KeyEnter:
			s.LoadPreview(DIRECTION_NONE)
			return nil
		}

		return event
	}
}

const (
	DIRECTION_NONE = iota
	DOWN
	UP
)

func (s *Search) SetNextRowIndex(direction int) {
	rowIndex, _ := s.ResultList.GetSelection()
	log.Debugf("s.ResultList.GetSelection() before: %d", rowIndex)
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
	log.Debugf("s.ResultList.GetSelection(): after %d", rowIndex)
	s.CurrentRowIndex = rowIndex
}

func (s *Search) GoToSearchResult() {
	s.App.SetFocus(s.ResultList)
	s.ResultList.SetSelectable(true, false)
	//if s.ResultList.GetRowCount() > 0 && !s.IsResetted {
	//	// down one because of header row
	//	s.LoadPreview(DOWN)
	//} else {
	s.LoadPreview(DIRECTION_NONE)
	//}
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

func (s *Search) LoadPreview(direction int) {
	if s.IsEditMode {
		s.Columns.RemoveItem(s.EditForm)
		s.Columns.AddItem(s.Detail, 0, 5, true)
		s.IsEditMode = false
	}

	s.SetNextRowIndex(direction)
	log.Debugf("current row %d", s.CurrentRowIndex)
	json, err := s.GetJsonFromCurrentRow()
	if err != nil {
		log.Errorf("minidoc from %v failed: %v", json, err)
		return
	}

	doc, err := MiniDocFrom(json)
	if err != nil {
		log.Errorf("minidoc from %v failed: %v", json, err)
		return
	}

	jh := NewJSONHandler(json)

	content := ""
	s.Detail.SetTitle(jh.string("type") + ":" + jh.string("id"))

	//fields := jh.fields()
	//keys := make([]string, 0, len(fields))
	//for key := range fields {
	//	keys = append(keys, key)
	//}
	//sort.Strings(keys)

	for _, fieldName := range doc.GetDisplayFields() {
		if fieldName == "type" || fieldName == "id" {
			continue
		}

		fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
		//s.debug("preview field for " + fieldNameCleaned)
		v := jh.string(fieldName)
		content += "\n"
		content += fmt.Sprintf("  [white]%s:[white] [darkcyan]%s[white]\n", fieldNameCleaned, v)
	}

	s.Detail.Clear()
	fmt.Fprintf(s.Detail, "%s", content)
}

func (s *Search) LoadEdit() {
	s.ResultList.SetSelectable(false, false)

	json, err := s.GetJsonFromCurrentRow()
	if err != nil {
		log.Debugf("error getting json from curr row: %v", err)
		return
	}

	s.EditForm = NewEdit(s, json).Form

	if !s.IsEditMode {
		s.Columns.RemoveItem(s.Detail)
		s.Columns.AddItem(s.EditForm, 0, 5, true)
		s.IsEditMode = true
		s.App.SetFocus(s.EditForm)
	}
	s.App.Draw()
}

func (s *Search) GetJsonFromCurrentRow() (interface{}, error) {
	rowIndex := s.CurrentRowIndex
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

	doc, err := s.App.BucketHandler.Read(id, doctype)
	if err != nil {
		log.Debugf("read error: %v", err)
		return nil, err
	}
	return doc, nil
}

func (s *Search) UnLoadEdit() {
	s.GoToSearchResult()
	s.LoadPreview(DIRECTION_NONE)
}
