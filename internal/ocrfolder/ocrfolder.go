package ocrfolder

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"

	"github.cmo/iluvx/screenshot-folder-ocr/internal/sharex"
)

var imageExts = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".bmp":  {},
	".gif":  {},
	".tif":  {},
	".tiff": {},
	".webp": {},
}

// Options controls how directory commands walk folders.
type Options struct {
	Recursive   bool
	Destination string
	ShareXDB    string
}

const (
	destinationTXT    = "txt"
	destinationShareX = "sharex"
	destinationBoth   = "both"
)

func (o Options) destinations() (txt, shareX bool, err error) {
	switch strings.ToLower(o.Destination) {
	case "", destinationTXT:
		return true, false, nil
	case destinationShareX:
		return false, true, nil
	case destinationBoth:
		return true, true, nil
	default:
		return false, false, fmt.Errorf("invalid destination %q: use txt, sharex, or both", o.Destination)
	}
}

// Process OCRs every image in dir, writing "<image>.txt" beside each file.
// Images that already have a matching .txt are skipped.
// When opts.Recursive is true, subdirectories are walked recursively.
func Process(dir string, opts Options) error {
	writeTXT, writeShareX, err := opts.destinations()
	if err != nil {
		return err
	}

	if err := requireDir(dir); err != nil {
		return err
	}

	if _, err := exec.LookPath("tesseract"); err != nil {
		return fmt.Errorf("tesseract not found in PATH: install Tesseract OCR and ensure `tesseract` is available")
	}

	var hist *sharex.History
	if writeShareX {
		var err error
		hist, err = sharex.OpenHistory(opts.ShareXDB)
		if err != nil {
			return err
		}
		defer hist.Close()
	}

	var n int
	err = walkFiles(dir, opts.Recursive, func(path string) error {
		ok, err := processImage(path, hist, writeTXT)
		if ok {
			n++
		}
		return err
	})
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println(color.GreenString("OCR'd %d file(s)", n))
	return nil
}

// Remove deletes OCR sidecar txt files that match an existing image
// (e.g. photo.png.txt only if photo.png exists). Other .txt files are left alone.
func Remove(dir string, opts Options) error {
	removeTXT, removeShareX, err := opts.destinations()
	if err != nil {
		return err
	}

	if err := requireDir(dir); err != nil {
		return err
	}

	var hist *sharex.History
	if removeShareX {
		var err error
		hist, err = sharex.OpenHistory(opts.ShareXDB)
		if err != nil {
			return err
		}
		defer hist.Close()
	}

	var n int
	err = walkFiles(dir, opts.Recursive, func(path string) error {
		ok, err := removeOCRData(path, hist, removeTXT)
		if ok {
			n++
		}
		return err
	})
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println(color.GreenString("cleared %d file(s)", n))
	return nil
}

func requireDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}
	return nil
}

