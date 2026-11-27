#!/bin/sh
# Sync and verify the vendored end-4 desktop (illogical-impulse's `ii`
# Quickshell config) that runs when the desktop shell is set to "end4".
#
#   sh shell-desktop/toren/vendor.sh sync     write the tree at the pinned commit
#   sh shell-desktop/toren/vendor.sh verify   check every file against the manifest
#   sh shell-desktop/toren/vendor.sh update <commit>   move the pin, then sync
#
# The tree is upstream's, byte for byte, like shell/vendor's: a change goes in
# an overlay (a file ASURA owns at upstream's path, listed in OVERLAYS) or in
# ASURA's own files under shell-desktop/toren/, never in a vendored file.
set -eu

HERE=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
TREE=$(CDPATH='' cd -- "$HERE/.." && pwd)
MANIFEST="$HERE/manifest.txt"
CACHE="${ASURA_VENDOR_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/toren/vendor}/desktop"

URL="https://github.com/end-4/dots-hyprland.git"
COMMIT="21733c3ca7f4edc20faafedea844dc96002c6f7f"
UPSTREAM_PATH="dots/.config/quickshell/ii"

# end-4 pulls its rounded-polygon library in as a git submodule at
# modules/common/widgets/shapes; it is vendored here at the commit end-4 pins,
# without its LICENSE (Apache-2.0, kept as toren/LICENSE.apache-2.0 because
# every file named LICENSE must be the root's GPL text), README or examples.
# end-4's matugen templates (dots/.config/matugen) turn a wallpaper into the
# desktop's palette. They are vendored at matugen/; the config.toml there is
# ASURA's overlay, which keeps the templates ASURA can use.
MATUGEN_PATH="dots/.config/matugen"
MATUGEN_DEST="matugen"

SHAPES_URL="https://github.com/end-4/rounded-polygon-qmljs.git"
SHAPES_DEST="modules/common/widgets/shapes"

# ASURA's files at upstream's paths. shell.qml is upstream's entry point with
# the ASURA bridge (toren/TorenBridge.qml) added, so the desktop's keybindings
# reach this shell exactly as they reach ASURA's own. modules/ii/dock/Dock.qml
# is ASURA's island dock: only the apps, floating above the edge, always
# shown, no pin or overview buttons, no window-count marks under its icons
# (DockAppButton.qml), auto-hiding by default. The Settings pages
# (modules/settings/{Interface,General,Services,Bar}Config.qml) offer what ASURA's
# desktop uses: the dock's Auto-hide switch, no AI/anime/translator options (the
# left sidebar is not loaded), no utility buttons, no unused workspace options. modules/ii/bar/Workspaces.qml shows only the
# workspaces that exist (the active one and those with windows), no empty slots.
# modules/ii/bar/BarContent.qml lays the bar out ASURA's way (workspaces at the
# start, the clock alone in the centre, resources at the end, no window title),
# and Resources.qml/Resource.qml are the labelled CPU/RAM/SWAP readouts;
# BatteryIndicator.qml is the battery in the same style, at the bar's end;
# BarGroup.qml gives the bar's groups a small fixed radius.
# services/GlobalFocusGrab.qml keeps the bar in the focus grab, so a sidebar
# opened from the bar is not closed by the grab at once.
# panelFamilies/IllogicalImpulseFamily.qml loads end-4's panels without the
# left sidebar (AI chat, anime). welcome.qml and
# modules/common/widgets/LightDarkPreferenceButton.qml show the ASURA desktop
# and links instead of upstream's previews and third-party links.
OVERLAYS="
shell.qml
matugen/config.toml
matugen/templates/colors.json
modules/ii/dock/Dock.qml
modules/ii/dock/DockAppButton.qml
modules/ii/bar/Workspaces.qml
modules/ii/bar/BarContent.qml
modules/ii/bar/Resources.qml
modules/ii/bar/Resource.qml
modules/ii/bar/BatteryIndicator.qml
modules/ii/bar/BarGroup.qml
services/GlobalFocusGrab.qml
panelFamilies/IllogicalImpulseFamily.qml
welcome.qml
modules/common/widgets/LightDarkPreferenceButton.qml
GlobalStates.qml
modules/ii/bar/ClockWidget.qml
modules/ii/sidebarRight/SidebarRightContent.qml
modules/ii/verticalBar/VerticalClockWidget.qml
modules/ii/overview/Overview.qml
modules/ii/overview/SearchBar.qml
modules/ii/overview/SearchWidget.qml
modules/ii/bar/StyledPopup.qml
modules/ii/bar/StyledPopupHeaderRow.qml
modules/ii/bar/StyledPopupValueRow.qml
modules/ii/bar/BatteryPopup.qml
modules/ii/bar/ResourcesPopup.qml
modules/ii/background/Background.qml
modules/common/Config.qml
scripts/colors/applycolor.sh
scripts/colors/switchwall.sh
modules/ii/bar/Media.qml
services/MprisController.qml
assets/images/default_wallpaper.png
modules/common/widgets/GroupButton.qml
modules/common/widgets/StyledSlider.qml
modules/ii/overview/SearchItem.qml
modules/ii/sidebarRight/QuickSliders.qml
modules/ii/sidebarRight/notifications/NotificationList.qml
modules/ii/sidebarRight/notifications/NotificationStatusButton.qml
modules/ii/sidebarRight/quickToggles/AndroidQuickPanel.qml
modules/ii/sidebarRight/quickToggles/androidStyle/AndroidQuickToggleButton.qml
modules/ii/sidebarRight/quickToggles/classicStyle/QuickToggleButton.qml
modules/waffle/background/WaffleBackground.qml
services/AppSearch.qml
services/LauncherSearch.qml
modules/common/functions/Session.qml
modules/common/widgets/AppIcon.qml
modules/ii/sessionScreen/SessionScreen.qml
modules/waffle/looks/WAppIcon.qml
modules/waffle/actionCenter/mainPage/MainPageFooter.qml
modules/waffle/bar/StartButton.qml
matugen/templates/gtk-3.0/gtk.css
matugen/templates/gtk-4.0/gtk.css
modules/ii/overview/OverviewWidget.qml
modules/ii/sidebarLeft/anime/BooruImage.qml
modules/waffle/taskView/TaskViewContent.qml
modules/waffle/taskView/TaskViewWindow.qml
modules/ii/cheatsheet/Cheatsheet.qml
services/HyprlandKeybinds.qml
modules/common/Appearance.qml
modules/common/widgets/FloatingActionButton.qml
services/Cliphist.qml
services/Notifications.qml
services/Translation.qml
modules/ii/sidebarRight/calendar/CalendarWidget.qml
modules/ii/sidebarRight/pomodoro/PomodoroWidget.qml
services/TimerService.qml
modules/ii/regionSelector/OptionsToolbar.qml
"

