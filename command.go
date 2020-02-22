package minidoc

import "strings"

func (s *Search) HandleCommand(command string) {
	command = command[1:]
	terms := strings.Split(command, " ")
	verb := terms[0]
	log.Debugf("command terms %s", terms)

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
		outputDoctype := terms[1]
		if outputDoctype == "markdown" {
			filepath := terms[2]
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
			if WriteToFile(filepath, markdown) {
				return
			}
			OpenVim(s.App, filepath)
			s.App.StatusBar.SetText("[white]markdown generated[white]")

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

	doc = EditWithVim(app, doc)
	id, err := app.DataHandler.Write(doc)
	if err != nil {
		log.Errorf("error writing: %v", err)
		return err
	}
	doc.SetID(id)

	newPage := NewNewPage(doc)
	app.PagesHandler.AddPage(app, newPage)
	app.PagesHandler.GotoPageByTitle("New")
	app.SetFocus(newPage.Form)
	return err
}
