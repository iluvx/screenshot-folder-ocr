package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/ocrfolder"
)

var (
	ocrRecursive bool
	ocrOutput    string
	ocrShareXDB  string
	ocrYes       bool
)

var ocrCmd = &cobra.Command{
	Use:   "ocr [path]",
	Short: "OCR screenshots in a folder",
	Long: `OCR screenshots into txt files, ShareX history tags, or both.
If path is omitted, the configured ShareX screenshots folder is scanned recursively.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, usedDefault, err := resolveDirectory(args)
		if err != nil {
			return err
		}
		recursive := ocrRecursive || usedDefault
		if !ocrYes {
			confirmed, err := confirmOCR(cmd, path, ocrOutput, recursive)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
				return nil
			}
		}
		return ocrfolder.Process(path, ocrfolder.Options{
			Recursive:   recursive,
			Destination: ocrOutput,
			ShareXDB:    ocrShareXDB,
		})
	},
}

func init() {
	rootCmd.AddCommand(ocrCmd)
	ocrCmd.Flags().BoolVarP(&ocrRecursive, "recursive", "r", false, "recursively OCR images in all subfolders")
	ocrCmd.Flags().StringVarP(&ocrOutput, "output", "o", "txt", "OCR output: txt, sharex, or both")
	ocrCmd.Flags().StringVar(&ocrShareXDB, "sharex-db", "", "path to ShareX History.db")
	ocrCmd.Flags().BoolVarP(&ocrYes, "yes", "y", false, "skip the confirmation prompt")
}

func confirmOCR(cmd *cobra.Command, path, output string, recursive bool) (bool, error) {
	fmt.Fprintf(
		cmd.OutOrStdout(),
		"OCR images in %q (output: %s, recursive: %t)? [y/N]: ",
		path,
		output,
		recursive,
	)

	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		return false, scanner.Err()
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}
