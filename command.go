package minidoc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/7onetella/minidoc/config"
	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"os"
	"strings"
	"time"
)

func (s *Search) HandleCommand(command string) {
	// remove @symbol
	command = command[1:]

	terms := strings.Split(command, " ")
	verb := terms[0]
	log.Debugf("command terms %s", terms)

	// if only @verb is present, don't process further
	if len(terms) == 1 {
		return
	}

	switch verb {
	case "new":
		doctype := terms[1]
		if !s.App.PagesHandler.HasPage("New") {
			if err := NewDocFlow(doctype, s.App); err != nil {
				return
			}
			defer s.App.Draw()
		}
	case "generate":
		str := terms[1]

		home, err := homedir.Dir()
		if err != nil {
			log.Errorf("finding home : %s", err)
		}
		tokens := strings.Split(str, ".")
		if len(tokens) == 1 {
			s.App.SetStatus("[black:red] please specify file extension[white]")
			return
		}

		filename := tokens[0]
		extension := tokens[1]
		generatedDocPath := home + config.Config().GetString("generated_doc_path")
		if !strings.HasSuffix(generatedDocPath, "/") {
			generatedDocPath += "/"
		}

		markdownFilePath := generatedDocPath + filename + ".md"

		// if the extension is pdf, write to tmp folder, don't write to generatedDocPath
		if extension == "pdf" {
			UUID := uuid.New().String()
			markdownFilePath = fmt.Sprintf("/tmp/%s", UUID)
		}

		log.Debugf("generating %s", markdownFilePath)

		markdown := ""
		for i := 0; i < s.ResultList.GetRowCount(); i++ {
			log.Debugf("current row %d", s.CurrentRowIndex)
			doc, err := s.LoadMiniDocFromDB(i)
			if err != nil {
				log.Errorf("minidoc from failed: %v", err)
				return
			}

			if !doc.IsSelected() {
				log.Debugf("row %d not selected skipping", i)
				continue
			}

			markdown += doc.GetMarkdown() + "\n\n"
		}

		err = OpenFileIfNoneExist(markdownFilePath, markdown)
		if err != nil {
			s.App.SetStatus("[black:red]generating content: " + err.Error() + "[white]")
			return
		}
		OpenVim(s.App, markdownFilePath)
		s.App.SetStatus("[white:darkcyan]markdown generated[white]")
		s.Notify("Status", "markdown generated")

		// convert content to pdf if the extension is pdf
		if extension == "pdf" {
			// does pandoc exist in path?
			if !DoesBinaryExists("pandoc") {
				s.App.SetStatus("[black:red]please install pandoc to generate pdf[white]")
				return
			}
			pdfFiePath := generatedDocPath + filename + ".pdf"
			err := Exec([]string{"pandoc", "-s", markdownFilePath, "-o", pdfFiePath})
			if err != nil {
				s.App.SetStatus("[black:red]generating pdf: " + err.Error() + "[white]")
				return
			}
			s.App.SetStatus("[white:darkcyan]pdf generated[white]")

			err = Exec([]string{"open", pdfFiePath})
			if err != nil {
				s.App.SetStatus("[black:red]opening pdf: " + err.Error() + "[white]")
				return
			}

			// delete temporary content in /tmp folder
			DeleteFile(markdownFilePath)

			t := s.App.PagesHandler.GetPageItem("Generated").GetInstance().(*TreePage)
			t.RefreshRootNode()
			t.App.SetFocus(t.Tree)
			return
		}

		t := s.App.PagesHandler.GetPageItem("Generated").GetInstance().(*TreePage)
		t.RefreshRootNode()
		t.App.SetFocus(t.Tree)

		time.Sleep(250 * time.Millisecond)
		s.GoToSearchBar(true, "search")
		s.App.SetStatus("[white:darkcyan]markdown generated[white]")

	case "list":
		doctype := terms[1]
		docs, err := s.App.DataHandler.BucketHandler.ReadAll(doctype)
		if err != nil {
			log.Errorf("error reading docs by type: %v", err)
			return
		}
		result := make([]MiniDoc, len(docs))
		for i, doc := range docs {
			doc.SetSearchFragments(doc.GetTitle())
			result[i] = doc
		}

		s.UpdateResult(result)
		s.ResultList.ScrollToBeginning()
		s.SelectRow(0)
		s.App.SetFocus(s.SearchBar)

		s.App.SetStatus("[white:darkcyan] listing docs by type '" + doctype + "'")
	case "tag":
		tags := terms[1:]
		for i := 0; i < s.ResultList.GetRowCount(); i++ {
			log.Debugf("current row %d", s.CurrentRowIndex)
			doc, err := s.LoadMiniDocFromDB(i)
			if err != nil {
				log.Errorf("minidoc from failed: %v", err)
				return
			}

			if !doc.IsSelected() {
				log.Debugf("row %d not selected skipping", i)
				continue
			}
			dtags := strings.Split(doc.GetTags(), " ")
			str := ""
			for _, tag := range tags {
				if !contains(dtags, tag) {
					str += " " + tag
				}
			}
			str = strings.Join(dtags, " ") + str
			str = strings.TrimSpace(str)
			doc.SetTags(str)
			s.App.DataHandler.Write(doc)
		}
		s.App.SetStatus("[white]tagged[white]")
	case "untag":
		tags := terms[1:]
		for i := 0; i < s.ResultList.GetRowCount(); i++ {
			log.Debugf("current row %d", s.CurrentRowIndex)
			doc, err := s.LoadMiniDocFromDB(i)
			if err != nil {
				log.Errorf("minidoc from failed: %v", err)
				return
			}

			if !doc.IsSelected() {
				log.Debugf("row %d not selected skipping", i)
				continue
			}

			dtags := strings.Split(doc.GetTags(), " ")
			log.Debugf("dtags: %v", dtags)
			str := ""
			for _, d := range dtags {
				if !contains(tags, d) {
					str += d + " "
				}
			}
			str = strings.TrimSpace(str)
			log.Debugf("str: [%s]", str)
			doc.SetTags(str)
			_, err = s.App.DataHandler.Write(doc)
			if err != nil {
				log.Errorf("minidoc write failed: %v", err)
				return
			}

		}
		s.App.SetStatus("[white]untagged[white]")
	case "export":
		str := terms[1]

		home, err := homedir.Dir()
		if err != nil {
			log.Errorf("finding home : %s", err)
		}
		tokens := strings.Split(str, ".")
		if len(tokens) == 1 {
			s.App.SetStatus("[black:red]please specify file extension[white]")
			return
		}

		filename := tokens[0]
		extension := tokens[1]

		generatedDocPath := home + config.Config().GetString("generated_doc_path")

		backupFilePath := generatedDocPath + filename + "." + extension

		log.Debugf("exporting %s", backupFilePath)

		content := ""
		for i := 0; i < s.ResultList.GetRowCount(); i++ {
			log.Debugf("current row %d", s.CurrentRowIndex)
			doc, err := s.LoadMiniDocFromDB(i)
			if err != nil {
				log.Errorf("minidoc from failed: %v", err)
				return
			}

			if !doc.IsSelected() {
				log.Debugf("row %d not selected skipping", i)
				continue
			}

			jsonBytes, err := json.Marshal(doc)
			if err != nil {
				log.Errorf("marshalling doc: %v", err)
				return
			}

			content += string(jsonBytes) + "\n"
		}

		err = OpenFileIfNoneExist(backupFilePath, content)
		if err != nil {
			s.App.SetStatus("[black:red]exporting content: " + err.Error() + "[white]")
			return
		}
		s.App.SetStatus("[white:darkcyan]exporting done[white]")

		err = Exec([]string{"open", backupFilePath})
		if err != nil {
			s.App.SetStatus("[black:red]opening exported: " + err.Error() + "[white]")
			return
		}
	case "import":
		str := terms[1]
		isWeb := strings.HasPrefix(str, "http")
		var errored bool

		if isWeb {
			errored = ImportFromWeb(str, s)
		} else {
			errored = ImportFile(str, s)
		}

		if errored {
			return
		}

		s.App.SetStatus("[white:darkcyan]importing done[white]")
	}
}

