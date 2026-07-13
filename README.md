# screenshot-folder-ocr

OCR a folder of screenshots so you can search them later.

Saves text as sidecar files (`photo.png.txt`), ShareX history tags, or both. Images that already have OCR data are skipped.

![Showcase](showcase.png)

## Requirements

- [Go](https://go.dev/dl/)
- [Tesseract OCR](https://github.com/UB-Mannheim/tesseract/wiki) on your `PATH`
- Optional: [ShareX](https://getsharex.com/) for history tagging

## Install

```bash
git clone <repo-url>
cd screenshot-folder-ocr
go build -o screenshot-folder-ocr .
```

## Quick start

```bash
# OCR a folder into .txt sidecars
screenshot-folder-ocr ocr ./screenshots

# OCR ShareX's screenshots folder (path omitted = ShareX default, recursive)
screenshot-folder-ocr ocr --output both

# Remove OCR data
screenshot-folder-ocr clean ./screenshots
```

`ocr` always asks for confirmation first. Pass `-y` / `--yes` to skip that in scripts.

## Commands

### `ocr [path]`

Runs Tesseract on images in a folder.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | `txt` | Where to store OCR text: `txt`, `sharex`, or `both` |
| `--recursive` | `-r` | off* | Include subfolders |
| `--yes` | `-y` | off | Skip confirmation |
| `--sharex-db` | | ShareX default | Path to `History.db` |

\* If `path` is omitted, the ShareX screenshots directory is used and recursion is enabled automatically.

```bash
screenshot-folder-ocr ocr ./screenshots
screenshot-folder-ocr ocr ./screenshots -o both -r
screenshot-folder-ocr ocr -o sharex -y
```

- `txt` writes `image.png.txt` next to each image
- `sharex` sets the `ocr` tag in ShareX history to the OCR text
- `both` does both

### `clean [path]`

Removes OCR data.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--target` | `-t` | `txt` | What to remove: `txt`, `sharex`, or `both` |
| `--recursive` | `-r` | off* | Include subfolders |
| `--sharex-db` | | ShareX default | Path to `History.db` |

```bash
screenshot-folder-ocr clean ./screenshots
screenshot-folder-ocr clean ./screenshots -t both -r
screenshot-folder-ocr clean -t sharex
```

- `txt` deletes sidecar files only when the matching image still exists
- `sharex` removes the `ocr` tag from ShareX history
- `both` does both (including ShareX tags when no txt file is present)

## ShareX defaults

When you omit `path`, the tool reads ShareX's config:

1. `%USERPROFILE%\Documents\ShareX\ApplicationConfig.json`
2. Uses `CustomScreenshotsPath` when custom screenshots path is enabled
3. Otherwise falls back to `%USERPROFILE%\Documents\ShareX\Screenshots`

History tagging uses `%USERPROFILE%\Documents\ShareX\History.db` unless you pass `--sharex-db`.

---

<sup>Kinda stole this idea from [labtec901/Auto-OCR-Screenshot-Directory](https://github.com/labtec901/Auto-OCR-Screenshot-Directory).</sup>
