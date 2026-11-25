#!/bin/sh
# Runs toren-perf-apply against throwaway config and cache dirs and checks which
# hyprctl calls it would make. Needs no compositor: --dry-run only prints them.
set -eu
here=$(cd "$(dirname "$0")" && pwd)
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/cfg/toren" "$tmp/cache/toren"

run() { XDG_CONFIG_HOME=$tmp/cfg XDG_CACHE_HOME=$tmp/cache sh "$here/toren-perf-apply" --dry-run; }
check() { [ "$(run)" = "$2" ] || { echo "FAIL $1: got '$(run)'" >&2; exit 1; }; }

check "nothing set changes nothing" ""
printf '{"disableBlur": true}' >"$tmp/cfg/toren/performance.json"
check "disableBlur" "hyprctl eval 'hl.config({ decoration = { blur = { enabled = false } } })'"
printf '{"lowPowerMode": true}' >"$tmp/cfg/toren/performance.json"
check "lowPowerMode" "hyprctl eval 'hl.config({ decoration = { blur = { enabled = false } } })'
hyprctl eval 'hl.config({ decoration = { shadow = { enabled = false } } })'"
printf '{}' >"$tmp/cfg/toren/performance.json"
printf '{"saver": true}' >"$tmp/cache/toren/hypr-perf.json"
check "power saver" "hyprctl eval 'hl.config({ decoration = { blur = { enabled = false } } })'
hyprctl eval 'hl.config({ decoration = { shadow = { enabled = false } } })'"
echo "ok"
