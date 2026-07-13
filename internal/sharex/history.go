package sharex

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

const ocrTagKey = "ocr"

// DefaultHistoryDB returns the default ShareX History.db path.
func DefaultHistoryDB() string {
	return filepath.Join(DefaultDirectory(), "History.db")
}

// History updates OCR tags in ShareX's History.db.
type History struct {
	db *sql.DB
}

// OpenHistory opens ShareX History.db for tag updates.
func OpenHistory(dbPath string) (*History, error) {
	if dbPath == "" {
		dbPath = DefaultHistoryDB()
	}
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("ShareX history db: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, err
	}
	return &History{db: db}, nil
}

func (h *History) Close() error {
	if h == nil || h.db == nil {
		return nil
	}
	return h.db.Close()
}

// AddOCRTag sets Tags.ocr to the OCR text for the history row matching imagePath.
// Returns false if no matching ShareX history entry was found.
func (h *History) AddOCRTag(imagePath, content string) (bool, error) {
	return h.updateTag(imagePath, content, true)
}

// RemoveOCRTag removes Tags.ocr for the history row matching imagePath.
// Returns false if no matching ShareX history entry was found.
func (h *History) RemoveOCRTag(imagePath string) (bool, error) {
	return h.updateTag(imagePath, "", false)
}

func (h *History) updateTag(imagePath, content string, add bool) (bool, error) {
	abs, err := filepath.Abs(imagePath)
	if err != nil {
		abs = imagePath
	}

	id, tagsJSON, err := h.findEntry(abs)
	if err != nil {
		return false, err
	}
	if id == 0 {
		return false, nil
	}

	tags := map[string]string{}
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			return false, fmt.Errorf("parse tags for id %d: %w", id, err)
		}
	}

	if add {
		removeOCRKeys(tags)
		if tags[ocrTagKey] == content {
			return true, nil
		}
		tags[ocrTagKey] = content
	} else {
		if !removeOCRKeys(tags) {
			return true, nil
		}
	}

	var out any
	if len(tags) == 0 {
		out = nil
	} else {
		b, err := json.Marshal(tags)
		if err != nil {
			return false, err
		}
		out = string(b)
	}

	_, err = h.db.Exec(`UPDATE History SET Tags = ? WHERE Id = ?`, out, id)
	if err != nil {
		return false, err
	}
	return true, nil
}

func removeOCRKeys(tags map[string]string) bool {
	changed := false
	for k := range tags {
		if strings.EqualFold(k, ocrTagKey) {
			delete(tags, k)
			changed = true
		}
	}
	return changed
}

func (h *History) findEntry(absPath string) (id int64, tags string, err error) {
	id, tags, err = h.findByFilePath(absPath)
	if err != nil || id != 0 {
		return id, tags, err
	}
	return h.findByFileName(filepath.Base(absPath))
}

func (h *History) findByFilePath(absPath string) (id int64, tags string, err error) {
	for _, p := range pathCandidates(absPath) {
		var tagsNull sql.NullString
		err = h.db.QueryRow(
			`SELECT Id, Tags FROM History WHERE FilePath = ? COLLATE NOCASE LIMIT 1`,
			p,
		).Scan(&id, &tagsNull)
		if err == sql.ErrNoRows {
			err = nil
			continue
		}
		if err != nil {
			return 0, "", err
		}
		if tagsNull.Valid {
			tags = tagsNull.String
		}
		return id, tags, nil
	}
	return 0, "", nil
}

func (h *History) findByFileName(name string) (id int64, tags string, err error) {
	if name == "" {
		return 0, "", nil
	}
	var tagsNull sql.NullString
	err = h.db.QueryRow(
		`SELECT Id, Tags FROM History WHERE FileName = ? COLLATE NOCASE ORDER BY Id DESC LIMIT 1`,
		name,
	).Scan(&id, &tagsNull)
	if err == sql.ErrNoRows {
		return 0, "", nil
	}
	if err != nil {
		return 0, "", err
	}
	if tagsNull.Valid {
		tags = tagsNull.String
	}
	return id, tags, nil
}

func pathCandidates(absPath string) []string {
	clean := filepath.Clean(absPath)
	slash := filepath.ToSlash(clean)
	back := strings.ReplaceAll(slash, "/", `\`)
	seen := map[string]struct{}{}
	var out []string
	for _, p := range []string{absPath, clean, slash, back} {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
