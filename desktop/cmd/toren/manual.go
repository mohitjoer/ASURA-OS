package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// `toren manual` shows the user manual (docs/manual in the repository,
// installed at /usr/share/doc/toren/manual): the list of chapters, one
// chapter, or all of them, through a pager.

var (
	manualDir   = "/usr/share/doc/toren/manual"
	manualPager = func(text string) error {
		if !isTerminal(os.Stdout) {
			_, err := fmt.Print(text)
			return err
		}
		pager := "less"
		args := []string{"-R"}
		if p, err := exec.LookPath("glow"); err == nil {
			pager, args = p, []string{"-p", "-"}
		}
		cmd := exec.Command(pager, args...)
		cmd.Stdin = strings.NewReader(text)
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			_, err = fmt.Print(text)
			return err
		}
		return nil
	}
)

const manualUsage = `Usage: toren manual [CHAPTER | --all]

  (none)    the chapters
  CHAPTER   one chapter, by name or number (toren manual updates, toren manual 4)
  --all     every chapter, one after the other
`

func init() {
	register("manual", "read the ASURA user manual", torenManual)
}

type manualChapter struct {
	Num   string // "04"
	Name  string // "updates"
	Title string // the first heading
	Path  string
}

func manualChapters(dir string) ([]manualChapter, error) {
	files, err := filepath.Glob(filepath.Join(dir, "[0-9][0-9]-*.md"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	var out []manualChapter
	for _, f := range files {
		base := strings.TrimSuffix(filepath.Base(f), ".md")
		num, name, _ := strings.Cut(base, "-")
		out = append(out, manualChapter{Num: num, Name: name, Title: firstHeading(f), Path: f})
	}
	return out, nil
}

func firstHeading(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		if t, ok := strings.CutPrefix(s.Text(), "# "); ok {
			return t
		}
	}
	if err := s.Err(); err != nil {
		return ""
	}
	return ""
}

func torenManual(args []string) int {
	chapters, err := manualChapters(manualDir)
	if err != nil || len(chapters) == 0 {
		fmt.Fprintf(os.Stderr, "toren manual: the manual is not installed (%s)\n", manualDir)
		return exitFail
	}
	if len(args) == 0 {
		fmt.Println("Toren Linux manual. Read a chapter with toren manual NAME:")
		for _, c := range chapters {
			fmt.Printf("  %s  %-10s %s\n", strings.TrimLeft(c.Num, "0"), c.Name, c.Title)
		}
		return exitOK
	}
	if len(args) > 1 {
		fmt.Fprint(os.Stderr, manualUsage)
		return exitUsage
	}
	var pick []manualChapter
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(manualUsage)
		return exitOK
	case "--all":
		pick = chapters
	default:
		want := strings.ToLower(args[0])
		for _, c := range chapters {
			if c.Name == want || c.Num == want || strings.TrimLeft(c.Num, "0") == want {
				pick = append(pick, c)
			}
		}
	}
	if len(pick) == 0 {
		fmt.Fprintf(os.Stderr, "toren manual: no chapter %q (toren manual lists them)\n", args[0])
		return exitUsage
	}
	var text strings.Builder
	for i, c := range pick {
		b, err := os.ReadFile(c.Path)
		if err != nil {
			fmt.Fprintln(os.Stderr, "toren manual:", err)
			return exitFail
		}
		if i > 0 {
			text.WriteString("\n---\n\n")
		}
		text.Write(b)
	}
	if err := manualPager(text.String()); err != nil {
		fmt.Fprintln(os.Stderr, "toren manual:", err)
		return exitFail
	}
	return exitOK
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
