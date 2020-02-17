package minidoc

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

// --------------------------------------------------------------------------------

type ToDoDoc struct {
	BaseDoc
	Task string `json:"task"`
	Done bool   `json:"done"`
}

func (n *ToDoDoc) GetJSON() interface{} {
	return Jsonize(n)
}

func (n *ToDoDoc) GetDisplayFields() []string {
	return []string{
		"id",
		"type",
		"task",
		"done",
		"tags",
		"created_date",
	}
}
