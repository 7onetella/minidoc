package minidoc

import "github.com/rivo/tview"

func GetInputValue(form *tview.Form, label string) *string {
	fi := form.GetFormItemByLabel(label)
	if fi == nil {
		log.Errorf("not found with label %s", label)
	}
	input, ok := fi.(*tview.InputField)
	if ok {
		v := input.GetText()
		return &v
	}
	return nil
}

func GetCheckBoxChecked(form *tview.Form, label string) bool {
	fi := form.GetFormItemByLabel(label)
	if fi == nil {
		return false
	}
	checkbox, ok := fi.(*tview.Checkbox)
	if ok {
		return checkbox.IsChecked()
	}
	return false
}
