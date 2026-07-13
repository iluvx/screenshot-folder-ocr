package cmd

import (
	"github.com/spf13/cobra"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/ocrfolder"
)

var cleanCrawl bool

var cleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Remove OCR txt files",
	Long: `Remove OCR sidecar txt files that match an existing image
(e.g. photo.png.txt is removed only if photo.png exists).
Other text files are left untouched.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return ocrfolder.Remove(args[0], ocrfolder.Options{Crawl: cleanCrawl})
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanCrawl, "crawl", "c", false, "recursively remove OCR files in all subfolders")
}
