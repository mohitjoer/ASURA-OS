package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// `asura mirror` chooses the Void mirror XBPS downloads from.
//
// The choice is written to /etc/xbps.d/00-asura-mirror.conf. XBPS reads that
// folder in name order and tries repositories in the order it reads them, so
// this file, sorting first, puts the chosen mirror ahead of the defaults in
// asura.conf, which stay as the fallback. It is a file of its own because
// apply-repo.sh rewrites asura.conf and 00-repository-main.conf on every
// update; this one no update touches.

type voidMirror struct {
	Name, Where, URL string
}

var voidMirrors = []voidMirror{
	{"fastly", "worldwide (CDN)", "https://repo-fastly.voidlinux.org"},
	{"default", "Finland", "https://repo-default.voidlinux.org"},
	{"fi", "Finland", "https://repo-fi.voidlinux.org"},
	{"de", "Germany", "https://repo-de.voidlinux.org"},
	{"se", "Sweden", "https://ftp.lysator.liu.se/pub/voidlinux"},
	{"us", "USA (Chicago)", "https://mirrors.servercentral.com/voidlinux"},
	{"au", "Australia", "https://mirror.aarnet.edu.au/pub/voidlinux"},
	{"cn", "China", "https://mirrors.tuna.tsinghua.edu.cn/voidlinux"},
	{"ru", "Russia", "https://mirror.yandex.ru/mirrors/voidlinux"},
}

// The repositories a mirror serves for ASURA (glibc, x86_64).
var mirrorRepos = []string{"current", "current/nonfree", "current/multilib", "current/multilib/nonfree"}

var (
	mirrorConf      = "/etc/xbps.d/00-asura-mirror.conf"
	mirrorASURAConf = "/etc/xbps.d/asura.conf"
	mirrorWrite     = writeFilePrivileged
	mirrorProbe     = probeMirror
)

const mirrorUsage = `Usage: asura mirror [list | fastest | set NAME|URL | reset]

Chooses the Void mirror packages download from.

  list      the known mirrors, and the one in use (the default)
  fastest   time each mirror and use the quickest
  set       use a mirror by name (see list) or by its https:// URL
  reset     back to ASURA's default (worldwide CDN)

The choice is /etc/xbps.d/00-asura-mirror.conf; updates keep it.
`

func init() {
	register("mirror", "choose the Void mirror packages download from", asuraMirror)
}

func asuraMirror(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		return listMirrors()
	}
	switch args[0] {
	case "-h", "--help", "help":
		fmt.Print(mirrorUsage)
		return exitOK
	case "reset":
		if err := mirrorWrite(mirrorConf, ""); err != nil {
			fmt.Fprintln(os.Stderr, "asura mirror:", err)
			return exitFail
		}
		fmt.Println("mirror: back to ASURA's default (worldwide CDN)")
		return exitOK
	case "set":
		if len(args) != 2 {
			fmt.Fprint(os.Stderr, mirrorUsage)
			return exitUsage
		}
		url, err := resolveMirror(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "asura mirror:", err)
			return exitUsage
		}
		return useMirror(url)
	case "fastest":
		fmt.Println("timing the mirrors...")
		ranked := rankMirrors(voidMirrors, mirrorProbe)
		if len(ranked) == 0 {
			fmt.Fprintln(os.Stderr, "asura mirror: no mirror answered; nothing changed")
			return exitFail
		}
		for _, r := range ranked {
			fmt.Printf("  %-8s %5d ms  %s\n", r.mirror.Name, r.took.Milliseconds(), r.mirror.Where)
		}
		return useMirror(ranked[0].mirror.URL)
	}
	fmt.Fprint(os.Stderr, mirrorUsage)
	return exitUsage
}

func resolveMirror(arg string) (string, error) {
	for _, m := range voidMirrors {
		if m.Name == arg {
			return m.URL, nil
		}
	}
	if strings.HasPrefix(arg, "https://") && !strings.ContainsAny(arg, " \t\n\"'") {
		return strings.TrimRight(arg, "/"), nil
	}
	return "", fmt.Errorf("%q is not a known mirror (asura mirror list) or an https:// URL", arg)
}

