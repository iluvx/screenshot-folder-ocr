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
	Crawl bool
}

// Process OCRs every image in dir, writing "<image>.txt" beside each file.
// Images that already have a matching .txt are skipped.
// When opts.Crawl is true, subdirectories are walked recursively.
func Process(dir string, opts Options) error {
	if err := requireDir(dir); err != nil {
		return err
	}

	if _, err := exec.LookPath("tesseract"); err != nil {
		return fmt.Errorf("tesseract not found in PATH: install Tesseract OCR and ensure `tesseract` is available")
	}

	var n int
	err := walkFiles(dir, opts.Crawl, func(path string) error {
		ok, err := processImage(path)
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
	if err := requireDir(dir); err != nil {
		return err
	}

	var n int
	err := walkFiles(dir, opts.Crawl, func(path string) error {
		ok, err := removeOCRFile(path)
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

func processImage(imagePath string) (bool, error) {
	if !isImagePath(imagePath) {
		return false, nil
	}

	txtPath := imagePath + ".txt"
	if _, err := os.Stat(txtPath); err == nil {
		fmt.Println(color.YellowString("skip %s (already OCR'd)", imagePath))
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("check %s: %w", txtPath, err)
	}

	text, err := ocrImage(imagePath)
	if err != nil {
		fmt.Println(color.YellowString("fail %s", imagePath))
		return false, fmt.Errorf("ocr %s: %w", filepath.Base(imagePath), err)
	}

	if err := os.WriteFile(txtPath, text, 0o644); err != nil {
		fmt.Println(color.YellowString("fail %s", imagePath))
		return false, fmt.Errorf("write %s: %w", txtPath, err)
	}
	fmt.Println(color.GreenString("ocr  %s", imagePath))
	return true, nil
}

func removeOCRFile(path string) (bool, error) {
	imagePath, ok := ocrSidecarImage(path)
	if !ok {
		return false, nil
	}

	if _, err := os.Stat(imagePath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("check %s: %w", imagePath, err)
	}

	if err := os.Remove(path); err != nil {
		fmt.Println(color.YellowString("fail %s", path))
		return false, fmt.Errorf("remove %s: %w", path, err)
	}
	fmt.Println(color.GreenString("rm   %s", path))
	return true, nil
}

// ocrSidecarImage returns the image path for an OCR txt sidecar, e.g.
// "photo.png.txt" -> "photo.png". ok is false if the name is not an OCR sidecar.
func ocrSidecarImage(path string) (string, bool) {
	if strings.ToLower(filepath.Ext(path)) != ".txt" {
		return "", false
	}
	imagePath := strings.TrimSuffix(path, filepath.Ext(path))
	if !isImagePath(imagePath) {
		return "", false
	}
	return imagePath, true
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
