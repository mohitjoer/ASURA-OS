package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// `asura channel get` is read by the Hub's Updates page to say which channel
// the machine follows. `set` is not here: moving a machine between channels
// is one line in the repository config file.

// xbpsRepoConf is the repository configuration file for Void Linux XBPS.
const xbpsRepoConf = "/etc/xbps.d/asura.conf"

const channelUsage = `Usage: asura channel get [--root DIR]
       asura channel set stable|unstable [--root DIR]

get prints the channel this machine follows. set chooses it:

  stable     follow main, where releases are cut
  unstable   follow dev-unstable, which every merged change reaches first

Either way, asura update only installs commits that passed CI.
`

func init() {
	register("channel", "the channel updates follow (get, or set stable|unstable)", asuraChannel)
}

func asuraChannel(args []string) int {
	set := newFlags("channel", channelUsage)
	root := set.String("root", "", "operate on a tree under this directory")
	positional, err := parseFlags(set, args)
	if err != nil {
		return exitUsage
	}
	if len(positional) == 2 && positional[0] == "set" {
		return channelSet(*root, positional[1])
	}
	if len(positional) != 1 || positional[0] != "get" {
		fmt.Fprint(os.Stderr, channelUsage)
		return exitUsage
	}
	if ch := pinnedChannel(*root); ch != "" {
		fmt.Println(ch)
		return exitOK
	}

	channel, err := xbpsChannel(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "asura channel:", err)
		return exitFail
	}
	fmt.Println(channel)
	return exitOK
}

// xbpsChannel reads the channel or primary repository from Asura's XBPS configuration.
func xbpsChannel(root string) (string, error) {
	path := filepath.Join(root, xbpsRepoConf)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			alt := filepath.Join(root, "/usr/share/xbps.d/asura.conf")
			data, err = os.ReadFile(alt)
			if err != nil {
				return "", fmt.Errorf("no Asura repository is configured: %s does not exist", path)
			}
		} else {
			return "", err
		}
	}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		if field, value, found := strings.Cut(trimmed, "="); found {
			f := strings.ToLower(strings.TrimSpace(field))
			if f == "channel" {
				return strings.TrimSpace(value), nil
			}
			if f == "repository" {
				repo := strings.TrimSpace(value)
				if repo != "" {
					return repo, nil
				}
			}
		}
	}
	return "", fmt.Errorf("%s names no channel or repository", path)
}
