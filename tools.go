package minidoc

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	xmlpath "gopkg.in/xmlpath.v2"
)

func OpenFileIfNoneExist(filepath, content string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		log.Errorf("creating file: %v", err)
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, content)
	if err != nil {
		log.Errorf("writing content: %v", err)
		return err
	}
	return nil
}

func DoesBinaryExists(binary string) bool {
	binary, lookErr := exec.LookPath("vim")
	if lookErr != nil {
		return false
	}
	return true
}

func WriteToFile(filepath, content string) (done bool) {
	file, err := os.Create(filepath)
	if err != nil {
		log.Errorf("creating file: %v", err)
		return true
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, content)
	if err != nil {
		log.Errorf("writing content: %v", err)
		return true
	}
	return false
}

func ReadFromFile(filepath string) (string, error) {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("reading content: %v", err)
		return "", err
	}

	return string(dat), nil
}

func DeleteFile(filepath string) (success bool) {
	err := os.Remove(filepath)
	if err != nil {
		log.Errorf("creating file: %v", err)
		return false
	}
	return true
}

// this works perfectly
func OpenVim(app *SimpleApp, filepath string) {
	app.Suspend(func() {
		cmd := exec.Command("vim", filepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Errorf("opening vi: %v", err)
		}
		log.Debug("returning the control back")
	})
}

// works create, all key inputs works, exit the minidoc since this seems to be replace the process
func (s *Search) openVim(filepath string) {
	binary, lookErr := exec.LookPath("vim")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"vim", filepath}

	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
	s.App.Draw()
}

func (s *Search) openVimForkExec(filepath string) {
	cmd := "vim"
	binary, lookErr := exec.LookPath(cmd)
	if lookErr != nil {
		panic(lookErr)
	}
	//fmt.Println(binary)

	os.Remove("/tmp/stdin")
	os.Remove("/tmp/stdout")
	os.Remove("/tmp/stderr")

	fstdin, err1 := os.Create("/tmp/stdin")
	fstdout, err2 := os.Create("/tmp/stdout")
	fstderr, err3 := os.Create("/tmp/stderr")
	if err1 != nil || err2 != nil || err3 != nil {
		log.Errorf("%v %v %v", err1, err2, err3)
		panic("WOW")
	}

	env := os.Environ()

	argv := []string{filepath}
	procAttr := syscall.ProcAttr{
		Dir:   "/tmp",
		Files: []uintptr{fstdin.Fd(), fstdout.Fd(), fstderr.Fd()},
		Env:   env,
		Sys: &syscall.SysProcAttr{
			Foreground: false,
		},
	}

	pid, err := syscall.ForkExec(binary, argv, &procAttr)
	log.Debugf("pid=%d err=%v", pid, err)
	s.App.Draw()
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

func HTTPGet(url string) ([]byte, error) {
	client := http.Client{
		Timeout: time.Second * 3,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("downloading status code: %d", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

// ScreenScrape hits the given URL and screen scrape  then return dom like object for searching
func ScreenScrape(url string) (*xmlpath.Node, error) {

	client := http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed")
	}

	pageContent, err := ioutil.ReadAll(resp.Body)

	reader := strings.NewReader(string(pageContent))
	root, err := html.Parse(reader)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	html.Render(&b, root)
	fixedHTML := b.String()

	reader = strings.NewReader(fixedHTML)
	xmlroot, xmlerr := xmlpath.ParseHTML(reader)

	if xmlerr != nil {
		log.Fatal(xmlerr)
	}

	return xmlroot, nil
}

// SearchByXPath will walk down the node and children using xpath expression
func SearchByXPath(context *xmlpath.Node, xpath string) []*xmlpath.Node {
	path := xmlpath.MustCompile(xpath)

	nodes := make([]*xmlpath.Node, 0, 100)

	iter := path.Iter(context)
	for iter.Next() {
		nodes = append(nodes, iter.Node())
	}

	return nodes
}

// XPathGet xpath get by index
func XPathGet(context *xmlpath.Node, xpath string, index int) string {
	nodes := SearchByXPath(context, xpath)
	if index >= len(nodes) {
		fmt.Println("failed to get ", xpath, " index:", index)
		return ""
	}
	return strings.TrimSpace(nodes[index].String())
}
