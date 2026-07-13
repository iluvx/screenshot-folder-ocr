package cmd

import (
	"github.com/spf13/cobra"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/ocrfolder"
)

var (
	cleanRecursive bool
	cleanTarget    string
	cleanShareXDB  string
)

var cleanCmd = &cobra.Command{
	Use:   "clean [path]",
	Short: "Remove OCR data",
	Long: `Remove OCR sidecar txt files, ShareX history tags, or both.
Sidecars are only removed for existing images, so other text files remain untouched.
If path is omitted, the configured ShareX screenshots folder is scanned recursively.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, usedDefault, err := resolveDirectory(args)
		if err != nil {
			return err
		}
		return ocrfolder.Remove(path, ocrfolder.Options{
			Recursive:   cleanRecursive || usedDefault,
			Destination: cleanTarget,
			ShareXDB:    cleanShareXDB,
		})
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
	cleanCmd.Flags().BoolVarP(&cleanRecursive, "recursive", "r", false, "recursively remove OCR data in all subfolders")
	cleanCmd.Flags().StringVarP(&cleanTarget, "target", "t", "txt", "OCR data to remove: txt, sharex, or both")
	cleanCmd.Flags().StringVar(&cleanShareXDB, "sharex-db", "", "path to ShareX History.db")
}