# Upstream's paths Toren leaves out: the desktop's own Settings app (Toren has
# one settings window, the Hub, toren-shell hub open, which carries those
# options on its Desktop page) and the cheatsheet's periodic table page
# (modules/ii/cheatsheet/Cheatsheet.qml shows the keybindings only). A path
# ending in / leaves out the directory.
REMOVED="
settings.qml
modules/settings/
modules/ii/cheatsheet/CheatsheetPeriodicTable.qml
modules/ii/cheatsheet/ElementTile.qml
modules/ii/cheatsheet/periodic_table.js
modules/ii/sidebarRight/pomodoro/PomodoroTimer.qml
"

log() { printf '%s\n' "$*" >&2; }

is_removed() {
	for r in $REMOVED; do
		case "$r" in
			*/) case "$1" in "$r"*) return 0 ;; esac ;;
			*) [ "$r" = "$1" ] && return 0 ;;
		esac
	done
	return 1
}

is_overlay() {
	for o in $OVERLAYS; do [ "$o" = "$1" ] && return 0; done
	return 1
}

fetch() {
	# One shallow fetch of the pinned commit, blobs included: a partial clone
	# would download the tree's files one by one.
	if [ ! -d "$CACHE/.git" ]; then
		mkdir -p "$CACHE"
		git -C "$CACHE" init -q
		git -C "$CACHE" remote add origin "$URL"
	fi
	git -C "$CACHE" cat-file -e "$COMMIT^{commit}" 2>/dev/null ||
		git -C "$CACHE" fetch -q --depth 1 origin "$COMMIT"
	git -C "$CACHE" cat-file -e "$COMMIT^{commit}"
}

