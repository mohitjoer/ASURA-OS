package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"toren/desktop/internal/overlay"
)

// torenReset drops user edits or resets the desktop configuration back to ASURA defaults.
func torenReset(args []string) int {
	set := newFlags("reset", resetUsage)
	var (
		root    = set.String("root", "", "operate on a tree under this directory")
		desktop = set.Bool("desktop", false, "reset desktop configuration back to ASURA defaults")
		restore = set.String("restore", "", "restore a previous desktop configuration backup")
	)
	positional, err := parseFlags(set, args)
	if err != nil {
		return exitUsage
	}

	paths, err := overlayPaths(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "toren reset:", err)
		return exitFail
	}

	// 1. Restore a previous backup
	if *restore != "" {
		if *desktop || len(positional) > 0 {
			fmt.Fprintln(os.Stderr, "toren reset: --restore cannot be combined with other arguments")
			return exitUsage
		}
		return restoreDesktopBackup(*restore, *root, paths)
	}

	// 2. Reset desktop configuration to defaults
	if *desktop {
		if len(positional) > 0 {
			fmt.Fprintln(os.Stderr, "toren reset: --desktop takes no positional path")
			return exitUsage
		}
		return resetDesktopConfig(*root, paths)
	}

	// 3. Reset one overlay file back to base
	if len(positional) != 1 {
		fmt.Fprintln(os.Stderr, "toren reset: specify a file path, --desktop, or --restore BACKUP")
		return exitUsage
	}

	if err := overlay.Reset(paths, positional[0]); err != nil {
		fmt.Fprintln(os.Stderr, "toren reset:", err)
		return exitFail
	}
	fmt.Printf("%s is back to the base\n", positional[0])
	return exitOK
}

// resolveUserHome returns the target user home directory.
func resolveUserHome(root string) string {
	if root != "" {
		return root
	}
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	h, err := os.UserHomeDir()
	if err == nil && h != "" {
		return h
	}
	return "/"
}

// resetDesktopConfig backs up user desktop configuration and restores packaged defaults.
func resetDesktopConfig(root string, paths overlay.Paths) int {
	home := resolveUserHome(root)
	stateDir := filepath.Join(home, ".local", "state", "toren")
	backupsBase := filepath.Join(stateDir, "backups")
	stamp := time.Now().UTC().Format("20060102-150405")
	backupDir := filepath.Join(backupsBase, "desktop-"+stamp)

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "toren reset: failed to create backup dir: %v\n", err)
		return exitFail
	}

	// Managed desktop targets relative to home
	targets := []string{
		filepath.Join(".config", "illogical-impulse"),
		filepath.Join(".config", "toren", "hyprland"),
		filepath.Join(".config", "toren", "user_edits"),
		filepath.Join(".local", "state", "toren"),
		filepath.Join(".local", "state", "quickshell"),
	}

	for _, rel := range targets {
		src := filepath.Join(home, rel)
		if _, err := os.Stat(src); err != nil {
			continue
		}
		dst := filepath.Join(backupDir, rel)
		// When backing up ~/.local/state/toren, skip the backups folder itself
		skip := ""
		if rel == filepath.Join(".local", "state", "toren") {
			skip = backupsBase
		}
		if err := copyTree(src, dst, skip); err != nil {
			fmt.Fprintf(os.Stderr, "toren reset: failed to backup %s: %v\n", rel, err)
			return exitFail
		}
	}

	// Clean out old user desktop settings and state
	_ = os.RemoveAll(filepath.Join(home, ".config", "illogical-impulse"))
	_ = os.RemoveAll(filepath.Join(home, ".config", "toren", "hyprland"))
	_ = os.RemoveAll(filepath.Join(home, ".config", "toren", "user_edits"))
	_ = os.RemoveAll(filepath.Join(home, ".local", "state", "quickshell"))

	// Clean entries in stateDir preserving the backups directory
	if entries, err := os.ReadDir(stateDir); err == nil {
		for _, e := range entries {
			if e.Name() == "backups" {
				continue
			}
			_ = os.RemoveAll(filepath.Join(stateDir, e.Name()))
		}
	}

	// Re-materialize the base overlay so default files exist
	if err := overlay.Materialize(paths); err != nil {
		// Non-fatal if base doesn't exist (e.g. running outside full system install)
		_ = err
	}

	fmt.Printf("desktop configuration backed up to %s\n", backupDir)
	fmt.Println("desktop configuration reset to ASURA defaults")
	return exitOK
}

// restoreDesktopBackup restores a previous desktop configuration backup byte-for-byte.
func restoreDesktopBackup(backupDir, root string, paths overlay.Paths) int {
	info, err := os.Stat(backupDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "toren reset: backup directory not found: %s\n", backupDir)
		return exitFail
	}

	home := resolveUserHome(root)

	// Walk backupDir and restore all files to home
	err = filepath.Walk(backupDir, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(backupDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(home, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode())
		}
		return copyFile(path, target)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "toren reset: restore failed: %v\n", err)
		return exitFail
	}

	// Re-materialize base overlay
	if err := overlay.Materialize(paths); err != nil {
		_ = err
	}

	fmt.Printf("restored desktop configuration from %s\n", backupDir)
	return exitOK
}

// copyFile copies a single file from src to dst preserving permissions.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	info, err := s.Stat()
	if err != nil {
		return err
	}

	d, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		return err
	}
	return nil
}

// copyTree copies a directory tree from src to dst.
func copyTree(src, dst, skipSubdir string) error {
	return filepath.Walk(src, func(path string, fi fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if skipSubdir != "" && (path == skipSubdir || strings.HasPrefix(path, skipSubdir+string(filepath.Separator))) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, fi.Mode())
		}
		return copyFile(path, target)
	})
}
