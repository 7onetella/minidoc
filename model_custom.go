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

// --------------------------------------------------------------------------------
// URL Doc
// --------------------------------------------------------------------------------
type URLDoc struct {
	BaseDoc
	URL        string `json:"url"`
	WatchLater bool   `json:"watch_later"`
}

func (d *URLDoc) GetJSON() interface{} {
	return JsonMapFrom(d)
}

func (d *URLDoc) HandleEvent(event *tcell.EventKey) {
	eventKey := event.Key()

	switch eventKey {
	case tcell.KeyRune:
		switch event.Rune() {
		case 'o':
			log.Debugf("opening url %s in the browser", d.URL)
			_, err := Execute(strings.Split("open "+d.URL, " "))
			if err != nil {
				log.Errorf("urldoc exection: %v", err)
				return
			}
		}
	}
}

func (d *URLDoc) GetAvailableActions() string {
	return "[yellow]o[white]pen url in browser | [yellow]t[white]oggle watch later"
}

func (d *URLDoc) GetMarkdown() string {
	return fmt.Sprintf(`link: [%s](%s)`, d.Title, d.URL)
}

func (d *URLDoc) GetDisplayFields() []string {
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

func (d *URLDoc) GetEditFields() []string {
	return []string{
		"url",
		"watch_later",
		"title",
		"description",
		"tags",
	}
}

func (d *URLDoc) GetToggleValueAsString() string {
	if d.WatchLater {
		return "✓️"
	}
	return " "
}

func (d *URLDoc) SetToggle(toggle bool) {
	d.WatchLater = toggle
}

func (d *URLDoc) GetToggle() bool {
	return d.WatchLater
}

// --------------------------------------------------------------------------------
// Note Doc
// --------------------------------------------------------------------------------
type NoteDoc struct {
	BaseDoc
	Note string `json:"note"`
}

func (d *NoteDoc) GetJSON() interface{} {
	return JsonMapFrom(d)
}

func (d *NoteDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"title",
		"note",
		"tags",
		"created_date",
	}
}

func (d *NoteDoc) GetAvailableActions() string {
	return ""
}

func (d *NoteDoc) GetEditFields() []string {
	return []string{
		"title",
		"note",
		"tags",
	}
}

func (d *NoteDoc) GetViEditFields() []string {
	return []string{"note"}
}

func (d *NoteDoc) GetMarkdown() string {
	return fmt.Sprintf(`## %s
%s
%s
%s`, d.Title, "```", d.Note, "```")
}

// --------------------------------------------------------------------------------
// TODO Doc
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

// --------------------------------------------------------------------------------
// ShortcutKey Doc
// --------------------------------------------------------------------------------
type ShortcutKeyDoc struct {
	BaseDoc
	ShortCutKey string `json:"shortcut"`
}

func (d *ShortcutKeyDoc) GetJSON() interface{} {
	return JsonMapFrom(d)
}

func (d *ShortcutKeyDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"title",
		"shortcut",
		"tags",
		"created_date",
	}
}

func (d *ShortcutKeyDoc) GetEditFields() []string {
	return []string{
		"title",
		"shortcut",
		"tags",
	}
}
