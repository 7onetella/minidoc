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
			var doc MiniDoc
			switch doctype {
			case "url":
				doc = &URLDoc{}
				doc.SetType("url")
			case "note":
				doc = &NoteDoc{}
				doc.SetType("note")
			case "todo":
				doc = &ToDoDoc{}
				doc.SetType("todo")
			}
			doc = s.EditWithVim(doc)
			id, err := s.App.DataHandler.Write(doc)
			if err != nil {
				log.Errorf("error writing: %v", err)
				return
			}
			doc.SetID(id)

			newPage := NewNewPage(doc)
			s.App.PagesHandler.AddPage(s.App, newPage)
			s.App.PagesHandler.GotoPageByTitle("New")
			s.App.SetFocus(newPage.Form)
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
			s.App.StatusBar.SetText("[green]markdown generated[white]")

		}
	case "list":
		doctype := terms[1]
		jsons, err := s.App.DataHandler.BucketHandler.ReadAll(doctype)
		if err != nil {
			log.Errorf("error reading docs by type: %v", err)
			return
		}
		result := make([]MiniDoc, len(jsons))
		for i, json := range jsons {
			doc, err := MiniDocFrom(json)
			if err != nil {
				log.Errorf("error converting json to minidoc: %v", err)
				return
			}
			doc.SetSearchFragments(doc.GetTitle())
			result[i] = doc
		}

		s.UpdateResult(result)
		s.ResultList.ScrollToBeginning()
		s.SelectRow(0)
		s.App.SetFocus(s.SearchBar)

		s.App.StatusBar.SetText("[green]listing docs by type[white]")
	}
}
