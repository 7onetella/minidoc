package minidoc

import (
	"github.com/7onetella/minidoc/config"
	"github.com/mitchellh/go-homedir"
	"strings"
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
			s.App.StatusBar.SetText("[red]please specify file extension[white]")
			return
		}

		filename := tokens[0]
		extension := tokens[1]
		generatedDocPath := home + config.Config().GetString("generated_doc_path")

		markdownFilePath := generatedDocPath + filename + ".md"
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
			s.App.StatusBar.SetText("[red]generating markdown: " + err.Error() + "[white]")
			return
		}
		OpenVim(s.App, markdownFilePath)
		s.App.StatusBar.SetText("[yellow]markdown generated[white]")

		// convert markdown to pdf if the extension is pdf
		if extension == "pdf" {
			// does pandoc exist in path?
			if !DoesBinaryExists("pandoc") {
				s.App.StatusBar.SetText("[red]please install pandoc to generate pdf[white]")
				return
			}
			pdfFiePath := generatedDocPath + filename + ".pdf"
			err := Exec([]string{"pandoc", "-s", markdownFilePath, "-o", pdfFiePath})
			if err != nil {
				s.App.StatusBar.SetText("[red]generating pdf: " + err.Error() + "[white]")
				return
			}
			s.App.StatusBar.SetText("[yellow]pdf generated[white]")

			err = Exec([]string{"open", pdfFiePath})
			if err != nil {
				s.App.StatusBar.SetText("[red]opening pdf: " + err.Error() + "[white]")
				return
			}
			return
		}

		err = Exec([]string{"open", markdownFilePath})
		if err != nil {
			s.App.StatusBar.SetText("[red]opening pdf: " + err.Error() + "[white]")
			return
		}

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

		s.App.StatusBar.SetText("[white]listing docs by type " + doctype)
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
		s.App.StatusBar.SetText("[white]tagged[white]")
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
		s.App.StatusBar.SetText("[white]untagged[white]")
	}
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