func walkFiles(dir string, crawl bool, fn func(path string) error) error {
	if crawl {
		return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			return fn(path)
		})
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := fn(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func processImage(imagePath string, hist *sharex.History, writeTXT bool) (bool, error) {
	if !isImagePath(imagePath) {
		return false, nil
	}

	txtPath := imagePath + ".txt"
	var text []byte
	txtExists := false
	if data, err := os.ReadFile(txtPath); err == nil {
		txtExists = true
		text = data
	} else if !os.IsNotExist(err) {
		if writeTXT || hist != nil {
			return false, fmt.Errorf("check %s: %w", txtPath, err)
		}
	}

	sharexTagged := false
	if hist != nil {
		_, hasTag, err := hist.HasOCRTag(imagePath)
		if err != nil {
			fmt.Println(color.YellowString("fail  %s — could not read ShareX tag", imagePath))
			return false, err
		}
		sharexTagged = hasTag
	}

	needTXT := writeTXT && !txtExists
	needShareX := hist != nil && !sharexTagged
	if !needTXT && !needShareX {
		switch {
		case writeTXT && hist != nil:
			fmt.Println(color.YellowString("skip  %s — OCR txt and ShareX tag already exist", imagePath))
		case hist != nil:
			fmt.Println(color.YellowString("skip  %s — ShareX OCR tag already exists", imagePath))
		default:
			fmt.Println(color.YellowString("skip  %s — OCR txt already exists", imagePath))
		}
		return false, nil
	}

	if needTXT || (needShareX && !txtExists) {
		var err error
		text, err = ocrImage(imagePath)
		if err != nil {
			fmt.Println(color.YellowString("fail  %s — OCR failed", imagePath))
			return false, fmt.Errorf("ocr %s: %w", filepath.Base(imagePath), err)
		}
	}

	txtWritten := false
	if needTXT {
		if err := os.WriteFile(txtPath, text, 0o644); err != nil {
			fmt.Println(color.YellowString("fail  %s — could not write OCR txt", imagePath))
			return false, fmt.Errorf("write %s: %w", txtPath, err)
		}
		txtWritten = true
	}

	if needShareX {
		found, err := updateShareXTag(hist, imagePath, string(text), true)
		if err != nil {
			fmt.Println(color.YellowString("fail  %s — could not update ShareX tag", imagePath))
			return txtWritten, err
		}
		if !found {
			if txtWritten {
				fmt.Println(color.YellowString("partial  %s — wrote OCR txt; ShareX history entry not found", imagePath))
				return true, nil
			}
			if writeTXT && txtExists {
				fmt.Println(color.YellowString("skip  %s — OCR txt exists; ShareX history entry not found", imagePath))
			} else {
				fmt.Println(color.YellowString("skip  %s — ShareX history entry not found", imagePath))
			}
			return false, nil
		}

		switch {
		case txtWritten:
			fmt.Println(color.GreenString("ocr   %s — wrote OCR txt and added OCR text to ShareX tag", imagePath))
		case writeTXT && txtExists:
			fmt.Println(color.GreenString("tag   %s — added OCR text to ShareX tag; txt already exists", imagePath))
		case txtExists:
			fmt.Println(color.GreenString("tag   %s — added existing OCR text to ShareX tag", imagePath))
		default:
			fmt.Println(color.GreenString("ocr   %s — added OCR text to ShareX tag", imagePath))
		}
		return true, nil
	}

	fmt.Println(color.GreenString("ocr   %s — wrote OCR txt", imagePath))
	return true, nil
}

func removeOCRData(imagePath string, hist *sharex.History, removeTXT bool) (bool, error) {
	if !isImagePath(imagePath) {
		return false, nil
	}

	txtRemoved := false
	if removeTXT {
		txtPath := imagePath + ".txt"
		if err := os.Remove(txtPath); err == nil {
			txtRemoved = true
		} else if !os.IsNotExist(err) {
			fmt.Println(color.YellowString("fail  %s — could not remove OCR txt", imagePath))
			return false, fmt.Errorf("remove %s: %w", txtPath, err)
		}
	}

	if hist != nil {
		found, err := updateShareXTag(hist, imagePath, "", false)
		if err != nil {
			if txtRemoved {
				fmt.Println(color.YellowString("partial  %s — removed OCR txt; could not update ShareX tag", imagePath))
			} else {
				fmt.Println(color.YellowString("fail  %s — could not update ShareX tag", imagePath))
			}
			return txtRemoved, err
		}
		if !found {
			if txtRemoved {
				fmt.Println(color.YellowString("partial  %s — removed OCR txt; ShareX history entry not found", imagePath))
				return true, nil
			}
			if removeTXT {
				fmt.Println(color.YellowString("skip  %s — OCR txt and ShareX history entry not found", imagePath))
			} else {
				fmt.Println(color.YellowString("skip  %s — ShareX history entry not found", imagePath))
			}
			return false, nil
		}

		if txtRemoved {
			fmt.Println(color.GreenString("clean  %s — removed OCR txt and ShareX OCR tag", imagePath))
		} else if removeTXT {
			fmt.Println(color.GreenString("clean  %s — removed ShareX OCR tag; txt not found", imagePath))
		} else {
			fmt.Println(color.GreenString("clean  %s — removed ShareX OCR tag", imagePath))
		}
		return true, nil
	}

	if txtRemoved {
		fmt.Println(color.GreenString("clean  %s — removed OCR txt", imagePath))
		return true, nil
	}
	fmt.Println(color.YellowString("skip  %s — OCR txt not found", imagePath))
	return false, nil
}

func updateShareXTag(hist *sharex.History, imagePath, content string, add bool) (bool, error) {
	var (
		found bool
		err   error
	)
	if add {
		found, err = hist.AddOCRTag(imagePath, content)
	} else {
		found, err = hist.RemoveOCRTag(imagePath)
	}
	if err != nil {
		return false, fmt.Errorf("sharex tag %s: %w", imagePath, err)
	}
	return found, nil
}

func isImagePath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := imageExts[ext]
	return ok
}

func ocrImage(imagePath string) ([]byte, error) {
	cmd := exec.Command("tesseract", imagePath, "stdout")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return out, nil
}
