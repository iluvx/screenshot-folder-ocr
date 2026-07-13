package sharex

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveOCRTagRemovesBothCasings(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "History.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
CREATE TABLE History (
    Id INTEGER PRIMARY KEY AUTOINCREMENT,
    FileName TEXT,
    FilePath TEXT,
    DateTime TEXT,
    Type TEXT,
    Host TEXT,
    URL TEXT,
    ThumbnailURL TEXT,
    DeletionURL TEXT,
    ShortenedURL TEXT,
    Tags TEXT
);
INSERT INTO History (FileName, FilePath, DateTime, Type, Tags) VALUES
 ('a.png', 'C:\shots\a.png', '2026-01-01T00:00:00Z', 'Image', '{"OCR":"true","ProcessName":"x"}'),
 ('b.png', 'C:\shots\b.png', '2026-01-01T00:00:00Z', 'Image', '{"ocr":"hello","ProcessName":"y"}');
`)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	h, err := OpenHistory(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	for _, name := range []string{"a.png", "b.png"} {
		ok, err := h.RemoveOCRTag(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("remove %s: %v", name, err)
		}
		if !ok {
			t.Fatalf("expected to find %s", name)
		}
	}

	rows, err := h.db.Query(`SELECT FileName, Tags FROM History ORDER BY FileName`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, tagsJSON string
		var tagsNull sql.NullString
		if err := rows.Scan(&name, &tagsNull); err != nil {
			t.Fatal(err)
		}
		if tagsNull.Valid {
			tagsJSON = tagsNull.String
		}
		tags := map[string]string{}
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
				t.Fatal(err)
			}
		}
		for k := range tags {
			if strings.EqualFold(k, "ocr") {
				t.Fatalf("%s still has OCR tag key %q in %s", name, k, tagsJSON)
			}
		}
		if _, ok := tags["ProcessName"]; !ok {
			t.Fatalf("%s lost ProcessName: %s", name, tagsJSON)
		}
	}
}
