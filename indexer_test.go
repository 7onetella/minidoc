package minidoc

import (
	"testing"
)

func TestIndexHandler_Index(t *testing.T) {
	db := NewBucketHandler()
	indexer := NewIndexHandler()
	doc := GetTestUrlMiniDoc()
	db.Write(doc)

	if err := indexer.Index(doc); err != nil {
		t.Fail()
	}
}

func TestIndexHandler_Search(t *testing.T) {
	db := NewBucketHandler()
	indexer := NewIndexHandler()
	doc := GetTestUrlMiniDoc()
	db.Write(doc)
	indexer.Index(doc)

	docs := indexer.Search("baz")

	if len(docs) == 0 {
		t.Fail()
	}
}

func TestIndexHandler_Delete(t *testing.T) {
	db := NewBucketHandler()
	indexer := NewIndexHandler()
	doc := GetTestUrlMiniDoc()
	db.Write(doc)

	if err := indexer.Delete(doc); err != nil {
		t.Fail()
	}
}

func GetTestUrlMiniDoc() *URLDoc {
	doc := &URLDoc{
		BaseDoc: BaseDoc{
			ID:          uint32(0),
			Type:        "url",
			Title:       "test title foo",
			Description: "test description bar",
			Tags:        "test tags baz",
		},
		URL:        "http://foo.com",
		WatchLater: true,
	}
	return doc
}

func GetTestNoteMiniDoc() *NoteDoc {
	doc := &NoteDoc{
		BaseDoc: BaseDoc{
			ID:          uint32(0),
			Type:        "note",
			Title:       "test title foo",
			Description: "test description bar",
			Tags:        "test tags baz",
		},
		Note: "my note foo",
	}
	return doc
}

func GetTestTodoMiniDoc() *ToDoDoc {
	doc := &ToDoDoc{
		BaseDoc: BaseDoc{
			ID:   uint32(0),
			Type: "todo",
			Tags: "test tags baz",
		},
		Task: "my task",
		Done: true,
	}
	return doc
}
