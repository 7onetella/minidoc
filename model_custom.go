package minidoc

import (
	"fmt"
	"github.com/gdamore/tcell"
	"strings"
)

var doctypes = []string{"url", "note", "todo"}

var indexedFields = map[string][]string{
	"url":  {"title", "description", "tags"},
	"note": {"title", "note", "tags"},
	"todo": {"task", "done", "tags"},
}

type URLDoc struct {
	BaseDoc
	URL        string `json:"url"`
	WatchLater bool   `json:"watch_later"`
}

func (u *URLDoc) GetJSON() interface{} {
	return JsonMapFrom(u)
}

func (m *URLDoc) HandleEvent(event *tcell.EventKey) {
	eventKey := event.Key()

	switch eventKey {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'o':
			log.Debugf("opening url %s in the browser", m.URL)
			_, err := Execute(strings.Split("open "+m.URL, " "))
			if err != nil {
				log.Errorf("urldoc exection: %v", err)
				return
			}
		}
	}
}

func (M *URLDoc) GetAvailableActions() string {
	return "[yellow]o[white] <= open url in browser | [yellow]t[white] toggle watch later"
}

func (m *URLDoc) GetMarkdown() string {
	return fmt.Sprintf(`link: [%s](%s)`, m.Title, m.URL)
}

func (u *URLDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"url",
		"watch_later",
		"title",
		"description",
		"tags",
		"created_date",
	}
}

func (u *URLDoc) GetEditFields() []string {
	return []string{
		"url",
		"watch_later",
		"title",
		"description",
		"tags",
	}
}

func (m *URLDoc) GetToggleValueAsString() string {
	if m.WatchLater {
		return "✓️"
	}
	return " "
}

func (m *URLDoc) SetToggle(toggle bool) {
	m.WatchLater = toggle
}

func (m *URLDoc) GetToggle() bool {
	return m.WatchLater
}

// --------------------------------------------------------------------------------

type NoteDoc struct {
	BaseDoc
	Note string `json:"note"`
}

func (n *NoteDoc) GetJSON() interface{} {
	return JsonMapFrom(n)
}

func (n *NoteDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"title",
		"note",
		"tags",
		"created_date",
	}
}

func (n *NoteDoc) GetEditFields() []string {
	return []string{
		"title",
		"note",
		"tags",
	}
}

func (n *NoteDoc) GetViEditFields() []string {
	return []string{"note"}
}

func (n *NoteDoc) GetMarkdown() string {
	return fmt.Sprintf(`## %s
%s
%s
%s`, n.Title, "```", n.Note, "```")
}

// --------------------------------------------------------------------------------

type ToDoDoc struct {
	BaseDoc
	Task string `json:"task"`
	Done bool   `json:"done"`
}

func (d *ToDoDoc) GetJSON() interface{} {
	return JsonMapFrom(d)
}

func (d *ToDoDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"task",
		"done",
		"tags",
		"created_date",
	}
}

func (d *ToDoDoc) GetEditFields() []string {
	return []string{
		"task",
		"done",
		"tags",
	}
}

func (d *ToDoDoc) HandleEvent(event *tcell.EventKey) {
	//eventKey := event.Key()
	//
	//switch eventKey {
	//case tcell.KeyRune:
	//	switch event.Rune() {
	//	case 'd':
	//		log.Debugf("marking %s as done or undone")
	//	}
	//}
}

func (d *ToDoDoc) GetAvailableActions() string {
	return "[yellow]t[white] <= toggle done"
}

func (d *ToDoDoc) GetMarkdown() string {
	return fmt.Sprintf(`###%s
  %s`, d.Task)
}

func (d *ToDoDoc) GetTitle() string {
	return d.Task
}

func (d *ToDoDoc) GetToggleValueAsString() string {
	if d.Done {
		return "✓️"
	}
	return " "
}

func (d *ToDoDoc) SetToggle(toggle bool) {
	d.Done = toggle
}

func (d *ToDoDoc) GetToggle() bool {
	return d.Done
}
