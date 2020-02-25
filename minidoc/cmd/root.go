package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"os"

	"github.com/7onetella/minidoc"
	"github.com/gdamore/tcell"
	"github.com/spf13/cobra"
)

var DevMode bool

var reindex bool

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
	flags.BoolVar(&DevMode, "dev",false, "development mode")

	flags.BoolVar(&reindex, "reindex",false, "reindex docs")

}

func launchMinidoc() {

	minidocHome := CreateMinidocHomeIfNotFound(DevMode)

	pageItems := []minidoc.PageItem{minidoc.NewSearch(), minidoc.NewHelp()}
	options := []minidoc.SimpleAppOption{
		GetWithSimpleAppDelegateKeyEvent(),
		minidoc.WithSimpleAppConfirmExit(false),
		minidoc.WithSimpleAppPages(pageItems),
		minidoc.WithSimpleAppDataFolderPath(minidocHome),
		minidoc.WithSimpleAppDocsReindexed(reindex),
	}
	if DevMode {
		options = append(options, minidoc.WithSimpleAppDebugOn())
	}

	app := minidoc.NewSimpleApp(options...)

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

func CreateMinidocHomeIfNotFound(devMode bool) string {
	homedir, _ := homedir.Dir() // return path with slash at the end
	minidocHome := homedir + "/.minidoc"

	if devMode {
		minidocHome = ".minidoc"
	}
	os.MkdirAll(minidocHome, os.ModePerm)
	return minidocHome
}