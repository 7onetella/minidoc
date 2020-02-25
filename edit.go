package minidoc

import (
	"github.com/rivo/tview"
	"strconv"
	"strings"
)

type Edit struct {
	Form    *tview.Form
	Search  *Search
	jsonMap interface{}
}

func NewEditForm(s *Search, doc MiniDoc) *tview.Form {
	jsonMap := JsonMapFrom(doc)

	edit := &Edit{
		Search:  s,
		jsonMap: jsonMap,
	}

	f := NewFormWithFields(doc)
	f.AddButton("Update", edit.UpdateAction)
	f.AddButton("Delete", edit.DeleteAction)
	f.AddButton("Cancel", edit.CancelAction)
	f.SetBorderPadding(1, 1, 2, 2)
	f.SetBorder(true)
	edit.Form = f

	return f
}

func NewFormWithFields(doc MiniDoc) *tview.Form {
	f := tview.NewForm()

	json := JsonMapFrom(doc)
	j := NewJsonMapWrapper(json)
	f.SetTitle(j.string("type") + ":" + j.string("id"))
	if doc == nil {
		log.Errorf("doc is nil")
		return nil
	}
mainLoop:
	for _, fieldname := range doc.GetEditFields() {

		fieldtype := j.fieldtype(fieldname)

		// if vim already populated the field then skip
		for _, editfieldname := range doc.GetViEditFields() {
			if editfieldname == fieldname {
				value := j.string(fieldname)
				if len(value) > 0 {
					continue mainLoop
				}
			}
		}

		cleanfieldname := strings.Replace(fieldname, "_", " ", -1)
		label := cleanfieldname + ":"
		switch fieldtype {
		case "string":
			f.AddInputField(label, j.string(fieldname), 0, nil, nil)
		case "bool":
			f.AddCheckbox(label, j.bool(fieldname), nil)
		}
	}
	return f
}

func (e *Edit) UpdateAction() {
	f := e.Form
	jh := NewJsonMapWrapper(e.jsonMap)

	ExtractFieldValues(jh, f)

	doc, err := MiniDocFrom(e.jsonMap)
	if err != nil {
		log.Errorf("MiniDocFrom failed: %v", err)
		return
	}
	log.Debugf("minidoc from json: %v", e.jsonMap)

	_, err = e.Search.App.DataHandler.Write(doc)
	if err != nil {
		log.Errorf("updating %v failed: %v", doc, err)
		return
	}

	e.Search.UnLoadEdit()
}

func ExtractFieldValues(jh *JsonMapWrapper, f *tview.Form) {
	for fieldName, _ := range jh.fields() {
		if fieldName == "type" || fieldName == "id" || fieldName == "created_date" || fieldName == "fragments" {
			continue
		}

		fieldtype := jh.fieldtype(fieldName)
		switch fieldtype {
		case "string":
			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			sptr := GetInputValue(f, fieldNameCleaned+":")
			if sptr != nil {
				jh.set(fieldName, *sptr)
			}
		case "float64":
			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			sptr := GetInputValue(f, fieldNameCleaned+":")
			if sptr != nil {
				fv, _ := strconv.ParseFloat(*sptr, 64)
				jh.set(fieldName, fv)
			}
		case "bool":
			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			bv := GetCheckBoxChecked(f, fieldNameCleaned+":")
			jh.set(fieldName, bv)
		}
		if !jh.ok() {
			log.Errorf("setting %s: %v", fieldName, jh.err.Error())
		}
	}
}

func (e *Edit) DeleteAction() {
	ConfirmDeleteModal(e.Search, e.jsonMap, e.Search.App.DataHandler.Delete)
}

func ConfirmDeleteModal(s *Search, json interface{}, deleteFunc func(doc MiniDoc) error) {
	app := s.App

	modal := tview.NewModal().
		SetText("Do you really want to delete?").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				doc, err := MiniDocFrom(json)
				if err != nil {
					log.Errorf("minidoc from %v failed: %v", json, err)
					return
				}
				log.Debugf("deleting %v", doc)
				err = deleteFunc(doc)
				if err != nil {
					log.Errorf("deleting %v failed: %v", doc, err)
					return
				}
				log.Debugf("removing row %v", s.CurrentRowIndex)
				s.ResultList.RemoveRow(s.CurrentRowIndex)
				// if at the end move up
				if s.CurrentRowIndex == s.ResultList.GetRowCount() {
					s.UpdateCurrRowIndexFromSelectedRow(UP)
				}
				// if at the beginning move down
				if s.CurrentRowIndex == 0 {
					s.UpdateCurrRowIndexFromSelectedRow(DIRECTION_NONE)
				}
			}

			// calling this will trigger update of current row index again
			// s.GoToSearchResult()
			s.App.SetFocus(s.ResultList)
			s.App.Draw()

			if err := app.SetRoot(app.Layout, true).Run(); err != nil {
				panic(err)
			}
			log.Debugf("delete modal after restoring the view")
		})

	if err := app.SetRoot(modal, false).Run(); err != nil {
		panic(err)
	}
}

func (e *Edit) CancelAction() {
	e.Search.GoToSearchResult()
}
