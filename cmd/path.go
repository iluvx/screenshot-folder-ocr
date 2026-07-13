package cmd

import "github.cmo/iluvx/screenshot-folder-ocr/internal/sharex"

func resolveDirectory(args []string) (path string, usedDefault bool, err error) {
	if len(args) > 0 {
		return args[0], false, nil
	}

	path, err = sharex.DefaultScreenshotsPath()
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}
