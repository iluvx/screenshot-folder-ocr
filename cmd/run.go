package cmd

import (
	"github.com/spf13/cobra"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/ocrfolder"
)

var crawl bool

var runCmd = &cobra.Command{
	Use:   "run [path]",
	Short: "OCR screenshots in a folder",
	Long:  `OCR screenshots in the given folder and save the results to txt files.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ocrfolder.Process(args[0], ocrfolder.Options{Crawl: crawl})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVarP(&crawl, "crawl", "c", false, "recursively OCR images in all subfolders")
}
