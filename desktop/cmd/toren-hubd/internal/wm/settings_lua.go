package wm

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"toren/desktop/internal/hypr"
)

// The parts of settings.lua beyond the options: what settings.conf carries for a
// hyprlang Hyprland, written in the Lua config's own calls for Hyprland 0.57.
// Without these the window rules, app overrides, layer rules, custom shortcuts,
// animations, default apps, autostart and plugins the settings window saves
// would never reach a Lua-config compositor.
//
// Each call that raises an error on a bad entry (an unknown animation leaf, a
// chord Hyprland cannot parse) goes through pcall, so one bad entry cannot stop
// the rest of the file loading.

// genSettingsLuaExtra is everything settings.lua holds after the options.
func genSettingsLuaExtra(o Overrides) string {
	var b strings.Builder
	for _, part := range []string{
		genAppsLua(o),
		genAnimationsLua(o),
		genRulesLua(o),
		genKeybindsLua(o),
		genPluginsLua(o),
		genStartHookLua(o),
	} {
		if part != "" {
			b.WriteString(part)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func luaQ(s string) string { return hypr.QuoteLua(s) }

func genAppsLua(o Overrides) string {
	var b strings.Builder
	if v := strings.TrimSpace(o.Apps["browser"]); v != "" {
		fmt.Fprintf(&b, "hl.env(\"BROWSER\", %s)\n", luaQ(v))
	}
	if v := strings.TrimSpace(o.Apps["terminal"]); v != "" {
		fmt.Fprintf(&b, "hl.env(\"TERMINAL\", %s)\n", luaQ(v))
	}
	return b.String()
}

// --- animations -------------------------------------------------------------

// genAnimationsLua is genAnimations in Lua: the base curves, the chosen preset,
// the motion toggles, the user's curves and leaves, and the animated border.
func genAnimationsLua(o Overrides) string {
	preset := presetLines(CurrentAnimPreset())
	motion := genMotion(o, false)
	needsBase := preset != "" || motion != "" || len(o.Anim.Curves) > 0 || len(o.Anim.Items) > 0 || o.Appearance.AnimatedBorder
	if !needsBase {
		return ""
	}
	var b strings.Builder
	for _, src := range []string{baseCurves(), preset, motion, genAnimBlock(o)} {
		b.WriteString(animLinesToLua(src))
	}
	if o.Appearance.AnimatedBorder {
		a, d := o.Appearance, defaultOverrides().Appearance
		fmt.Fprintf(&b, "hl.config({ general = { col = { active_border = { colors = { %s, %s }, angle = 45 } } } })\n",
			luaQ(hlRGB(a.ActiveBorder, d.ActiveBorder)), luaQ(hlRGB(a.InactiveBorder, d.InactiveBorder)))
		fmt.Fprintf(&b, "pcall(hl.animation, { leaf = \"borderangle\", enabled = true, speed = %s, bezier = \"linear\", style = \"loop\" })\n",
			hlNum(borderAngleHyprSpeed(a.BorderAngleSpeed)))
	}
	return b.String()
}

// animLinesToLua turns hyprlang `bezier =` and `animation =` lines into
// hl.curve and hl.animation calls. Other lines (comments, blank) are dropped.
func animLinesToLua(src string) string {
	var b strings.Builder
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		key, val, ok := strings.Cut(line, "=")
		if !ok || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(val, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		switch strings.TrimSpace(key) {
		case "bezier":
			if len(parts) != 5 || !allNumbers(parts[1:]) {
				continue
			}
			fmt.Fprintf(&b, "pcall(hl.curve, %s, { type = \"bezier\", points = { {%s, %s}, {%s, %s} } })\n",
				luaQ(parts[0]), parts[1], parts[2], parts[3], parts[4])
		case "animation":
			if len(parts) < 2 || parts[0] == "" {
				continue
			}
			if parts[1] == "0" {
				fmt.Fprintf(&b, "pcall(hl.animation, { leaf = %s, enabled = false })\n", luaQ(parts[0]))
				continue
			}
			if len(parts) < 3 || !allNumbers(parts[2:3]) {
				continue
			}
			fields := []string{"leaf = " + luaQ(parts[0]), "enabled = true", "speed = " + parts[2]}
			if len(parts) >= 4 && parts[3] != "" {
				fields = append(fields, "bezier = "+luaQ(parts[3]))
			}
			if len(parts) >= 5 {
				if style := strings.TrimSpace(strings.Join(parts[4:], ",")); style != "" {
					fields = append(fields, "style = "+luaQ(style))
				}
			}
			fmt.Fprintf(&b, "pcall(hl.animation, { %s })\n", strings.Join(fields, ", "))
		}
	}
	return b.String()
}

func allNumbers(xs []string) bool {
	for _, x := range xs {
		if _, err := strconv.ParseFloat(x, 64); err != nil {
			return false
		}
	}
	return true
}

// --- window and layer rules -------------------------------------------------

// genRulesLua is the window rules, layer rules and per-app overrides, in the
// order settings.conf has them, each as one hl.window_rule / hl.layer_rule.
func genRulesLua(o Overrides) string {
	var blocks []string
	blocks = append(blocks, genTameMaximizeOnOpen(o))
	for i, r := range o.WindowRules {
		blocks = append(blocks, genWindowRule(i, r))
	}
	for i, r := range o.LayerRules {
		blocks = append(blocks, genLayerRule(i, r))
	}
	for i, a := range o.AppOverrides {
		blocks = append(blocks, genAppOverride(i, a))
	}
	var b strings.Builder
	for _, blk := range blocks {
		b.WriteString(ruleBlockToLua(blk))
	}
	return b.String()
}

// Rule fields whose Lua value is a string even when it looks like a number
// (opacity takes "0.9 0.8"-style values) or carries a space (size, move).
var ruleStringFields = map[string]bool{
	"opacity": true, "size": true, "move": true, "workspace": true, "monitor": true,
	"idle_inhibit": true, "suppress_event": true, "fullscreen_state": true, "tag": true,
	"animation": true, "content": true, "group": true, "max_size": true, "min_size": true,
}

// Rule fields that are integers in Lua although hyprlang writes them on/off.
var ruleIntOnOff = map[string]bool{"above_lock": true}

var ruleHeader = regexp.MustCompile(`^(windowrule|layerrule)\s*\{$`)

// ruleBlockToLua turns one block ruleBlock wrote into its Lua call; "" for an
// empty block.
func ruleBlockToLua(block string) string {
	lines := strings.Split(strings.TrimSpace(block), "\n")
	if len(lines) < 2 {
		return ""
	}
	m := ruleHeader.FindStringSubmatch(strings.TrimSpace(lines[0]))
	if m == nil {
		return ""
	}
	fn := "hl.window_rule"
	if m[1] == "layerrule" {
		fn = "hl.layer_rule"
	}
	var fields, match []string
	for _, l := range lines[1:] {
		l = strings.TrimSpace(l)
		if l == "}" || l == "" {
			continue
		}
		k, v, ok := strings.Cut(l, "=")
		if !ok {
			continue
		}
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if mk, isMatch := strings.CutPrefix(k, "match:"); isMatch {
			match = append(match, mk+" = "+luaQ(v))
			continue
		}
		fields = append(fields, k+" = "+ruleValueLua(k, v))
	}
	if len(match) == 0 || len(fields) < 2 { // name plus at least one effect
		return ""
	}
	return fmt.Sprintf("%s({ %s, match = { %s } })\n", fn, strings.Join(fields, ", "), strings.Join(match, ", "))
}

func ruleValueLua(k, v string) string {
	if k == "name" || ruleStringFields[k] {
		return luaQ(v)
	}
	if ruleIntOnOff[k] {
		switch v {
		case "on", "true":
			return "1"
		case "off", "false":
			return "0"
		}
	}
	switch v {
	case "on", "true":
		return "true"
	case "off", "false":
		return "false"
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}
	return luaQ(v)
}

// --- shortcuts --------------------------------------------------------------

// genKeybindsLua is the user's unbinds and their own shortcuts (Settings →
// Keybinds → Custom) as hl.unbind / hl.bind. Rebinds of shipped shortcuts are
// rebinds.lua's.
func genKeybindsLua(o Overrides) string {
	var b strings.Builder
	for _, c := range o.Unbinds {
		if c = strings.TrimSpace(c); c != "" {
			fmt.Fprintf(&b, "pcall(hl.unbind, %s)\n", luaQ(c))
		}
	}
	for _, k := range o.Keybinds {
		b.WriteString(keybindLua(k))
	}
	return b.String()
}

func keybindLua(k Keybind) string {
	chord := strings.TrimSpace(k.Keys)
	if chord == "" {
		return ""
	}
	var dsp string
	switch k.Action {
	case "exec", "":
		v := strings.TrimSpace(k.Value)
		if v == "" {
			return ""
		}
		dsp = "hl.dsp.exec_cmd(" + luaQ(v) + ")"
	case "close":
		dsp = "hl.dsp.window.close()"
	case "fullscreen":
		dsp = "hl.dsp.window.fullscreen(0)"
	case "togglefloating":
		dsp = "hl.dsp.window.float()"
	default:
		return ""
	}
	opts := ""
	if k.Release {
		opts = ", { release = true }"
	}
	var b strings.Builder
	// The number pad fires whichever way NumLock sits: a chord on KP_<digit>
	// binds its NumLock-off twin too, as settings.conf does.
	for _, c := range append([]string{chord}, NumpadAliases(chord)...) {
		fmt.Fprintf(&b, "pcall(hl.bind, %s, %s%s)\n", luaQ(c), dsp, opts)
	}
	return b.String()
}

// --- plugins and autostart --------------------------------------------------

// genPluginsLua loads each enabled, installed plugin and sets its options. A
// plugin that is switched on but not installed is a comment, as in
// settings.conf.
func genPluginsLua(o Overrides) string {
	var b strings.Builder
	var opts hlOpts
	for _, pb := range pluginBlocks(o) {
		if pb.so == "" {
			fmt.Fprintf(&b, "-- %s: not installed\n", pb.id)
			continue
		}
		fmt.Fprintf(&b, "pcall(hl.plugin.load, %s)\n", luaQ(pb.so))
		opts = append(opts, pb.opts...)
		opts = append(opts, pb.adds...)
	}
	if len(opts) > 0 {
		b.WriteString(opts.luaBlocks())
	}
	return b.String()
}

// genStartHookLua runs the cursor theme and the user's autostart commands once,
// when the compositor starts (exec-once in settings.conf).
func genStartHookLua(o Overrides) string {
	var cmds []string
	d := defaultOverrides().Cursor
	if o.Cursor.Theme != d.Theme || o.Cursor.Size != d.Size {
		cmds = append(cmds, fmt.Sprintf("hyprctl setcursor %s %d", hlStr(o.Cursor.Theme), o.Cursor.Size))
	}
	for _, a := range o.Autostart {
		if c := strings.TrimSpace(a.Command); c != "" {
			cmds = append(cmds, c)
		}
	}
	if len(cmds) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("hl.on(\"hyprland.start\", function()\n")
	for _, c := range cmds {
		fmt.Fprintf(&b, "    hl.exec_cmd(%s)\n", luaQ(c))
	}
	b.WriteString("end)\n")
	return b.String()
}

// --- the live preview -------------------------------------------------------

// keywordsToLua turns the preview's `keyword KEY VALUE` commands into one Lua
// chunk for `hyprctl eval`: Hyprland 0.57's Lua config has no keyword command.
func keywordsToLua(cmds []string) string {
	var b strings.Builder
	for _, c := range cmds {
		rest, ok := strings.CutPrefix(strings.TrimSpace(c), "keyword ")
		if !ok {
			continue
		}
		key, val, ok := strings.Cut(strings.TrimSpace(rest), " ")
		if !ok {
			continue
		}
		val = strings.TrimSpace(val)
		switch key {
		case "animation", "bezier":
			b.WriteString(animLinesToLua(key + " = " + strings.ReplaceAll(val, ",", ", ")))
		case "general:col.active_border", "general:col.inactive_border":
			b.WriteString(borderColourLua(key, val))
		default:
			if luaOnlyGone[key] {
				continue
			}
			b.WriteString(hypr.KeywordToLua(key, val))
			b.WriteString("\n")
		}
	}
	return b.String()
}

var gradientAngle = regexp.MustCompile(`^(\d+)deg$`)

// borderColourLua sets a border colour, a gradient ("rgb(a) rgb(b) 45deg")
// written as the table the Lua config wants.
func borderColourLua(key, val string) string {
	leaf := strings.TrimPrefix(key, "general:col.")
	var colours []string
	angle := ""
	for _, f := range strings.Fields(val) {
		if m := gradientAngle.FindStringSubmatch(f); m != nil {
			angle = m[1]
			continue
		}
		colours = append(colours, luaQ(f))
	}
	if len(colours) == 0 {
		return ""
	}
	v := colours[0]
	if len(colours) > 1 || angle != "" {
		v = "{ colors = { " + strings.Join(colours, ", ") + " }"
		if angle != "" {
			v += ", angle = " + angle
		}
		v += " }"
	}
	return fmt.Sprintf("hl.config({ general = { col = { %s = %s } } })\n", leaf, v)
}
