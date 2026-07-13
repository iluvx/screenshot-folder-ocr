package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

var usageTemplate = `{{HeadingStyle "Usage:"}}{{if .Runnable}}{{UseLineStyle .UseLine}}{{end}}{{if .HasAvailableSubCommands}} {{ExecStyle .CommandPath}} [COMMAND]{{end}}{{if gt (len .Aliases) 0}}

{{HeadingStyle "Aliases:"}}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{HeadingStyle "Examples:"}}
{{ExampleStyle .Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

{{HeadingStyle "Commands:"}}{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12)}} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12)}} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad (CommandStyle .Name) (sum .NamePadding 12)}} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{HeadingStyle "Options:"}}
{{FlagStyle .LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{HeadingStyle "Global Options:"}}
{{FlagStyle .InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{HeadingStyle "Additional help topics:"}}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad (CommandStyle .CommandPath) (sum .CommandPathPadding 12)}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}{{end}}
`

var rootCmd = &cobra.Command{
	Use:           "screenshot-folder-ocr",
	Short:         "Screenshot folder OCR",
	Long:          `Screenshot folder OCR is a tool that OCRs screenshots of a folder and saves the results to txt files.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Execute() {
	cc.Init(&cc.Config{
		RootCmd:         rootCmd,
		Headings:        cc.Bold + cc.Underline,
		Commands:        cc.Bold,
		Example:         cc.Italic,
		ExecName:        cc.Bold,
		Flags:           cc.Bold,
		NoExtraNewlines: true,
	})

	rootCmd.SetUsageTemplate(usageTemplate)

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(color.HiRedString("error:"), err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
