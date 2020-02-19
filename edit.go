package minidoc

import (
	"github.com/rivo/tview"
	"strconv"
	"strings"
)

type Edit struct {
	Form   *tview.Form
	Search *Search
	json   interface{}
	debug  func(string)
}

func NewEdit(s *Search, json interface{}) *Edit {
	edit := &Edit{
		Search: s,
		debug:  s.App.DebugView.Debug,
		json:   json,
	}

	f := NewEditorForm(json)

	f.AddButton("Update", edit.UpdateAction)
	f.AddButton("Delete", edit.DeleteAction)
	f.AddButton("Cancel", edit.CancelAction)
	f.SetBorderPadding(1, 1, 2, 2)
	f.SetBorder(true)

	edit.Form = f

	return edit
}

func NewEditorForm(json interface{}) *tview.Form {
	f := tview.NewForm()
	doc, err := MiniDocFrom(json)
	if err != nil {
		log.Errorf("converting to minidoc failed json=%v: %v", json, err)
	}

	jh := NewJSONHandler(json)
	f.SetTitle(jh.string("type") + ":" + jh.string("id"))

	for _, fieldName := range doc.GetDisplayFields() {
		if fieldName == "type" || fieldName == "id" || fieldName == "created_date" {
			continue
		}

		fieldtype := jh.fieldtype(fieldName)
		fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
		//edit.debug("adding input field for " + fieldNameCleaned)
		switch fieldtype {
		case "string":
			f.AddInputField(fieldNameCleaned+":", jh.string(fieldName), 60, nil, nil)
		case "bool":
			f.AddCheckbox(fieldNameCleaned+":", jh.bool(fieldName), nil)
		}
	}
	return f
}

func (e *Edit) Update(doc MiniDoc) error {
	_, err := e.Search.App.BucketHandler.Write(doc)
	if err != nil {
		return err
	}
	err = e.Search.App.IndexHandler.Index(doc)
	return err
}

func (e *Edit) Delete(doc MiniDoc) error {
	err := e.Search.App.BucketHandler.Delete(doc)
	if err != nil {
		return err
	}
	err = e.Search.App.IndexHandler.Delete(doc)
	return err
}

func (e *Edit) UpdateAction() {
	f := e.Form
	jh := NewJSONHandler(e.json)

	ExtractFieldValues(jh, f)

	doc, err := MiniDocFrom(e.json)
	if err != nil {
		log.Errorf("MiniDocFrom failed: %v", err)
		return
	}
	log.Debugf("minidoc from json: %v", e.json)

	err = e.Update(doc)
	if err != nil {
		log.Errorf("updating %v failed: %v", doc, err)
		return
	}

	e.Search.UnLoadEdit()
}

func ExtractFieldValues(jh *JSONHandler, f *tview.Form) {
	for fieldName, _ := range jh.fields() {
		if fieldName == "type" || fieldName == "id" || fieldName == "created_date" || fieldName == "fragments" {
			continue
		}

		fieldtype := jh.fieldtype(fieldName)
		var v string
		switch fieldtype {
		case "string":
			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			sv := GetInputValue(f, fieldNameCleaned+":")
			jh.set(fieldName, sv)
		case "float64":
			fieldNameCleaned := strings.Replace(fieldName, "_", " ", -1)
			v = GetInputValue(f, fieldNameCleaned+":")
			fv, _ := strconv.ParseFloat(v, 64)
			jh.set(fieldName, fv)
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
	ConfirmDeleteModal(e.Search, e.json, e.Delete)
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
					s.SetCurrentRowIndex(UP)
				}
				// if at the beginning move down
				if s.CurrentRowIndex == 0 {
					s.SetCurrentRowIndex(DIRECTION_NONE)
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