func ImportFromWeb(str string, s *Search) bool {
	data, err := HTTPGet(str)
	if err != nil {
		s.App.SetStatus(fmt.Sprintf("[black:red]http: %v[white]", err))
		return true
	}
	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		errored := ImportLineByLine(line, s)
		if errored {
			continue
		}
	}

	content = strings.TrimSpace(content)
	if len(content) == 0 {
		s.App.SetStatus(fmt.Sprintf("[black:red]downloaded content empty: %v[white]", err))
		return true
	}
	return false
}

func ImportFile(str string, s *Search) bool {
	home, err := homedir.Dir()
	if err != nil {
		log.Errorf("finding home : %s", err)
	}

	generatedDocPath := home + config.Config().GetString("generated_doc_path")
	tokens := strings.Split(str, ".")
	if len(tokens) == 1 {
		s.App.SetStatus("[black:red]please specify file extension[white]")
		return true
	}

	filename := tokens[0]
	extension := tokens[1]

	backupFilePath := generatedDocPath + filename + "." + extension

	log.Debugf("importing %s", backupFilePath)

	file, err := os.Open(backupFilePath)
	if err != nil {
		log.Errorf("opening: %v", err)
		s.App.SetStatus(fmt.Sprintf("[black:red]opening: %v[white]", err))
		return true
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		errored := ImportLineByLine(line, s)
		if errored {
			return errored
		}
	}
	return false
}

func ImportLineByLine(line string, s *Search) bool {
	var err error
	line = strings.TrimSpace(line)

	var jsonMap interface{}
	err = json.Unmarshal([]byte(line), &jsonMap)
	if err != nil {
		log.Errorf("unmarshaling json=%v", jsonMap)
		s.App.SetStatus(fmt.Sprintf("[black:red]unmarshaling: %v[white]", err))
		return true
	}

	doc, err := MiniDocFrom(jsonMap)
	if err != nil {
		log.Errorf("minidoc from doc=%v", doc)
		s.App.SetStatus(fmt.Sprintf("[black:red]minidoc from: %v[white]", err))
		return true
	}
	// set the id to 0 so new sequence will be generated
	doc.SetID(0)
	s.App.DataHandler.Write(doc)
	return false
}

func NewDocFlow(doctype string, app *SimpleApp) error {
	doc, err := NewDoc(doctype)
	if err != nil {
		log.Errorf("instantiating %s", doctype)
		return err
	}

	doc, changed := EditWithVim(app, doc)
	if changed {
		id, err := app.DataHandler.Write(doc)
		if err != nil {
			log.Errorf("error writing: %v", err)
			return err
		}
		doc.SetID(id)
	}

	newPage := NewNewPage(doc)
	app.PagesHandler.AddPage(app, newPage)
	app.PagesHandler.GotoPageByTitle("New")
	app.SetFocus(newPage.Form)
	return err
}
