package prowl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

const (
	maxIndexFileBytes = 512 * 1024 // 512 KiB cap per file to keep indexing bounded
	chunkLines        = 80
	chunkOverlap      = 10
)

var skipDirs = map[string]bool{
	".git":         true,
	".hg":          true,
	".svn":         true,
	"node_modules": true,
	".next":        true,
	"dist":         true,
	"build":        true,
	"target":       true,
	".cache":       true,
	"__pycache__":  true,
}

var symbolLineRe = regexp.MustCompile(`^(?:func|type|class|def|fn|struct|enum|interface|export\s+(?:default\s+)?(?:async\s+)?(?:function|class|const|interface|type)|component|function)\s+([A-Za-z_][A-Za-z0-9_]*)`)

// Index wraps the cited SQLite code index.
type Index struct {
	db   *sql.DB
	root string
}

// IndexStats summarizes the current index state.
type IndexStats struct {
	Root          string `json:"root"`
	Files         int    `json:"files"`
	Chunks        int    `json:"chunks"`
	MaskedSecrets int    `json:"masked_secrets"`
	IndexedAt     int64  `json:"indexed_at,omitempty"`
}

// IndexResult reports what an IndexWorkspace run did.
type IndexResult struct {
	Scanned       int            `json:"scanned"`
	Indexed       int            `json:"indexed"`
	Skipped       int            `json:"skipped"`
	Pruned        int            `json:"pruned"`
	Chunks        int            `json:"chunks"`
	MaskedSecrets int            `json:"masked_secrets"`
	MaskedVendors map[string]int `json:"masked_vendors,omitempty"`
}

// Citation is a single cited code search hit.
type Citation struct {
	Path      string  `json:"path"`
	StartLine int     `json:"start_line"`
	EndLine   int     `json:"end_line"`
	Symbol    string  `json:"symbol,omitempty"`
	Language  string  `json:"language"`
	Snippet   string  `json:"snippet"`
	Score     float64 `json:"score"`
}

