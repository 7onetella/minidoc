package minidoc

import (
	"fmt"
	"github.com/7onetella/minidoc/config"
	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"
	"github.com/rivo/tview"
	"os"
	"path/filepath"
	"strings"
)

type TreePage struct {
	Title        string
	Tree         *tview.TreeView
	RootNode     *Node
	SelectedNode *Node
	App          *SimpleApp
	Nodes        map[string]*Node
	Detail       *tview.TextView
	Columns      *tview.Flex
}

type Node struct {
	Text     string
	Path     string
	Expand   bool
	Selected func()
	Children []*Node
}

func (t *TreePage) ConvertToTreeNode(target *Node, expanded bool) *tview.TreeNode {
	node := tview.NewTreeNode(target.Text).
		SetSelectable(target.Expand || target.Selected != nil).
		SetExpanded(expanded).
		SetReference(target)

	if target.Expand {
		node.SetColor(tcell.ColorGreen)
	} else if target.Selected != nil {
		node.SetColor(tcell.ColorWhite)
	}

	for _, child := range target.Children {
		node.AddChild(t.ConvertToTreeNode(child, true))
	}
	return node
}

func GetParentPath(childPath string) string {
	if strings.HasSuffix(childPath, "/") {
		s := childPath[:strings.LastIndex(childPath, "/")]
		return s[:strings.LastIndex(s, "/")]
	} else {
		return childPath[:strings.LastIndex(childPath, "/")]
	}
}

func (t *TreePage) Visit() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		nodes := t.Nodes

		if err != nil {
			log.Fatal(err)
		}

		if info.IsDir() {
			parent := nodes[GetParentPath(path)]
			if parent != nil {
				n := NewDirNode(info, path)
				parent.Children = append(parent.Children, n)
				//log.Debugf("adding child dir %s to %s", path, parent.Path)
				nodes[path] = n
			} else {
				//log.Debugf("adding root dir %s", path)
				nodes[path] = NewDirNode(info, path)
			}
			return nil
		}

		// file add as child node
		if !info.IsDir() {
			// find parent from root
			parent := nodes[GetParentPath(path)]
			if parent != nil {
				parent.Children = append(parent.Children, NewFileNode(info, path))
				//log.Errorf("adding child file %s to %s", path, parent.Path)
			}
			return nil
		}

		return nil
	}
}

func NewFileNode(info os.FileInfo, path string) *Node {
	return &Node{
		Text: info.Name(),
		Path: path,
		Selected: func() {
		},
	}
}

func NewDirNode(info os.FileInfo, path string) *Node {
	return &Node{
		Text: info.Name(),
		Selected: func() {
		},
		Path:     path,
		Children: []*Node{},
		Expand:   true,
	}
}

func NewTree() *TreePage {

	t := &TreePage{
		Tree:  tview.NewTreeView(),
		Title: "Generated",
	}
	t.Tree.SetAlign(false).SetTopLevel(0).SetGraphics(true).SetPrefixes(nil)

	t.RootNode = t.WalkGenDirAsNode(GetMiniDocGenDir())

	return t
}

func (t *TreePage) GetInstance() interface{} {
	return t
}

func (t *TreePage) WalkGenDirAsNode(minidocGenDir string) *Node {
	var nodes = map[string]*Node{}
	t.Nodes = nodes
	err := filepath.Walk(minidocGenDir, t.Visit())
	if err != nil {
		panic(err)
	}
	rootNode := nodes[minidocGenDir]
	rootNode.Text = rootNode.Path

	return rootNode
}

func GetMiniDocGenDir() string {
	cfg := config.Config()
	generatedDir := cfg.GetString("generated_doc_path")
	homedir, _ := homedir.Dir() // return path with slash at the end
	minidocGenDir := homedir + generatedDir
	if strings.HasSuffix(minidocGenDir, "/") {
		minidocGenDir = minidocGenDir[:strings.LastIndex(minidocGenDir, "/")]
	}
	return minidocGenDir
}

// TreeView demonstrates the tree view.
func (t *TreePage) Page() (title string, content tview.Primitive) {
	t.Tree.SetBorder(true).SetTitle(t.Title)
	t.Tree.SetBorderPadding(1, 2, 2, 1)

	t.RefreshRootNode()
	t.Tree.SetInputCapture(t.InputCapture())

	t.Detail = tview.NewTextView()
	t.Detail.SetBorder(true)
	t.Detail.SetTitle("Preview")
	t.Detail.SetDynamicColors(false)
	t.Detail.SetBorderPadding(1, 1, 2, 2)
	t.Detail.SetTextColor(tcell.ColorDarkCyan)

	t.Columns = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.Tree, 0, 5, true).
		AddItem(t.Detail, 0, 5, false)

	return t.Title, tview.NewFlex().AddItem(t.Columns, 0, 1, true)
}

