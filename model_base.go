package minidoc

import (
	"github.com/gdamore/tcell"
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
	SetSearchFragments(string)
	GetJSON() interface{}
	SetCreatedDate(string)
	GetDisplayFields() []string
	IsSelected() bool
	SetIsSelected(bool)
	IsSelectedString() string
	HandleEvent(event *tcell.EventKey)
	GetMarkdown() string
	GetAvailableActions() string
	GetViEditFields() []string
	GetToggleValueAsString() string
	SetToggle(toggle bool)
	GetToggle() bool
	IsTogglable() bool
}

type BaseDoc struct {
	CreatedDate string `json:"created_date"`
	ID          uint32 `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Fragments   string `json:"fragments"`
	Selected    bool   `json:"selected"`
	Toggled     bool   `json:"toggle"`
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
	return m.Fragments
}

func (m *BaseDoc) SetSearchFragments(fragments string) {
	m.Fragments = fragments
}

func (m *BaseDoc) GetJSON() interface{} {
	return Jsonize(m)
}

func (m *BaseDoc) SetCreatedDate(createdDate string) {
	m.CreatedDate = createdDate
}

func (m *BaseDoc) IsSelected() bool {
	return m.Selected
}

func (m *BaseDoc) SetIsSelected(selected bool) {
	m.Selected = selected
}

func (m *BaseDoc) IsSelectedString() string {
	if m.Selected {
		return "*️" // let's use ✓️ for todo
	}
	return " "
}

func (m *BaseDoc) HandleEvent(event *tcell.EventKey) {
}

func (m *BaseDoc) GetMarkdown() string {
	return "### BaseDoc"
}

func (M *BaseDoc) GetAvailableActions() string {
	return "nothing really"
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

func (m *BaseDoc) GetViEditFields() []string {
	return []string{}
}

func (m *BaseDoc) GetToggleValueAsString() string {
	return " "
}

func (m *BaseDoc) SetToggle(toggle bool) {
	m.Toggled = toggle
}

func (m *BaseDoc) GetToggle() bool {
	return m.Toggled
}

func (m *BaseDoc) IsTogglable() bool {
	return m.Type == "todo" || m.Type == "url"
}