// OpenIndex opens (or creates) the SQLite code index at dbPath for root.
func OpenIndex(dbPath, root string) (*Index, error) {
	if dbPath == "" {
		return nil, errors.New("prowl: dbPath is required")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("prowl: create index dir: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("prowl: resolve root: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("prowl: open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := initSchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Index{db: db, root: absRoot}, nil
}

// Close closes the underlying SQLite database.
func (idx *Index) Close() error {
	if idx == nil || idx.db == nil {
		return nil
	}
	return idx.db.Close()
}

func initSchema(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY,
	path TEXT NOT NULL UNIQUE,
	sha256 TEXT NOT NULL,
	size INTEGER NOT NULL,
	mtime_ns INTEGER NOT NULL,
	language TEXT NOT NULL,
	masked_count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS chunks (
	id INTEGER PRIMARY KEY,
	file_id INTEGER NOT NULL REFERENCES files(id) ON DELETE CASCADE,
	path TEXT NOT NULL,
	start_line INTEGER NOT NULL,
	end_line INTEGER NOT NULL,
	symbol TEXT NOT NULL DEFAULT '',
	language TEXT NOT NULL,
	content TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_chunks_file_id ON chunks(file_id);

CREATE VIRTUAL TABLE IF NOT EXISTS chunks_fts USING fts5(
	path,
	symbol,
	content,
	content='chunks',
	content_rowid='id',
	tokenize='porter unicode61'
);

CREATE TRIGGER IF NOT EXISTS chunks_ai AFTER INSERT ON chunks BEGIN
	INSERT INTO chunks_fts(rowid, path, symbol, content)
	VALUES (new.id, new.path, new.symbol, new.content);
END;

CREATE TRIGGER IF NOT EXISTS chunks_ad AFTER DELETE ON chunks BEGIN
	INSERT INTO chunks_fts(chunks_fts, rowid, path, symbol, content)
	VALUES ('delete', old.id, old.path, old.symbol, old.content);
END;

CREATE TRIGGER IF NOT EXISTS chunks_au AFTER UPDATE ON chunks BEGIN
	INSERT INTO chunks_fts(chunks_fts, rowid, path, symbol, content)
	VALUES ('delete', old.id, old.path, old.symbol, old.content);
	INSERT INTO chunks_fts(rowid, path, symbol, content)
	VALUES (new.id, new.path, new.symbol, new.content);
END;
`
	_, err := db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("prowl: init schema: %w", err)
	}
	return nil
}

// IndexWorkspace walks idx.root, masks unambiguous secrets, and updates SQLite.
func (idx *Index) IndexWorkspace(ctx context.Context) (IndexResult, error) {
	var res IndexResult
	res.MaskedVendors = map[string]int{}

	seen := map[string]bool{}
	type pendingFile struct {
		rel     string
		sha     string
		size    int64
		mtimeNs int64
		lang    string
		masked  string
		maskRpt MaskReport
	}

	existing := map[string]string{}
	rows, err := idx.db.QueryContext(ctx, `SELECT path, sha256 FROM files`)
	if err != nil {
		return res, fmt.Errorf("prowl: query existing files: %w", err)
	}
	for rows.Next() {
		var p, h string
		if err := rows.Scan(&p, &h); err != nil {
			_ = rows.Close()
			return res, err
		}
		existing[p] = h
	}
	_ = rows.Close()

	var toUpdate []pendingFile
	err = filepath.WalkDir(idx.root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		name := d.Name()
		if d.IsDir() {
			if skipDirs[name] || (strings.HasPrefix(name, ".") && path != idx.root) {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		lang := detectLanguage(path)
		if lang == "" {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() == 0 || info.Size() > maxIndexFileBytes {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil || isBinary(raw) {
			return nil
		}
		res.Scanned++

		rel, err := filepath.Rel(idx.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		seen[rel] = true

		sum := sha256.Sum256(raw)
		hashHex := hex.EncodeToString(sum[:])
		if prev, ok := existing[rel]; ok && prev == hashHex {
			res.Skipped++
			return nil
		}

		masked, mRpt := MaskSecrets(string(raw))
		toUpdate = append(toUpdate, pendingFile{
			rel:     rel,
			sha:     hashHex,
			size:    info.Size(),
			mtimeNs: info.ModTime().UnixNano(),
			lang:    lang,
			masked:  masked,
			maskRpt: mRpt,
		})
		return nil
	})
	if err != nil {
		return res, err
	}

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return res, err
	}
	defer func() { _ = tx.Rollback() }()

	for oldPath := range existing {
		if !seen[oldPath] {
			if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE path = ?`, oldPath); err != nil {
				return res, err
			}
			res.Pruned++
		}
	}

	for _, f := range toUpdate {
		if _, err := tx.ExecContext(ctx, `DELETE FROM files WHERE path = ?`, f.rel); err != nil {
			return res, err
		}
		insRes, err := tx.ExecContext(ctx,
			`INSERT INTO files(path, sha256, size, mtime_ns, language, masked_count) VALUES (?, ?, ?, ?, ?, ?)`,
			f.rel, f.sha, f.size, f.mtimeNs, f.lang, f.maskRpt.Total,
		)
		if err != nil {
			return res, err
		}
		fileID, err := insRes.LastInsertId()
		if err != nil {
			return res, err
		}

		chunks := chunkFile(f.masked)
		for _, c := range chunks {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO chunks(file_id, path, start_line, end_line, symbol, language, content) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				fileID, f.rel, c.startLine, c.endLine, c.symbol, f.lang, c.content,
			); err != nil {
				return res, err
			}
			res.Chunks++
		}
		res.Indexed++
		res.MaskedSecrets += f.maskRpt.Total
		for v, count := range f.maskRpt.Vendors {
			res.MaskedVendors[v] += count
		}
	}

	if err := tx.Commit(); err != nil {
		return res, err
	}
	if len(res.MaskedVendors) == 0 {
		res.MaskedVendors = nil
	}
	return res, nil
}