func (t *TreePage) SetApp(app *SimpleApp) {
	t.App = app
}

func (t *TreePage) InputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		eventKey := event.Key()

		switch eventKey {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'e':
				path := t.SelectedNode.Path
				if filepath.Ext(path) == ".md" {
					OpenVim(t.App, path)
				}
			case 'o':
				path := t.SelectedNode.Path
				if filepath.Ext(path) == ".md" {
					OpenVim(t.App, path)
				}
				if filepath.Ext(path) == ".pdf" {
					err := Exec([]string{"open", path})
					if err != nil {
						t.App.SetStatus("[black:red]opening pdf: " + err.Error() + "[white]")
						return nil
					}
				}
			case 'n':
				if t.SelectedNode != nil {
					path := t.SelectedNode.Path
					if filepath.Ext(path) == ".md" || filepath.Ext(path) == ".pdf" {
						t.RenameFile(t.App, path)
						t.RefreshRootNode()
						t.App.SetFocus(t.Tree)
					}
				}
			case 'r':
				t.RefreshRootNode()
				t.App.SetFocus(t.Tree)
			case 'p':
				markdownFilePath := t.SelectedNode.Path
				filename := t.SelectedNode.Text[:strings.LastIndex(t.SelectedNode.Text, ".")]

				if filepath.Ext(markdownFilePath) == ".md" {
					generatedDocPath := GetMiniDocGenDir()
					// does pandoc exist in path?
					if !DoesBinaryExists("pandoc") {
						t.App.SetStatus("[black:red]please install pandoc to generate pdf[white]")
						return nil
					}
					pdfFiePath := generatedDocPath + "/" + filename + ".pdf"
					err := Exec([]string{"pandoc", "-s", markdownFilePath, "-o", pdfFiePath})
					if err != nil {
						t.App.SetStatus("[black:red]generating pdf: " + err.Error() + "[white]")
						return nil
					}
					t.App.SetStatus("[white:darkcyan]generating pdf[white]")
					t.RefreshRootNode()
					t.App.SetFocus(t.Tree)
				}
			}
		case tcell.KeyCtrlD:
			if t.SelectedNode != nil {
				path := t.SelectedNode.Path
				os.Remove(path)
				// parentPath := t.Nodes[GetParentPath(t.SelectedNode.Path)].Path
				t.RefreshRootNode()
				t.App.SetFocus(t.Tree)
			}
		}

		return event
	}
}

func (t *TreePage) RefreshRootNode() {
	root := t.ConvertToTreeNode(t.WalkGenDirAsNode(GetMiniDocGenDir()), true)
	t.Tree.SetRoot(root).
		SetCurrentNode(root).
		SetSelectedFunc(func(n *tview.TreeNode) {
			original := n.GetReference().(*Node)
			if original.Expand {
				n.SetExpanded(!n.IsExpanded())
			} else if original.Selected != nil {
				original.Selected()
			}
			t.SelectedNode = original
		}).
		SetChangedFunc(func(n *tview.TreeNode) {
			original := n.GetReference().(*Node)
			t.SelectedNode = original
			log.Debugf("selected node path: %s", t.SelectedNode.Path)

			path := t.SelectedNode.Path
			t.Detail.Clear()
			if filepath.Ext(path) == ".md" {
				content, _ := ReadFromFile(path)
				fmt.Fprintln(t.Detail, content)
			}
		})
}

func (t *TreePage) RenameFile(app *SimpleApp, path string) {
	base := path[:strings.LastIndex(path, "/")+1]
	file := path[strings.LastIndex(path, "/")+1:]

	form, input, pages := SingleEntryModalForm("Rename File", "Name:", file, 40, 7)

	form.AddButton("Update", func() {
		filename := input.GetText()
		os.Rename(path, base+filename)
		t.RefreshRootNode()
		if err := t.App.SetRoot(t.App.Layout, true).Run(); err != nil {
			panic(err)
		}
	})
	form.AddButton("Cancel", func() {
		if err := t.App.SetRoot(t.App.Layout, true).Run(); err != nil {
			panic(err)
		}
	})

	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}

func SingleEntryModalForm(title, label, value string, width, height int) (*tview.Form, *tview.InputField, *tview.Pages) {
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0, width, 0).
			SetRows(0, height, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}
	background := tview.NewTextView().SetTextColor(tcell.ColorBlue)
	form := tview.NewForm()
	form.AddInputField(label, value, 0, nil, nil)
	form.SetBorderPadding(1, 1, 1, 1)
	form.SetBorder(true)
	form.SetFieldTextColor(tcell.ColorYellow)
	form.SetTitle(title)
	item := form.GetFormItem(0)
	input, _ := item.(*tview.InputField)
	pages := tview.NewPages().
		AddPage("background", background, true, true).
		AddPage("modal", modal(form, width, height), true, true)
	return form, input, pages
}
