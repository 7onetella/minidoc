package minidoc

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type JsonMapWrapper struct {
	jsonMap interface{}
	err     error
}

func NewJsonMapWrapper(jsonMap interface{}) *JsonMapWrapper {
	return &JsonMapWrapper{
		jsonMap,
		nil,
	}
}

func (j *JsonMapWrapper) ok() bool {
	return j.err == nil
}

func (j *JsonMapWrapper) fields() map[string]interface{} {
	m, ok := j.jsonMap.(map[string]interface{})
	if ok {
		return m
	}
	j.err = fmt.Errorf("failed to convert to map")
	return nil
}

func (j *JsonMapWrapper) string(field string) string {
	m := j.fields()
	if j.ok() {
		v, found := m[field]
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
	j.err = fmt.Errorf("field not found: " + field)
	return ""
}

func (j *JsonMapWrapper) float64(property string) float64 {
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

func (j *JsonMapWrapper) bool(property string) bool {
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

func (j *JsonMapWrapper) set(property string, v interface{}) {
	m := j.fields()
	if !j.ok() {
		j.err = fmt.Errorf("error getting fields")
		log.Errorf("json handler: %v", j.err)
	}

	if j.ok() {
		_, found := m[property]
		if !found {
			j.err = fmt.Errorf("property not found: " + property)
			log.Errorf("json handler: %v", j.err)
			return
		}
		m[property] = v
	}
}

func (j *JsonMapWrapper) fieldtype(property string) string {
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

func JsonMapFrom(minidoc MiniDoc) interface{} {
	jsonBytes, err := json.Marshal(minidoc)
	if err != nil {
		log.Errorf("marshaling to bytes minidoc=%v", minidoc)
		return nil
	}

	var jsonMap interface{}
	err = json.Unmarshal(jsonBytes, &jsonMap)
	if err != nil {
		log.Errorf("unmarshaling: %v", err)
		return nil
	}
	return jsonMap
}

func MiniDocFrom(jsonMap interface{}) (MiniDoc, error) {
	jh := NewJsonMapWrapper(jsonMap)
	doctype := jh.string("type")
	jsonBytes, err := json.Marshal(jsonMap)
	if err != nil {
		log.Errorf("marshaling %s json=%v", doctype, jsonMap)
		return nil, err
	}

	doc, err := NewDoc(doctype)
	if err != nil {
		log.Errorf("instantiating %s", doctype)
		return nil, err
	}

	err = json.Unmarshal(jsonBytes, doc)
	if err != nil {
		log.Errorf("unmarshaling %s json=%v", doctype, jsonMap)
		return nil, err
	}

	return doc, nil
}

func NewDoc(doctype string) (MiniDoc, error) {
	var doc MiniDoc
	switch doctype {
	case "note":
		doc = &NoteDoc{}
		doc.SetType("note")
	case "url":
		doc = &URLDoc{}
		doc.SetType("url")
	case "todo":
		doc = &ToDoDoc{}
		doc.SetType("todo")
	default:
		return nil, fmt.Errorf("doctype %s not handled", doctype)
	}
	return doc, nil
}
