package cmd

import (
	"github.com/spf13/cobra"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/ocrfolder"
)

var ocrCmd = &cobra.Command{
	Use:   "ocr [path]",
	Short: "OCR screenshots in a folder",
	Long:  `OCR screenshots in the given folder and save the results to txt files.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ocrfolder.Process(args[0], ocrfolder.Options{Crawl: crawl})
	},
}

func init() {
	rootCmd.AddCommand(ocrCmd)
	ocrCmd.Flags().BoolVarP(&crawl, "crawl", "c", false, "recursively OCR images in all subfolders")
}
