package minidoc

import (
	"fmt"
	"os"
	"testing"
)

func init() {
	os.MkdirAll(".minidoc", os.ModePerm)
}

func TestBucketHandler_Write(t *testing.T) {
	db := NewBucketHandler()
	doc := GetTestUrlMiniDoc()
	ID, err := db.Write(doc)
	if err != nil || ID == 0 {
		t.Log(err)
		t.Fail()
	}
}

func TestBucketHandler_Read(t *testing.T) {
	for i := 0; i < 0; i++ {
		db := NewBucketHandler()
		doc := GetTestUrlMiniDoc()
		ID, err := db.Write(doc)
		if err != nil || ID == 0 {
			t.Fail()
		}
		doc2, err := db.Read(ID, doc.GetType())
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		fmt.Println(doc2)
	}
	for i := 0; i < 10; i++ {
		db := NewBucketHandler()
		doc := GetTestNoteMiniDoc()
		ID, err := db.Write(doc)
		if err != nil || ID == 0 {
			t.Fail()
		}
		doc2, err := db.Read(ID, doc.GetType())
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		fmt.Println(doc2)
	}
	for i := 0; i < 10; i++ {
		db := NewBucketHandler()
		doc := GetTestTodoMiniDoc()
		ID, err := db.Write(doc)
		if err != nil || ID == 0 {
			t.Fail()
		}
		doc2, err := db.Read(ID, doc.GetType())
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		fmt.Println(doc2)
	}
}

func TestBucketHandler_Delete(t *testing.T) {
	db := NewBucketHandler()
	doc := GetTestNoteMiniDoc()
	ID, err := db.Write(doc)
	if err != nil || ID == 0 {
		t.Log(err)
		t.Fail()
	}
	err = db.Delete(doc)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	_, err = db.Read(ID, doc.GetType())
	if err == nil {
		t.Log("we should get record not found error")
		t.Fail()
	}
}
