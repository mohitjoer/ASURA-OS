#!/usr/bin/env python3
"""Translate end-4's Lua Hyprland dispatches to classic hyprlang dispatchers.

end-4's desktop is written for Hyprland 0.55+, whose config and IPC speak Lua:

    Hyprland.dispatch(`hl.dsp.focus({workspace = ${id}})`)

ASURA ships Hyprland 0.49, where the same thing is the hyprlang dispatcher

    Hyprland.dispatch(`workspace ${id}`)

and a Lua dispatch is an unknown dispatcher that silently does nothing. Every
`hl.dsp.*(...)` and `hl.config(...)` call in a file is rewritten in place; the
JavaScript `${...}` expressions inside are carried over untouched.

    python3 hyprlang.py FILE...      rewrite files in place
    python3 hyprlang.py --test       run the translation table

shell-end4/asura/vendor.sh runs this on every synced file that contains a Lua
dispatch, so a new pin is translated the same way. A form it does not know is
an error, not a guess: the sync stops and names the call.
"""
import re
import sys


class Unknown(Exception):
    pass


def _split_top(s, sep=","):
    """Split on sep outside ${...}, quotes, braces and parentheses."""
    parts, depth, quote, buf, i = [], 0, None, "", 0
    while i < len(s):
        c = s[i]
        if quote:
            buf += c
            if c == quote:
                quote = None
        elif c in "\"'":
            quote = c
            buf += c
        elif s.startswith("${", i):
            depth += 1
            buf += "${"
            i += 2
            continue
        elif c in "({[":
            depth += 1
            buf += c
        elif c in ")}]":
            depth -= 1
            buf += c
        elif c == sep and depth == 0:
            parts.append(buf)
            buf = ""
        else:
            buf += c
        i += 1
    if buf.strip():
        parts.append(buf)
    return [p.strip() for p in parts]


def _unquote(v):
    v = v.strip()
    if len(v) >= 2 and v[0] == v[-1] and v[0] in "\"'":
        return v[1:-1]
    return v


def _table(arg):
    """Parse a Lua table literal `{ k = v, ... }` into a dict of raw values."""
    arg = arg.strip()
    if not (arg.startswith("{") and arg.endswith("}")):
        raise Unknown(arg)
    out = {}
    for item in _split_top(arg[1:-1]):
        if "=" not in item:
            raise Unknown(arg)
        k, v = item.split("=", 1)
        out[k.strip()] = v.strip()
    return out


def translate_call(name, arg):
    """One Lua call (without the `hl.` prefix) to a hyprlang dispatch string."""
    if name == "dsp.focus":
        t = _table(arg)
        if set(t) == {"workspace"}:
            return "workspace " + _unquote(t["workspace"])
        if set(t) == {"window"}:
            return "focuswindow " + _unquote(t["window"])
        if set(t) == {"monitor"}:
            return "focusmonitor " + _unquote(t["monitor"])
    elif name == "dsp.window.move":
        t = _table(arg)
        win = _unquote(t.get("window", ""))
        if "workspace" in t and set(t) <= {"workspace", "follow", "window"}:
            silent = _unquote(t.get("follow", "true")) == "false"
            cmd = "movetoworkspacesilent " if silent else "movetoworkspace "
            return cmd + _unquote(t["workspace"]) + ("," + win if win else "")
        if "x" in t and "y" in t and set(t) <= {"x", "y", "window"}:
            return "movewindowpixel exact " + _unquote(t["x"]) + " " + _unquote(t["y"]) + ("," + win if win else "")
    elif name == "dsp.window.close":
        t = _table(arg)
        if set(t) == {"window"}:
            return "closewindow " + _unquote(t["window"])
    elif name == "dsp.window.pin":
        t = _table(arg)
        if set(t) == {"window"}:
            return "pin " + _unquote(t["window"])
    elif name == "dsp.workspace.toggle_special":
        return "togglespecialworkspace " + _unquote(arg)
    elif name == "dsp.cursor.move":
        t = _table(arg)
        if set(t) == {"x", "y"}:
            return "movecursor " + _unquote(t["x"]) + " " + _unquote(t["y"])
    elif name == "dsp.global":
        return "global " + _unquote(arg)
    elif name == "config":
        # hl.config({ section = { key = value } }) -> `hyprctl keyword section:key value`,
        # run through the exec dispatcher because hyprlang has no keyword dispatch.
        t = _table(arg)
        keys = []
        for section, inner in t.items():
            for k, v in _table(inner).items():
                keys.append(f"{section}:{k} {_unquote(v)}")
        if len(keys) == 1:
            return "exec hyprctl keyword " + keys[0]
    raise Unknown(f"hl.{name}({arg})")


