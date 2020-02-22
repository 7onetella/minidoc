package minidoc

type DataHandler struct {
	BucketHandler *BucketHandler
	IndexHandler  *IndexHandler
}

func (dh *DataHandler) Write(doc MiniDoc) (uint32, error) {
	id, err := dh.BucketHandler.Write(doc)
	if err != nil {
		return 0, err
	}
	err = dh.IndexHandler.Index(doc)
	return id, err
}

func (dh *DataHandler) Delete(doc MiniDoc) error {
	err := dh.BucketHandler.Delete(doc)
	if err != nil {
		return err
	}
	return dh.IndexHandler.Delete(doc)
}
