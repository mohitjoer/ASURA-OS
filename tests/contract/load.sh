#!/bin/sh
# Load a Quickshell tree in a sandbox and fail on module-resolution errors.
#
#   load.sh [TREE]    TREE is anything `quickshell -p` takes; default is shell/
#
# This is the check that catches an empty bar: a tree that imports qs.Services but
# does not ship Services/. Point it at an extracted package to test the package.
#
# The sandbox is env -i with a throwaway HOME and XDG_* and the offscreen Qt platform:
# no Wayland socket and no D-Bus session, so it cannot touch the running desktop. It
# is the one place these tests launch quickshell, and only in this form.
#
# Offscreen has no layer-shell, so a healthy tree ends with `No PanelWindow backend
# loaded` at its first window. That is expected and is not a failure: it means every
# import and type resolved. Proven: the packaged tree logs only that line; a copy with
# Services/ deleted logs `module "qs.Services" is not installed`.
tree=${1:-$(dirname "$0")/../../shell}
command -v quickshell >/dev/null || { echo "load: skip, no quickshell"; exit 0; }
t=$(mktemp -d) && trap 'rm -rf "$t"' EXIT
mkdir -m 700 "$t/run"
env -i PATH="$PATH" LANG=C.UTF-8 HOME="$t" XDG_RUNTIME_DIR="$t/run" XDG_CONFIG_HOME="$t/c" XDG_CACHE_HOME="$t/k" \
    XDG_STATE_HOME="$t/s" XDG_DATA_HOME="$t/d" QT_QPA_PLATFORM=offscreen \
    timeout -k 2 10 quickshell -p "$tree" >"$t/log" 2>&1 || true
# A type that fails to compile (a property the element does not have, a syntax
# error) makes every type that uses it "unavailable" and the whole config fails
# to load: the end-4 desktop's "Quickshell: Reload failed" was one wrong
# property name in one file.
#
# Offscreen has no layer-shell, so a PanelWindow produces "No PanelWindow backend
# loaded" which cascades into "Type ... unavailable" and "Failed to load configuration".
# We allow that expected offscreen cascade while catching real syntax/property errors.
bad=$(grep -E 'is not installed|is not a type|unresolvable import|ReferenceError|non-existent property|Cannot assign to|Syntax ?Error|Expected token' "$t/log" | sort -u | head -20)
if [ -z "$bad" ] && ! grep -qF 'No PanelWindow backend loaded.' "$t/log"; then
	bad=$(grep -E 'Failed to load configuration|Type [A-Za-z_.]+ unavailable' "$t/log" | sort -u | head -20)
fi
[ -z "$bad" ] || { echo "$bad"; echo "load: $tree does not resolve"; exit 1; }
echo "load: ok, $tree resolves"
