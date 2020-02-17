package minidoc

import (
	"fmt"
	"github.com/blevesearch/bleve"
	_ "github.com/blevesearch/bleve/config"
	"github.com/blevesearch/bleve/search/highlight/highlighter/ansi"
	"strconv"
	"strings"

	//"github.com/blevesearch/bleve/search/highlight/format/ansi"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
	"github.com/blevesearch/bleve/analysis/lang/en"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
)

const (
	ErrorGeneric Error = iota
	ErrorCSVDoesNotExist
)

type Error int

func (e Error) Error() string {
	return errorMessages[e]
}

var errorMessages = map[Error]string{
	ErrorGeneric:         "generic error",
	ErrorCSVDoesNotExist: "cannot open csv, path does not exist",
}

type IndexHandler struct {
	debug     func(string)
	index     bleve.Index
	indexPath string
}

type IndexHandlerOption func(*IndexHandler)

func WithIndexHandlerDebug(debug func(string)) IndexHandlerOption {
	return func(ih *IndexHandler) {
		ih.debug = debug
	}
}

func WithIndexHandlerIndexPath(indexPath string) IndexHandlerOption {
	return func(ih *IndexHandler) {
		log.Debug("using index path: " + indexPath)
		ih.indexPath = indexPath
	}
}

const indexPathDefault = ".minidoc/index"

func NewIndexHandler(opts ...IndexHandlerOption) *IndexHandler {
	ih := &IndexHandler{
		indexPath: indexPathDefault,
		debug:     func(string) {},
	}

	for _, opt := range opts {
		opt(ih)
	}

	index, err := bleve.Open(ih.indexPath)
	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping, err := IndexMapping()
		if err != nil {
			log.Fatalf("error during loading index mapping: %v", err)
			return nil
		}
		index, err = bleve.New(ih.indexPath, mapping)
		if err != nil {
			log.Fatalf("error during loading index: %v", err)
			return nil
		}
	}
	ih.index = index
	log.Debug("index loaded successfully")

	return ih
}

func (ih *IndexHandler) Delete(doc MiniDoc) error {
	return ih.index.Delete(doc.GetIDString())
}

func (ih *IndexHandler) Index(doc MiniDoc) error {
	return ih.index.Index(doc.GetIDString(), doc.GetJSON())
}

// indexCmd will index given csv file
func (ih *IndexHandler) Search(queryString string) []MiniDoc {
	log.Debug("index search")

	// search for some text
	query := bleve.NewMatchQuery(queryString)
	search := &bleve.SearchRequest{
		Query:     query,
		Size:      100,
		From:      0,
		Explain:   false,
		Sort:      search.SortOrder{&search.SortScore{Desc: true}},
		Fields:    []string{"type", "title", "description", "tags"},
		Highlight: bleve.NewHighlightWithStyle(ansi.Name),
	}
	sr, err := ih.index.Search(search)
	if err != nil {
		log.Errorf("index search error: %v", err)
		return nil
	}
	stat := fmt.Sprintf("%d matches, showing %d through %d, took %s\n", sr.Total, sr.Request.From+1, sr.Request.From+len(sr.Hits), sr.Took)
	ih.debug(stat)
	//i.debug(sr.String())

	docs := make([]MiniDoc, sr.Hits.Len())
	for ri, hit := range sr.Hits {
		idparts := strings.Split(hit.ID, ":")
		v, _ := strconv.Atoi(idparts[1])
		log.Debugf("found minidoc[%d]", v)
		minidoc := &BaseDoc{
			ID: uint32(v),
		}
		if doctype, ok := hit.Fields["type"].(string); ok {
			minidoc.Type = doctype
		}
		if title, ok := hit.Fields["title"].(string); ok {
			minidoc.Title = title
		}
		if tags, ok := hit.Fields["description"].(string); ok {
			minidoc.Tags = tags
		}
		if tags, ok := hit.Fields["tags"].(string); ok {
			minidoc.Tags = tags
		}

		log.Debug("# of fragments: " + strconv.Itoa(len(hit.Fragments)))
		for _, fragments := range hit.Fragments {
			rv := ""
			for _, fragment := range fragments {
				// [43m [0m
				fragment = strings.ReplaceAll(fragment, "[43m", "[yellow]")
				fragment = strings.ReplaceAll(fragment, "[0m", "[white]")
				rv += fmt.Sprintf("%s", fragment)
			}
			minidoc.SearchFragments = rv
		}

		docs[ri] = minidoc
		log.Debugf("%s %f %s", hit.ID, hit.Score, hit.Fields["title"])
	}
	return docs
}

func IndexMapping() (*mapping.IndexMappingImpl, error) {

	// a generic reusable mapping for english text
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	// a generic reusable mapping for keyword text
	keywordFieldMapping := bleve.NewTextFieldMapping()
	keywordFieldMapping.Analyzer = keyword.Name

	indexMapping := bleve.NewIndexMapping()
	for _, doctype := range doctypes {
		documentMapping := DocumentMapping(indexedFields[doctype])
		indexMapping.AddDocumentMapping(doctype, documentMapping)
	}
	indexMapping.TypeField = "type"
	indexMapping.DefaultAnalyzer = "en"

	return indexMapping, nil
}

func DocumentMapping(indexedFields []string) *mapping.DocumentMapping {
	englishTextFieldMapping := bleve.NewTextFieldMapping()
	englishTextFieldMapping.Analyzer = en.AnalyzerName

	documentMapping := bleve.NewDocumentMapping()
	for _, f := range indexedFields {
		documentMapping.AddFieldMappingsAt(f, englishTextFieldMapping)
	}

	return documentMapping
}
