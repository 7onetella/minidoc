package cmd

import (
	"fmt"
	//"github.com/lacion/cookiecutter_golang_example/log"
	"os"

	"github.com/7onetella/minidoc"
	l "github.com/7onetella/minidoc/log"
	"github.com/gdamore/tcell"
	"github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
)

var log = l.GetNewLogrusLogger()

var cfgFile string

var devMode bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generated code example",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	 Run: func(cmd *cobra.Command, args []string) {
		 log.Debug("root.Execute")
		 launchMinidoc()
	 },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	flags := rootCmd.Flags()
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	flags.BoolVar(&devMode, "dev",false, "development mode")

}

func launchMinidoc() {

	homedir, _ := homedir.Dir() // return path with slash at the end
	minidocHome := homedir + "/.minidoc"

	if devMode {
		minidocHome = ".minidoc"
	}
	os.MkdirAll(minidocHome, os.ModePerm)

	pageItems := []minidoc.PageItem{minidoc.NewSearch(), minidoc.NewHelp()}
	options := []minidoc.SimpleAppOption{
		GetWithSimpleAppDelegateKeyEvent(),
		minidoc.WithSimpleAppConfirmExit(false),
		minidoc.WithSimpleAppPages(pageItems),
		minidoc.WithSimpleAppDataFolderPath(minidocHome),
	}
	if devMode {
		options = append(options, minidoc.WithSimpleAppDebugOn())
	}

	app := minidoc.NewSimpleApp(options...)

	defer l.Logfile.Close()

	if err := app.SetRoot(app.Layout, true).Run(); err != nil {
		panic(err)
	}
}

func GetWithSimpleAppDelegateKeyEvent() minidoc.SimpleAppOption {
	return minidoc.WithSimpleAppDelegateKeyEvent(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlA:
			os.Exit(1)
		// do
		default:
			// do nothing
		}
		return event
	})
}