cmd_sync() {
	fetch
	list=$(mktemp)
	git -C "$CACHE" ls-tree -r --format='%(objectmode) %(objectname) %(path)' "$COMMIT" "$UPSTREAM_PATH" "$MATUGEN_PATH" |
	while read -r mode blob path; do
		[ "$mode" = 160000 ] && continue
		case "$path" in
			"$MATUGEN_PATH"/*) rel="$MATUGEN_DEST/${path#"$MATUGEN_PATH"/}" ;;
			*) rel=${path#"$UPSTREAM_PATH"/} ;;
		esac
		is_overlay "$rel" && continue
		is_removed "$rel" && continue
		mkdir -p "$TREE/$(dirname "$rel")"
		git -C "$CACHE" cat-file blob "$blob" >"$TREE/$rel"
		[ "$mode" = 100755 ] && chmod 755 "$TREE/$rel"
		printf '%s\n' "$rel"
	done | sort >"$list"

	{
		echo "# end-4/dots-hyprland $COMMIT $UPSTREAM_PATH"
		echo "# Generated by shell-desktop/toren/vendor.sh sync. Do not edit."
		while read -r rel; do
			printf '%s %s\n' "$(git hash-object "$TREE/$rel")" "$rel"
		done <"$list"
	} >"$MANIFEST"
	rm -f "$list"

	# The submodule, at the commit end-4's tree points to.
	shapes_commit=$(git -C "$CACHE" ls-tree "$COMMIT" "$UPSTREAM_PATH/$SHAPES_DEST" | awk '$1 == "160000" {print $3}')
	[ -n "$shapes_commit" ] || { log "sync: no shapes submodule at $SHAPES_DEST"; exit 1; }
	SC="$CACHE-shapes"
	if [ ! -d "$SC/.git" ]; then
		mkdir -p "$SC"; git -C "$SC" init -q; git -C "$SC" remote add origin "$SHAPES_URL"
	fi
	git -C "$SC" cat-file -e "$shapes_commit^{commit}" 2>/dev/null || git -C "$SC" fetch -q --depth 1 origin "$shapes_commit"
	git -C "$SC" ls-tree -r --format='%(objectname) %(path)' "$shapes_commit" |
	while read -r blob path; do
		case "$path" in LICENSE|README.md|.gitignore|example*.qml) continue ;; esac
		mkdir -p "$TREE/$SHAPES_DEST/$(dirname "$path")"
		git -C "$SC" cat-file blob "$blob" >"$TREE/$SHAPES_DEST/$path"
		printf '%s %s\n' "$blob" "$SHAPES_DEST/$path"
	done >>"$MANIFEST"
	log "synced $(grep -vc '^#' "$MANIFEST") files at $COMMIT"
}

cmd_verify() {
	[ -f "$MANIFEST" ] || { log "no manifest"; return 1; }
	pinned=$(sed -n '1s/^# end-4\/dots-hyprland \([0-9a-f]*\) .*/\1/p' "$MANIFEST")
	bad=0
	if [ "$pinned" != "$COMMIT" ]; then
		log "PIN      vendor.sh has $COMMIT, manifest.txt has $pinned"
		bad=$((bad + 1))
	fi
	total=0
	while read -r blob rel _mark; do
		total=$((total + 1))
		if [ ! -f "$TREE/$rel" ]; then
			log "MISSING  $rel"; bad=$((bad + 1)); continue
		fi
		[ "$(git hash-object "$TREE/$rel")" = "$blob" ] || { log "MODIFIED $rel"; bad=$((bad + 1)); }
	done <<LIST
$(grep -v '^#' "$MANIFEST")
LIST
	for o in $OVERLAYS; do
		[ -f "$TREE/$o" ] || { log "MISSING  overlay $o"; bad=$((bad + 1)); }
	done
	for r in $REMOVED; do
		[ -e "$TREE/${r%/}" ] && { log "REMOVED  $r is left out of Toren but is there"; bad=$((bad + 1)); }
	done
	if [ "$bad" -eq 0 ]; then
		log "ok - $total desktop files are byte-identical to $COMMIT"
		return 0
	fi
	log "FAIL - $bad problem(s)"
	return 1
}

case "${1:-}" in
	sync) cmd_sync ;;
	verify) cmd_verify ;;
	update)
		new=${2:-}
		case "$new" in *[!0-9a-f]*|"") log "update: needs a full 40-character commit id"; exit 2 ;; esac
		[ "${#new}" -eq 40 ] || { log "update: needs a full 40-character commit id"; exit 2; }
		sed "s/^COMMIT=\"[0-9a-f]*\"\$/COMMIT=\"$new\"/" "$0" >"$0.tmp" && cat "$0.tmp" >"$0" && rm -f "$0.tmp"
		exec sh "$0" sync
		;;
	*) log "usage: vendor.sh sync|verify|update <commit>"; exit 2 ;;
esac
