package minidoc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/joyrexus/buckets"
	"time"
)

const defaultDbFileName = ".minidoc/store.db"

type BucketHandler struct {
	debug  func(string)
	bx     *buckets.DB
	DBPath string
}

type BucketHandlerOption func(*BucketHandler)

func WithBucketHandlerDebug(debug func(string)) BucketHandlerOption {
	return func(bh *BucketHandler) {
		bh.debug = debug
	}
}

func WithBucketHandlerDBPath(path string) BucketHandlerOption {
	return func(bh *BucketHandler) {
		bh.DBPath = path
		log.Debug("opt() db path = " + path)
	}
}

func NewBucketHandler(opts ...BucketHandlerOption) *BucketHandler {
	bh := &BucketHandler{
		debug:  func(string) {},
		DBPath: defaultDbFileName,
	}

	for _, opt := range opts {
		opt(bh)
	}

	return bh
}

func (bh *BucketHandler) Write(doc MiniDoc) (uint32, error) {
	bx, err := buckets.Open(bh.DBPath)
	if err != nil {
		log.Errorf("error while opening bucket[%s] at %s: %v", doc.GetType(), bh.DBPath, err)
		return 0, err
	}
	defer bx.Close()
	doctype := doc.GetType()

	// Create a new bucket
	bucket, err := bx.New([]byte(doctype))
	if err != nil {
		log.Errorf("error while opening or creating bucket[%s]: %v", doctype, err)
		return 0, err
	}

	key := toBytes(doc.GetID())
	if doc.GetID() == 0 {
		log.Debugf("id == 0 doctype [%s] generating new sequence", doctype)

		key, err = NextSequence(bx, doctype)
		if err != nil {
			log.Errorf("error while next sequence doctype [%s]: %v", doctype, err)
		}
		doc.SetID(toUint32(key))
	}
	nowstr := time.Now().Format("2006-01-02 15:04:05")
	doc.SetCreatedDate(nowstr)

	data, err := json.Marshal(doc.GetJSON())
	if err != nil {
		log.Errorf("error while marshalling: %v", err)
		return 0, err
	}
	err = bucket.Put(key, data)
	if err != nil {
		log.Errorf("error while bucket put: %v", err)
	}

	return toUint32(key), nil
}

func (bh *BucketHandler) ReadAll(doctype string) ([]interface{}, error) {
	bx, err := buckets.Open(bh.DBPath)
	if err != nil {
		log.Errorf("error while opening bucket[%s] at %s: %v", doctype, bh.DBPath, err)
	}
	defer bx.Close()

	bucket, err := bx.New([]byte(doctype))
	if err != nil {
		log.Errorf("error while opening or creating bucket[%s]: %v", doctype, err)
		return nil, err
	}

	items, err := bucket.Items()
	if err != nil {
		log.Errorf("error while iterating all items in bucket[%s]: %v", doctype, err)
		return nil, err
	}

	docs := make([]interface{}, len(items))
	for i, item := range items {
		var jsonDoc interface{}
		err = json.Unmarshal(item.Value, &jsonDoc)
		if err != nil {
			log.Errorf("error while unmarshalling %s: %v", doctype, err)
			return nil, err
		}
		docs[i] = jsonDoc
	}

	return docs, nil
}

func (bh *BucketHandler) Read(key uint32, doctype string) (interface{}, error) {
	bx, err := buckets.Open(bh.DBPath)
	if err != nil {
		log.Errorf("error while opening bucket[%s] at %s: %v", doctype, bh.DBPath, err)
	}
	defer bx.Close()

	bucket, err := bx.New([]byte(doctype))
	if err != nil {
		log.Errorf("error while opening or creating bucket[%s]: %v", doctype, err)
		return nil, err
	}

	data, err := bucket.Get(toBytes(key))
	if err != nil {
		log.Errorf("error while getting item in bucket[%s] with key[%d]: %v", doctype, key, err)
		return nil, err
	}

	if data == nil {
		msg := fmt.Sprintf("item not found in bucket[%s] with key[%d]", doctype, key)
		log.Errorf(msg)
		return nil, fmt.Errorf(msg)
	}

	var jsonDoc interface{}
	err = json.Unmarshal(data, &jsonDoc)
	if err != nil {
		log.Errorf("error while unmarshalling in bucket[%s] with key[%d]: %v", doctype, key, err)
		return nil, err
	}

	return jsonDoc, nil
}

func (bh *BucketHandler) Delete(doc MiniDoc) error {
	key := toBytes(doc.GetID())
	doctype := doc.GetType()

	bx, err := buckets.Open(bh.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer bx.Close()

	bucket, err := bx.New([]byte(doctype))
	if err != nil {
		log.Errorf("error while opening or creating bucket[%s]: %v", doctype, err)
		return err
	}

	log.Debugf("deleting in bucket[%s] with key[%d]", doctype, key)
	err = bucket.Delete(key)
	if err != nil {
		log.Errorf("deleting in bucket[%s] with key[%d]: %v", doctype, key, err)
		return err
	}

	return nil
}

// NextSequence returns next sequence
func NextSequence(bx *buckets.DB, sequenceName string) ([]byte, error) {
	// get or create sequence bucket
	bucket, err := bx.New([]byte("_sequence"))
	if err != nil {
		log.Errorf("error while opening bucket for _sequence for doctype[%s]: %v", sequenceName, err)
		return nil, err
	}

	// get or create next sequence for given bucket
	key := []byte(sequenceName)

	next, err := bucket.Get(key)
	if err != nil {
		log.Errorf("error while retrieving current key for _sequence for doctype[%s]: %v", sequenceName, err)
		return nil, err
	}

	if next == nil || len(next) == 0 {
		bs := toBytes(1)
		bucket.Put(key, bs)
		return bs, nil
	}

	nextVal := toUint32(next)
	nextVal++
	err = bucket.Put(key, toBytes(nextVal))
	if err != nil {
		log.Errorf("error while putting next key for _sequence for doctype[%s]: %v", sequenceName, err)
		return nil, err
	}

	return toBytes(nextVal), nil
}

func toBytes(value uint32) []byte {
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, value)
	return bs
}

func toUint32(bs []byte) uint32 {
	return binary.LittleEndian.Uint32(bs)
}
