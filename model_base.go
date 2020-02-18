package minidoc

import (
	"strconv"
)

type MiniDoc interface {
	GetID() uint32
	SetID(uint32)
	GetIDString() string
	GetType() string
	SetType(string)
	GetTitle() string
	GetDescription() string
	GetTags() string
	GetSearchFragments() string
	GetJSON() interface{}
	SetCreatedDate(string)
	GetDisplayFields() []string
}

type BaseDoc struct {
	CreatedDate     string `json:"created_date"`
	ID              uint32 `json:"id"`
	Type            string `json:"type"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Tags            string `json:"tags"`
	SearchFragments string `json:"fragments"`
}

func (m *BaseDoc) GetID() uint32 {
	return m.ID
}

func (m *BaseDoc) SetID(ID uint32) {
	m.ID = ID
}

func (m *BaseDoc) GetIDString() string {
	return m.Type + ":" + strconv.Itoa(int(m.ID))
}

func (m *BaseDoc) GetType() string {
	return m.Type
}

func (m *BaseDoc) SetType(doctype string) {
	m.Type = doctype
}

func (m *BaseDoc) GetTitle() string {
	return m.Title
}

func (m *BaseDoc) GetDescription() string {
	return m.Description
}

func (m *BaseDoc) GetTags() string {
	return m.Tags
}

func (m *BaseDoc) GetSearchFragments() string {
	return m.SearchFragments
}

func (m *BaseDoc) GetJSON() interface{} {
	return Jsonize(m)
}

func (m *BaseDoc) SetCreatedDate(createdDate string) {
	m.CreatedDate = createdDate
}

func (m *BaseDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"title",
		"created_date",
		"description",
		"tags",
	}
}