CALL = re.compile(r"hl\.((?:dsp\.)?[a-z_]+(?:\.[a-z_]+)*)\(")


def translate_text(text):
    out, i = [], 0
    while True:
        m = CALL.search(text, i)
        if not m:
            out.append(text[i:])
            return "".join(out)
        # balanced parentheses for the argument
        j, depth = m.end(), 1
        while depth:
            if j >= len(text):
                raise Unknown(text[m.start():m.start() + 80])
            if text[j] == "(":
                depth += 1
            elif text[j] == ")":
                depth -= 1
            j += 1
        out.append(text[i:m.start()])
        out.append(translate_call(m.group(1), text[m.end():j - 1]))
        i = j


CASES = [
    ('hl.dsp.focus({workspace = ${id}})', 'workspace ${id}'),
    ('hl.dsp.focus({ workspace = ${workspace.workspaceValue} })', 'workspace ${workspace.workspaceValue}'),
    ('hl.dsp.focus({workspace = ${wsModel.getWorkspaceIdAt(hoverIndex)}})', 'workspace ${wsModel.getWorkspaceIdAt(hoverIndex)}'),
    ('hl.dsp.focus({workspace = "r+1"})', 'workspace r+1'),
    ('hl.dsp.focus({workspace=${2147483647 - ws}})', 'workspace ${2147483647 - ws}'),
    ('hl.dsp.focus({window = "address:${a?.address}"})', 'focuswindow address:${a?.address}'),
    ('hl.dsp.focus({monitor="${monName}"})', 'focusmonitor ${monName}'),
    ('hl.dsp.window.move({ workspace = ${t}, follow = false, window = "address:${w.address}" })',
     'movetoworkspacesilent ${t},address:${w.address}'),
    ('hl.dsp.window.move({ x = "${px * w}", y = "${py * h}", window = "address:${w.address}" })',
     'movewindowpixel exact ${px * w} ${py * h},address:${w.address}'),
    ('hl.dsp.window.close({window = "address:${a}"})', 'closewindow address:${a}'),
    ('hl.dsp.window.pin({window = "address:${a}"})', 'pin address:${a}'),
    ('hl.dsp.workspace.toggle_special("special")', 'togglespecialworkspace special'),
    ('hl.dsp.cursor.move({x=9999,y=9999})', 'movecursor 9999 9999'),
    ('hl.dsp.cursor.move({x=${cx},y=${cy}})', 'movecursor ${cx} ${cy}'),
    ('hl.dsp.global("quickshell:wallpaperSelectorToggle")', 'global quickshell:wallpaperSelectorToggle'),
    ('hl.config({cursor = {no_warps = true}})', 'exec hyprctl keyword cursor:no_warps true'),
    ("hyprctl dispatch 'hl.dsp.focus({monitor=\"${m}\"})'; x", "hyprctl dispatch 'focusmonitor ${m}'; x"),
]


def main(argv):
    if argv == ["--test"]:
        bad = 0
        for src, want in CASES:
            got = translate_text(src)
            if got != want:
                bad += 1
                print(f"FAIL {src}\n  got  {got}\n  want {want}")
        try:
            translate_text("hl.dsp.window.resize({ w = 1 })")
            print("FAIL an unknown form was not refused")
            bad += 1
        except Unknown:
            pass
        print(f"hyprlang: {len(CASES) - bad}/{len(CASES)} translations ok")
        return 1 if bad else 0
    for path in argv:
        with open(path, encoding="utf-8") as f:
            text = f.read()
        try:
            new = translate_text(text)
        except Unknown as e:
            print(f"hyprlang: {path}: no hyprlang form for {e}", file=sys.stderr)
            return 1
        if new != text:
            with open(path, "w", encoding="utf-8") as f:
                f.write(new)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
