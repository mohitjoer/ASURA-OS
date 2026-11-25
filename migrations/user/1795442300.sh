#!/bin/sh
echo "Migrate stored desktop shell names and state from end4 to desktop"

# 1. GSettings: org.asura.shell desktop-shell 'end4' -> 'desktop'
if command -v gsettings >/dev/null 2>&1; then
    val=$(gsettings get org.asura.shell desktop-shell 2>/dev/null | tr -d "'\"")
    if [ "$val" = "end4" ]; then
        gsettings set org.asura.shell desktop-shell desktop 2>/dev/null || true
    fi
fi

# 2. State files in ~/.local/state/asura/ (or $XDG_STATE_HOME/asura/)
for statedir in "${XDG_STATE_HOME:-$HOME/.local/state}/asura" "${XDG_STATE_HOME:-$HOME/.local/state}"; do
    [ -d "$statedir" ] || continue
    for name in quick-exits last-start first-run.log look-2 layout-6; do
        old="$statedir/end4-$name"
        new="$statedir/desktop-$name"
        if [ -e "$old" ]; then
            if [ ! -e "$new" ]; then
                mv "$old" "$new"
            else
                rm -f "$old"
            fi
        fi
    done
done

exit 0
