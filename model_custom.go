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
	return Jsonize(u)
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
	return "[yellow]o[white] <= open url in browser"
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

func (m *URLDoc) GetViEditFields() []string {
	return []string{"description"}
}

// --------------------------------------------------------------------------------

type NoteDoc struct {
	BaseDoc
	Note string `json:"note"`
}

func (n *NoteDoc) GetJSON() interface{} {
	return Jsonize(n)
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

func (n *NoteDoc) GetViEditFields() []string {
	return []string{"note"}
}

func (n *NoteDoc) GetMarkdown() string {
	return fmt.Sprintf(`###%s
  %s`, n.Title, n.Note)
}

// --------------------------------------------------------------------------------

type ToDoDoc struct {
	BaseDoc
	Task string `json:"task"`
	Done bool   `json:"done"`
}

func (d *ToDoDoc) GetJSON() interface{} {
	return Jsonize(d)
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

func (d *ToDoDoc) HandleEvent(event *tcell.EventKey) {
	eventKey := event.Key()

	switch eventKey {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'd':
			log.Debugf("marking %s as done or undone")
			if d.Done {
				d.Done = false
			} else {
				d.Done = true
			}
		}
	}
}

func (d *ToDoDoc) GetAvailableActions() string {
	return "[yellow]d[white] <= mark task as done or undone"
}

func (d *ToDoDoc) GetMarkdown() string {
	return fmt.Sprintf(`###%s
  %s`, d.Task)
}
