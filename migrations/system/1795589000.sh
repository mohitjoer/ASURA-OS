#!/bin/sh
echo "Point xbps and git update remotes at mohitjoer/Toren-Linux"

ROOT="${ROOT:-}"
OLD_REPO="repository=https://mohitjoer.github.io/ASURA-OS/current"
NEW_REPO="repository=https://mohitjoer.github.io/Toren-Linux/current"

for xbps_dir in "$ROOT/etc/xbps.d" "$ROOT/usr/share/xbps.d"; do
	[ -d "$xbps_dir" ] || continue
	for f in "$xbps_dir"/*.conf; do
		[ -f "$f" ] || continue
		if grep -qxF "$OLD_REPO" "$f" 2>/dev/null; then
			tmp=$(mktemp "${f}.tmp.XXXXXX") || continue
			awk -v old="$OLD_REPO" -v new="$NEW_REPO" '{ if ($0 == old) print new; else print }' "$f" >"$tmp" && mv -f "$tmp" "$f" || rm -f "$tmp"
		fi
	done
done

rewrite_git_remote() {
	dir="$1"
	[ -n "$dir" ] && [ -d "$dir/.git" ] || return 0
	command -v git >/dev/null 2>&1 || return 0
	url=$(git -C "$dir" remote get-url origin 2>/dev/null || true)
	case "$url" in
		https://github.com/mohitjoer/ASURA-OS|https://github.com/mohitjoer/ASURA-OS.git|https://github.com/mohitjoer/Asura-OS|https://github.com/mohitjoer/Asura-OS.git)
			git -C "$dir" remote set-url origin https://github.com/mohitjoer/Toren-Linux.git 2>/dev/null || true
			;;
		git@github.com:mohitjoer/ASURA-OS.git|git@github.com:mohitjoer/Asura-OS.git)
			git -C "$dir" remote set-url origin git@github.com:mohitjoer/Toren-Linux.git 2>/dev/null || true
			;;
	esac
}

for d in "$ROOT/var/cache/asura/source" "$ROOT/usr/src/asura" "$ROOT/usr/share/asura" "$ROOT"/home/*/.cache/asura/source; do
	rewrite_git_remote "$d"
done

if [ -f "$ROOT/usr/share/asura/applied-source" ]; then
	recorded=$(head -n 1 "$ROOT/usr/share/asura/applied-source" 2>/dev/null || true)
	if [ -n "$recorded" ]; then
		case "$recorded" in
			/*) rewrite_git_remote "$ROOT$recorded" ;;
		esac
	fi
fi

exit 0
