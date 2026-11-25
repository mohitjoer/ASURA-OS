#!/bin/sh
# First login for a user: give the desktop what it expects to find in
# the home directory, without overwriting anything the user already has.
# Run by AsuraBridge.qml once the desktop config has loaded.
#
# 1. ~/.config/matugen/config.toml, the config switchwall.sh runs
#    matugen with (the shell's palette comes from it).
# 2. The wallpaper and its palette: switchwall.sh on the user's
#    ASURA wallpaper, or a colourful ASURA background when that is the near-
#    black default, so the desktop starts in vivid colours, not grey.
set -u

HERE=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
SHELL_DIR=$(dirname "$HERE")
CONFIG="${XDG_CONFIG_HOME:-$HOME/.config}"
STATE="${XDG_STATE_HOME:-$HOME/.local/state}"
LOG="$STATE/asura/desktop-first-run.log"
mkdir -p "$STATE/asura"
exec >>"$LOG" 2>&1
echo "== $(date) first-run"

if [ ! -f "$CONFIG/matugen/config.toml" ]; then
	mkdir -p "$CONFIG/matugen"
	cp "$SHELL_DIR/matugen/config.toml" "$CONFIG/matugen/config.toml"
	echo "seeded $CONFIG/matugen/config.toml"
fi

# The palette and the wallpaper, once per user (a marker), or again whenever
# no palette exists. The desktop takes all its colours from the wallpaper, so a near-
# black one (ASURA's default galaxy) gives a grey desktop: unless the user chose
# another wallpaper, the desktop starts on a colourful one of ASURA's.
LOOK_MARK="$STATE/asura/desktop-look-2"
VIVID=/usr/share/asura/backgrounds/eva-berserk.jpg
if [ ! -e "$LOOK_MARK" ] || [ ! -s "$STATE/quickshell/user/generated/colors.json" ]; then
	if ! command -v matugen >/dev/null 2>&1; then
		echo "matugen is not installed; desktop keeps its default palette (sudo xbps-install -S matugen)"
		exit 0
	fi
	wp=$(gsettings get org.asura.shell wallpaper-path 2>/dev/null | tr -d "'")
	case "$wp" in ''|*/galaxy.png|*/default_wallpaper.png) wp=$VIVID ;; esac
	if [ ! -f "$wp" ]; then
		for f in /usr/share/asura/backgrounds/*; do
			[ -f "$f" ] && { wp=$f; break; }
		done
	fi
	if [ -f "$wp" ]; then
		echo "wallpaper and palette from $wp"
		bash "$SHELL_DIR/scripts/colors/switchwall.sh" --image "$wp" && touch "$LOOK_MARK"
	else
		echo "no wallpaper found; desktop keeps its default palette"
	fi
fi
