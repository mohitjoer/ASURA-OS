package hypr

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// QuoteLua returns s as a safely escaped Lua string literal with double quotes.
func QuoteLua(s string) string {
	return strconv.Quote(s)
}

// formatLuaKey formats a Lua table key. If it's a valid Lua identifier, it returns it as is.
// Otherwise it returns ["key"].
func formatLuaKey(k string) string {
	if len(k) == 0 {
		return `[""]`
	}
	for i, r := range k {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			return "[" + strconv.Quote(k) + "]"
		}
		if i > 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return "[" + strconv.Quote(k) + "]"
		}
	}
	return k
}

// formatLuaVal converts a string value to appropriate Lua representation (bool, int, float, or quoted string).
func formatLuaVal(v string) string {
	vTrim := strings.TrimSpace(v)
	switch vTrim {
	case "true", "on", "yes":
		return "true"
	case "false", "off", "no":
		return "false"
	}
	if n, err := strconv.ParseInt(vTrim, 10, 64); err == nil {
		return strconv.FormatInt(n, 10)
	}
	if f, err := strconv.ParseFloat(vTrim, 64); err == nil && strings.Contains(vTrim, ".") {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return QuoteLua(v)
}

func parseWsWin(arg string) (ws, win string) {
	if idx := strings.Index(arg, ","); idx != -1 {
		return strings.TrimSpace(arg[:idx]), strings.TrimSpace(arg[idx+1:])
	}
	return strings.TrimSpace(arg), ""
}

func parseResizeActive(arg string) string {
	parts := strings.Fields(arg)
	if len(parts) >= 3 && parts[0] == "exact" {
		return fmt.Sprintf("hl.dsp.window.resize({ x = %s, y = %s })", parts[1], parts[2])
	} else if len(parts) >= 2 {
		return fmt.Sprintf("hl.dsp.window.resize({ x = %s, y = %s, relative = true })", parts[0], parts[1])
	}
	return "hl.dsp.window.resize()"
}

func parseMoveWindowPixel(arg string) string {
	var win string
	mainPart := arg
	if idx := strings.Index(arg, ","); idx != -1 {
		mainPart = strings.TrimSpace(arg[:idx])
		win = strings.TrimSpace(arg[idx+1:])
	}
	parts := strings.Fields(mainPart)
	var x, y string
	if len(parts) >= 3 && parts[0] == "exact" {
		x, y = parts[1], parts[2]
	} else if len(parts) >= 2 {
		x, y = parts[0], parts[1]
	}
	if x != "" && y != "" {
		if win != "" {
			return fmt.Sprintf("hl.dsp.window.move({ x = %s, y = %s, window = %s })", x, y, QuoteLua(win))
		}
		return fmt.Sprintf("hl.dsp.window.move({ x = %s, y = %s })", x, y)
	}
	if win != "" {
		return fmt.Sprintf("hl.dsp.window.move({ window = %s })", QuoteLua(win))
	}
	return "hl.dsp.window.move()"
}

// DispatchToLua converts a legacy Hyprland dispatch call (name and arguments)
// to a Hyprland 0.57.2 Lua expression.
func DispatchToLua(name string, args ...string) string {
	argStr := strings.TrimSpace(strings.Join(args, " "))
	if strings.HasPrefix(name, "hl.") {
		if argStr == "" {
			return name
		}
		return name + " " + argStr
	}

	switch name {
	case "workspace":
		return fmt.Sprintf("hl.dsp.focus({ workspace = %s })", QuoteLua(argStr))

	case "movetoworkspace":
		ws, win := parseWsWin(argStr)
		if win != "" {
			return fmt.Sprintf("hl.dsp.window.move({ workspace = %s, window = %s })", QuoteLua(ws), QuoteLua(win))
		}
		return fmt.Sprintf("hl.dsp.window.move({ workspace = %s })", QuoteLua(ws))

	case "movetoworkspacesilent":
		ws, win := parseWsWin(argStr)
		if win != "" {
			return fmt.Sprintf("hl.dsp.window.move({ workspace = %s, follow = false, window = %s })", QuoteLua(ws), QuoteLua(win))
		}
		return fmt.Sprintf("hl.dsp.window.move({ workspace = %s, follow = false })", QuoteLua(ws))

	case "focuswindow":
		return fmt.Sprintf("hl.dsp.focus({ window = %s })", QuoteLua(argStr))

	case "closewindow":
		if argStr != "" {
			return fmt.Sprintf("hl.dsp.window.close({ window = %s })", QuoteLua(argStr))
		}
		return "hl.dsp.window.close()"

	case "killactive":
		return "hl.dsp.window.close()"

	case "pin":
		if argStr != "" {
			return fmt.Sprintf("hl.dsp.window.pin({ window = %s })", QuoteLua(argStr))
		}
		return "hl.dsp.window.pin()"

	case "togglefloating":
		if argStr == "" {
			return "hl.dsp.window.float()"
		}
		if argStr == "enable" || argStr == "disable" || argStr == "toggle" {
			return fmt.Sprintf("hl.dsp.window.float({ action = %s })", QuoteLua(argStr))
		}
		if strings.HasPrefix(argStr, "address:") {
			return fmt.Sprintf("hl.dsp.window.float({ window = %s })", QuoteLua(argStr))
		}
		parts := strings.Split(argStr, ",")
		if len(parts) == 2 {
			return fmt.Sprintf("hl.dsp.window.float({ action = %s, window = %s })", QuoteLua(parts[0]), QuoteLua(parts[1]))
		}
		return fmt.Sprintf("hl.dsp.window.float({ window = %s })", QuoteLua(argStr))

	case "movewindowpixel":
		return parseMoveWindowPixel(argStr)

	case "resizeactive":
		return parseResizeActive(argStr)

	case "centerwindow":
		return "hl.dsp.window.center()"

	case "bringactivetotop":
		return "hl.dsp.window.bring_to_top()"

	case "cyclenext":
		if strings.TrimSpace(argStr) == "prev" {
			return "hl.dsp.window.cycle_next({ next = false })"
		}
		return "hl.dsp.window.cycle_next()"

	case "exec":
		return fmt.Sprintf("hl.dsp.exec_cmd(%s)", QuoteLua(argStr))

	case "global":
		return fmt.Sprintf("hl.dsp.global(%s)", QuoteLua(argStr))

	case "submap":
		return fmt.Sprintf("hl.dsp.submap(%s)", QuoteLua(argStr))

	case "dpms":
		parts := strings.Fields(argStr)
		if len(parts) >= 2 {
			return fmt.Sprintf("hl.dsp.dpms({ action = %s, monitor = %s })", QuoteLua(parts[0]), QuoteLua(parts[1]))
		} else if len(parts) == 1 {
			return fmt.Sprintf("hl.dsp.dpms({ action = %s })", QuoteLua(parts[0]))
		}
		return "hl.dsp.dpms()"

	case "fullscreen":
		if argStr == "1" || argStr == "maximized" {
			return "hl.dsp.window.fullscreen({ mode = \"1\" })"
		} else if argStr == "0" || argStr == "fullscreen" || argStr == "" {
			return "hl.dsp.window.fullscreen()"
		}
		return fmt.Sprintf("hl.dsp.window.fullscreen({ mode = %s })", QuoteLua(argStr))

	case "exit":
		return "hl.dsp.exit()"

	case "togglespecialworkspace":
		if argStr != "" {
			return fmt.Sprintf("hl.dsp.workspace.toggle_special(%s)", QuoteLua(argStr))
		}
		return "hl.dsp.workspace.toggle_special()"

	case "moveworkspacetomonitor":
		parts := strings.Fields(argStr)
		if len(parts) >= 2 {
			return fmt.Sprintf("hl.dsp.workspace.move({ workspace = %s, monitor = %s })", QuoteLua(parts[0]), QuoteLua(parts[1]))
		} else if len(parts) == 1 {
			return fmt.Sprintf("hl.dsp.workspace.move({ monitor = %s })", QuoteLua(parts[0]))
		}
		return "hl.dsp.workspace.move()"

	default:
		if argStr != "" {
			return fmt.Sprintf("hl.dsp.%s(%s)", name, QuoteLua(argStr))
		}
		return fmt.Sprintf("hl.dsp.%s()", name)
	}
}

// KeywordToLua converts a legacy `hyprctl keyword <key> <value>` into a
// Hyprland 0.57.2 Lua expression evaluated via `hyprctl eval '<expr>'`.
func KeywordToLua(key, value string) string {
	if key == "workspace" {
		if idx := strings.Index(value, ",layout:"); idx != -1 {
			ws := strings.TrimSpace(value[:idx])
			layout := strings.TrimSpace(value[idx+len(",layout:"):])
			return fmt.Sprintf("hl.workspace_rule({ workspace = %s, layout = %s })", QuoteLua(ws), QuoteLua(layout))
		}
	}
	if key == "monitor" {
		parts := strings.Split(value, ",")
		if len(parts) >= 2 && strings.TrimSpace(parts[1]) == "disable" {
			return fmt.Sprintf("hl.monitor({ output = %s, disabled = true })", QuoteLua(strings.TrimSpace(parts[0])))
		}
		if len(parts) >= 4 {
			return fmt.Sprintf("hl.monitor({ output = %s, mode = %s, position = %s, scale = %s })",
				QuoteLua(strings.TrimSpace(parts[0])),
				QuoteLua(strings.TrimSpace(parts[1])),
				QuoteLua(strings.TrimSpace(parts[2])),
				QuoteLua(strings.TrimSpace(parts[3])))
		}
	}

	sections := strings.Split(key, ":")
	valExpr := formatLuaVal(value)

	var sb strings.Builder
	sb.WriteString("hl.config({ ")
	for i, sec := range sections {
		if i > 0 {
			sb.WriteString("{ ")
		}
		sb.WriteString(formatLuaKey(sec))
		sb.WriteString(" = ")
	}
	sb.WriteString(valExpr)
	for i := len(sections) - 1; i >= 0; i-- {
		sb.WriteString(" }")
	}
	sb.WriteString(")")
	return sb.String()
}
