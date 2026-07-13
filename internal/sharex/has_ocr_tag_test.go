package sharex

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestHasOCRTag(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "History.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
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
 ('tagged.png', 'C:\shots\tagged.png', '2026-01-01T00:00:00Z', 'Image', '{"ocr":"hello"}'),
 ('empty.png', 'C:\shots\empty.png', '2026-01-01T00:00:00Z', 'Image', '{"ProcessName":"x"}'),
 ('legacy.png', 'C:\shots\legacy.png', '2026-01-01T00:00:00Z', 'Image', '{"OCR":"old"}');
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

	found, has, err := h.HasOCRTag(filepath.Join(dir, "tagged.png"))
	if err != nil || !found || !has {
		t.Fatalf("tagged: found=%v has=%v err=%v", found, has, err)
	}
	found, has, err = h.HasOCRTag(filepath.Join(dir, "empty.png"))
	if err != nil || !found || has {
		t.Fatalf("empty: found=%v has=%v err=%v", found, has, err)
	}
	found, has, err = h.HasOCRTag(filepath.Join(dir, "legacy.png"))
	if err != nil || !found || !has {
		t.Fatalf("legacy: found=%v has=%v err=%v", found, has, err)
	}
	found, has, err = h.HasOCRTag(filepath.Join(dir, "missing.png"))
	if err != nil || found || has {
		t.Fatalf("missing: found=%v has=%v err=%v", found, has, err)
	}
}