// Search queries the cited FTS5 index and returns ranked citations.
func (idx *Index) Search(ctx context.Context, query string, limit int) ([]Citation, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	ftsQuery := buildFTSQuery(query)
	if ftsQuery == "" {
		return nil, nil
	}

	rows, err := idx.db.QueryContext(ctx, `
SELECT c.path, c.start_line, c.end_line, c.symbol, c.language,
       snippet(chunks_fts, 2, '[[', ']]', '...', 24) AS snip,
       bm25(chunks_fts, 2.0, 3.0, 1.0) AS score
FROM chunks_fts
JOIN chunks c ON c.id = chunks_fts.rowid
WHERE chunks_fts MATCH ?
ORDER BY score
LIMIT ?`, ftsQuery, limit)
	if err != nil {
		return nil, fmt.Errorf("prowl: search: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var out []Citation
	for rows.Next() {
		var c Citation
		if err := rows.Scan(&c.Path, &c.StartLine, &c.EndLine, &c.Symbol, &c.Language, &c.Snippet, &c.Score); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ReadRange returns masked lines [startLine..endLine] (1-based, inclusive) for relPath.
func (idx *Index) ReadRange(relPath string, startLine, endLine int) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relPath))
	if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
		return "", errors.New("prowl: path outside workspace")
	}
	full := filepath.Join(idx.root, clean)
	raw, err := os.ReadFile(full)
	if err != nil {
		return "", err
	}
	masked, _ := MaskSecrets(string(raw))
	lines := strings.Split(masked, "\n")
	if startLine < 1 {
		startLine = 1
	}
	if endLine <= 0 || endLine > len(lines) {
		endLine = len(lines)
	}
	if startLine > endLine {
		return "", fmt.Errorf("prowl: invalid line range %d-%d (file has %d lines)", startLine, endLine, len(lines))
	}
	var b strings.Builder
	for i := startLine; i <= endLine; i++ {
		fmt.Fprintf(&b, "%d: %s\n", i, lines[i-1])
	}
	return b.String(), nil
}

// Stats returns counts from the SQLite index.
func (idx *Index) Stats(ctx context.Context) (IndexStats, error) {
	st := IndexStats{Root: idx.root}
	err := idx.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(masked_count), 0) FROM files`,
	).Scan(&st.Files, &st.MaskedSecrets)
	if err != nil {
		return st, err
	}
	err = idx.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chunks`).Scan(&st.Chunks)
	return st, err
}

type textChunk struct {
	startLine int
	endLine   int
	symbol    string
	content   string
}

func chunkFile(text string) []textChunk {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		return nil
	}
	var out []textChunk
	step := chunkLines - chunkOverlap
	if step < 1 {
		step = chunkLines
	}
	for start := 0; start < len(lines); start += step {
		end := start + chunkLines
		if end > len(lines) {
			end = len(lines)
		}
		sym := ""
		for i := start; i < end; i++ {
			trimmed := strings.TrimSpace(lines[i])
			if m := symbolLineRe.FindStringSubmatch(trimmed); len(m) == 2 {
				sym = m[1]
				break
			}
		}
		out = append(out, textChunk{
			startLine: start + 1,
			endLine:   end,
			symbol:    sym,
			content:   strings.Join(lines[start:end], "\n"),
		})
		if end == len(lines) {
			break
		}
	}
	return out
}

func buildFTSQuery(raw string) string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-')
	})
	if len(fields) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(fields))
	for _, f := range fields {
		if f == "" {
			continue
		}
		quoted = append(quoted, `"`+strings.ReplaceAll(f, `"`, `""`)+`"`)
	}
	return strings.Join(quoted, " AND ")
}

func isBinary(b []byte) bool {
	if bytes.IndexByte(b, 0) >= 0 {
		return true
	}
	return !utf8.Valid(b)
}

func detectLanguage(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".qml":
		return "qml"
	case ".sh", ".bash", ".zsh":
		return "shell"
	case ".c", ".h":
		return "c"
	case ".cpp", ".hpp", ".cc":
		return "cpp"
	case ".json", ".yaml", ".yml", ".toml", ".md", ".sql":
		return strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	default:
		return ""
	}
}
