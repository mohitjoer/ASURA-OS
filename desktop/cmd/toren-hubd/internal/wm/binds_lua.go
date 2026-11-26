package wm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"toren/desktop/internal/hypr"
)

// luaBind is one hl.bind call of the packaged binds.lua: the action expression
// and the options table, exactly as written.
type luaBind struct {
	expr string
	opts string // "" or "{ release = true }"
}

// PackagedBindsLuaPath is the binds.lua beside the packaged configuration.
func PackagedBindsLuaPath() string {
	return filepath.Join(filepath.Dir(PackagedConfigPath()), "binds.lua")
}

// parseLuaBinds reads every top-level hl.bind("CHORD", ACTION[, OPTS]) call of a
// binds.lua, keyed by chordKey. A bind written over several lines (a function
// body) is read whole. A file that cannot be read gives an empty map.
func parseLuaBinds(path string) map[string][]luaBind {
	out := map[string][]luaBind{}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	src := string(data)
	for i := 0; ; {
		j := strings.Index(src[i:], "hl.bind(")
		if j < 0 {
			break
		}
		start := i + j
		// only a call at the start of a line, not one inside another expression
		if start > 0 && src[start-1] != '\n' {
			i = start + len("hl.bind(")
			continue
		}
		args, end := callArgs(src, start+len("hl.bind("))
		i = end
		if len(args) < 2 {
			continue
		}
		chord := strings.TrimSpace(args[0])
		if len(chord) < 2 || chord[0] != '"' || chord[len(chord)-1] != '"' {
			continue
		}
		b := luaBind{expr: strings.TrimSpace(args[1])}
		if len(args) > 2 {
			b.opts = strings.TrimSpace(args[2])
		}
		k := chordKey(chord[1 : len(chord)-1])
		out[k] = append(out[k], b)
	}
	return out
}

// callArgs splits the argument list that starts at src[from] (just past the
// opening parenthesis) on its top-level commas, skipping strings and comments,
// and returns the arguments and the index past the closing parenthesis.
func callArgs(src string, from int) ([]string, int) {
	var args []string
	depth, last := 0, from
	for i := from; i < len(src); i++ {
		switch c := src[i]; c {
		case '"', '\'':
			for i++; i < len(src) && src[i] != c; i++ {
				if src[i] == '\\' {
					i++
				}
			}
		case '-':
			if i+1 < len(src) && src[i+1] == '-' {
				for i < len(src) && src[i] != '\n' {
					i++
				}
			}
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			if depth == 0 {
				return append(args, src[last:i]), i + 1
			}
			depth--
		case ',':
			if depth == 0 {
				args = append(args, src[last:i])
				last = i + 1
			}
		}
	}
	return nil, len(src)
}

// confBindLua is a hyprlang bind's action in the Lua form: the fallback when the
// packaged binds.lua cannot be read. flags carries r (release), l (locked) and e
// (repeating).
func confBindLua(cb confBind) luaBind {
	disp, args, _ := strings.Cut(cb.rest, ",")
	disp, args = strings.TrimSpace(disp), strings.TrimSpace(args)
	b := luaBind{expr: hypr.DispatchToLua(disp, args)}
	var opts []string
	for flag, name := range map[byte]string{'r': "release", 'l': "locked", 'e': "repeating"} {
		if strings.IndexByte(cb.flags, flag) >= 0 {
			opts = append(opts, name+" = true")
		}
	}
	if len(opts) > 0 {
		b.opts = "{ " + strings.Join(opts, ", ") + " }"
	}
	return b
}

func writeLuaBind(b *strings.Builder, chord string, lb luaBind) {
	if lb.opts != "" {
		fmt.Fprintf(b, "hl.bind(%s, %s, %s)\n", hypr.QuoteLua(chord), lb.expr, lb.opts)
		return
	}
	fmt.Fprintf(b, "hl.bind(%s, %s)\n", hypr.QuoteLua(chord), lb.expr)
}
