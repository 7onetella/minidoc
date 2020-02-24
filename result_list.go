package minidoc

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"reflect"
	"strings"
)

type ResultList struct {
	*tview.Table
	Search *Search
}

func NewResultList(s *Search) *ResultList {
	rl := &ResultList{
		tview.NewTable(),
		s,
	}

	rl.SetBorders(false).
		SetSeparator(' ').
		SetSelectedStyle(tcell.ColorGray, tcell.ColorWhite, tcell.AttrNone).
		SetTitle("Results")

	rl.SetBorder(true)
	rl.SetBorderPadding(1, 1, 2, 2)

	rl.SetInputCapture(rl.InputCapture())

	return rl
}

func (rl *ResultList) InsertColumns(size int) {
	for i := 0; i < size; i++ {
		rl.InsertColumn(0)
	}
}

func (rl *ResultList) InputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		s := rl.Search
		//log.Debug("EventKey: " + event.Name())
		eventKey := event.Key()

		switch eventKey {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'i':
				s.Edit()
			case 'j':
				// if the current mode is edit then remove edit and load detail
				s.Preview(DOWN)
			case 'k':
				s.Preview(UP)
			case 'e':
				doc, err := s.LoadMiniDocFromDB(s.CurrentRowIndex)
				if err != nil {
					log.Debugf("error getting json from curr row: %v", err)
					return event
				}
				doc, changed := EditWithVim(rl.Search.App, doc)
				if changed {
					_, err = s.App.DataHandler.Write(doc)
					if err != nil {
						log.Errorf("error writing: %v", err)
					}
				}
				s.Preview(DIRECTION_NONE)
				return nil
			case ' ':
				s.ToggleSelected()
			case 't':
				s.ToggleTogglable()
				s.Preview(DIRECTION_NONE)
			default:
				return s.DelegateEventHandlingMiniDoc(event)
			}

		case tcell.KeyCtrlA:
			s.SelectAllRows()
			return nil
		case tcell.KeyCtrlK:
			rl.MoveRow(UP)
			return nil
		case tcell.KeyCtrlJ:
			rl.MoveRow(DOWN)
			return nil
		case tcell.KeyCtrlT:
			s.ToggleAllRows()
			return nil
		case tcell.KeyTab:
			s.GoToSearchBar(false)
			return nil
		case tcell.KeyEnter:
			s.Preview(DIRECTION_NONE)
			return nil
		case tcell.KeyCtrlD:
			s.BatchDeleteConfirmation()
		case tcell.KeyCtrlSpace:
			s.GoToSearchBar(true)
		default:
			return s.DelegateEventHandlingMiniDoc(event)
		}

		return event
	}
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

type CellData struct {
	ref  interface{}
	text string
}

func (rl *ResultList) SetColumnCells(rowIndex int, cd []CellData) {
	for i, c := range cd {
		rl.SetCell(rowIndex, i, NewCell(c.ref, c.text, tcell.ColorWhite))
	}
}

// pad empty space to keep the result row width wider than few character wide
const cellpadding = "                                                             "

func (rl *ResultList) GetCellRefUint32(rowIndex, colIndex int) (uint32, error) {
	ref := rl.GetCell(rowIndex, colIndex).GetReference()
	v, ok := ref.(uint32)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] column not uint32 but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return 0, fmt.Errorf(msg)
	}
	return v, nil
}

func (rl *ResultList) GetCellRefString(rowIndex, colIndex int) (string, error) {
	ref := rl.GetCell(rowIndex, colIndex).GetReference()
	v, ok := ref.(string)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] column not string but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return "", fmt.Errorf(msg)
	}
	return v, nil
}

func (rl *ResultList) GetCellRefBool(rowIndex, colIndex int) (bool, error) {
	ref := rl.GetCell(rowIndex, colIndex).GetReference()
	v, ok := ref.(bool)
	if !ok {
		msg := fmt.Sprintf("ref for row index[%d] column not string but is %v", rowIndex, reflect.TypeOf(ref))
		log.Errorf(msg)
		return false, fmt.Errorf(msg)
	}
	return v, nil
}

func (rl *ResultList) LoadMiniDocFromDB(rowIndex int) (MiniDoc, error) {
	id, _ := rl.GetCellRefUint32(rowIndex, idColumnIndex)

	doctype, _ := rl.GetCellRefString(rowIndex, typeColumnIndex)

	fragments, _ := rl.GetCellRefString(rowIndex, fragmentsColumnIndex)

	isSelected, _ := rl.GetCellRefBool(rowIndex, selectedColumnIndex)

	doc, err := rl.Search.App.BucketHandler.Read(id, doctype)
	if err != nil {
		log.Debugf("read error: %v", err)
		return nil, err
	}

	doc.SetIsSelected(isSelected)
	doc.SetSearchFragments(fragments)

	return doc, nil
}

func (rl *ResultList) UpdateRow(row int, doc MiniDoc) {
	doctype := doc.GetType()
	doctype = strings.TrimSpace(doctype)
	fragments := doc.GetSearchFragments()
	selected := doc.IsSelected()

	if doc.IsTogglable() {
		// swap it out with the one from db
		docFromDB, _ := rl.Search.App.DataHandler.BucketHandler.Read(doc.GetID(), doc.GetType())
		doc = docFromDB
		doc.SetSearchFragments(fragments)
		doc.SetIsSelected(selected)
	}

	cd := []CellData{
		CellData{doctype, doc.GetIDString()},
		CellData{doc.IsSelected(), doc.IsSelectedString()},
		CellData{doc.GetToggle(), doc.GetToggleValueAsString()},
		CellData{fragments + cellpadding, fragments + cellpadding},
		CellData{doc.GetID(), ""},
	}
	rl.SetColumnCells(row, cd)
}

func (rl *ResultList) MoveRow(direction int) {
	prevRow := rl.Search.CurrentRowIndex
	doc, err := rl.Search.LoadMiniDocFromDB(rl.Search.CurrentRowIndex)
	if err != nil {
		log.Debugf("error getting json from curr row: %v", err)
		return
	}

	rl.Search.UpdateCurrRowIndexFromSelectedRow(direction)
	row := rl.Search.CurrentRowIndex

	switch direction {
	case UP:
		rl.InsertRow(row)
		rl.UpdateRow(row, doc)
		rl.RemoveRow(prevRow + 1)
		rl.Search.SelectRow(row)
	case DOWN:
		rl.InsertRow(row + 1)
		rl.UpdateRow(row+1, doc)
		rl.RemoveRow(prevRow)
		rl.Search.SelectRow(row)
	}
}
