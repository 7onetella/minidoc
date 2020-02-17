package minidoc

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type JSONHandler struct {
	json interface{}
	err  error
}

func NewJSONHandler(json interface{}) *JSONHandler {
	return &JSONHandler{
		json,
		nil,
	}
}

func (j *JSONHandler) ok() bool {
	return j.err == nil
}

func (j *JSONHandler) fields() map[string]interface{} {
	m, ok := j.json.(map[string]interface{})
	if ok {
		return m
	}
	j.err = fmt.Errorf("failed to convert to map")
	return nil
}

func (j *JSONHandler) string(property string) string {
	m := j.fields()
	if j.ok() {
		v, found := m[property]
		if found {
			vtype := reflect.TypeOf(v).String()
			switch vtype {
			case "string":
				v, _ := v.(string)
				return v
			case "float64":
				v, _ := v.(float64)
				i := int(v)
				s := fmt.Sprintf("%d", i)
				return s
			case "bool":
				b, _ := v.(bool)
				s := fmt.Sprintf("%v", b)
				return s
			}
		}
	}
	j.err = fmt.Errorf("property not found: " + property)
	return ""
}

func (j *JSONHandler) float64(property string) float64 {
	m := j.fields()
	if j.ok() {
		v, found := m[property].(float64)
		if found {
			return v
		}
	}

	j.err = fmt.Errorf("property not found: "+property+" type: %s", reflect.TypeOf(m[property]))
	return 0
}

func (j *JSONHandler) bool(property string) bool {
	m := j.fields()
	if j.ok() {
		v, found := m[property].(bool)
		if found {
			return v
		}
	}
	j.err = fmt.Errorf("property not found: " + property)
	return false
}

func (j *JSONHandler) set(property string, v interface{}) {
	m := j.fields()
	if !j.ok() {
		j.err = fmt.Errorf("error getting fields")
	}

	if j.ok() {
		_, found := m[property]
		if !found {
			j.err = fmt.Errorf("property not found: " + property)
			return
		}
		m[property] = v
	}
}

func (j *JSONHandler) fieldtype(property string) string {
	m := j.fields()
	if j.ok() {
		v, found := m[property]
		if found {
			return reflect.TypeOf(v).String()
		}
		j.err = fmt.Errorf("property not found")
		return ""
	}
	return ""
}

func Jsonize(d interface{}) interface{} {
	jsonBytes, err := json.Marshal(d)

	var jsonDoc interface{}
	err = json.Unmarshal(jsonBytes, &jsonDoc)
	if err != nil {
		return nil
	}
	return jsonDoc
}

func MiniDocFrom(d interface{}) (MiniDoc, error) {
	jh := NewJSONHandler(d)
	doctype := jh.string("type")

	if doctype == "url" {
		urldoc := &URLDoc{}
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Errorf("marshaling %s json=%v", doctype, d)
			return nil, err
		}

		err = json.Unmarshal(jsonBytes, urldoc)
		if err != nil {
			log.Errorf("unmarshaling %s json=%v", doctype, d)
			return nil, err
		}
		return urldoc, nil
	}

	if doctype == "note" {
		urldoc := &NoteDoc{}
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Errorf("marshaling %s json=%v", doctype, d)
			return nil, err
		}

		err = json.Unmarshal(jsonBytes, urldoc)
		if err != nil {
			log.Errorf("unmarshaling %s json=%v", doctype, d)
			return nil, err
		}
		return urldoc, nil
	}

	if doctype == "todo" {
		doc := &ToDoDoc{}
		jsonBytes, err := json.Marshal(d)
		if err != nil {
			log.Errorf("marshaling %s json=%v", doctype, d)
			return nil, err
		}

		err = json.Unmarshal(jsonBytes, doc)
		if err != nil {
			log.Errorf("unmarshaling %s json=%v", doctype, d)
			return nil, err
		}
		return doc, nil
	}

	return nil, fmt.Errorf("doctype %s not handled", doctype)
}