func useMirror(url string) int {
	if err := mirrorWrite(mirrorConf, mirrorConfig(url, readLines(mirrorASURAConf))); err != nil {
		fmt.Fprintln(os.Stderr, "asura mirror:", err)
		return exitFail
	}
	fmt.Printf("mirror: %s (sudo xbps-install -S refreshes the package lists)\n", url)
	return exitOK
}

// mirrorConfig is the 00-asura-mirror.conf for a mirror. The non-Void
// repositories asura.conf names (the Hyprland one) come first, as they do
// there, so the mirror never changes which repository a package comes from.
func mirrorConfig(url string, asuraConf []string) string {
	var b strings.Builder
	b.WriteString("# Written by `asura mirror`; `asura mirror reset` removes it.\n")
	for _, line := range asuraConf {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, "repository="); ok && !strings.Contains(v, "voidlinux") {
			b.WriteString(line + "\n")
		}
	}
	for _, repo := range mirrorRepos {
		fmt.Fprintf(&b, "repository=%s/%s\n", url, repo)
	}
	return b.String()
}

// currentMirror is the Void mirror the machine uses: the chosen one, else the
// first in asura.conf.
func currentMirror() string {
	for _, file := range []string{mirrorConf, mirrorASURAConf} {
		for _, line := range readLines(file) {
			if v, ok := strings.CutPrefix(strings.TrimSpace(line), "repository="); ok {
				if base, ok := strings.CutSuffix(v, "/current"); ok {
					return base
				}
			}
		}
	}
	return ""
}

func listMirrors() int {
	current := currentMirror()
	for _, m := range voidMirrors {
		mark := " "
		if m.URL == current {
			mark = "*"
		}
		fmt.Printf("%s %-8s %-18s %s\n", mark, m.Name, m.Where, m.URL)
	}
	if current != "" && !strings.Contains(strings.Join(mirrorURLs(), " "), current) {
		fmt.Printf("* %-8s %-18s %s\n", "custom", "", current)
	}
	fmt.Println("\nasura mirror fastest picks the quickest; asura mirror set NAME picks one.")
	return exitOK
}

func mirrorURLs() []string {
	var out []string
	for _, m := range voidMirrors {
		out = append(out, m.URL)
	}
	return out
}

type mirrorTiming struct {
	mirror voidMirror
	took   time.Duration
}

// rankMirrors probes every mirror at once and returns those that answered,
// quickest first.
func rankMirrors(mirrors []voidMirror, probe func(string) (time.Duration, error)) []mirrorTiming {
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		out []mirrorTiming
	)
	for _, m := range mirrors {
		wg.Add(1)
		go func(m voidMirror) {
			defer wg.Done()
			if took, err := probe(m.URL); err == nil {
				mu.Lock()
				out = append(out, mirrorTiming{m, took})
				mu.Unlock()
			}
		}(m)
	}
	wg.Wait()
	sort.Slice(out, func(i, j int) bool { return out[i].took < out[j].took })
	return out
}

// probeMirror times fetching the start of the mirror's package index.
func probeMirror(url string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/current/x86_64-repodata", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", "bytes=0-65535")
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	buf := make([]byte, 32*1024)
	for {
		if _, err := resp.Body.Read(buf); err != nil {
			break
		}
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("%s", resp.Status)
	}
	return time.Since(start), nil
}

func readLines(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var lines []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines
}

// writeFilePrivileged writes (or, given "", removes) a root-owned file,
// through sudo when not root.
func writeFilePrivileged(path, content string) error {
	if content == "" {
		return runPrivileged("rm", "-f", path)
	}
	tmp, err := os.CreateTemp("", "asura-mirror-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()
	if err := runPrivileged("mkdir", "-p", filepath.Dir(path)); err != nil {
		return err
	}
	return runPrivileged("install", "-m", "0644", tmp.Name(), path)
}
