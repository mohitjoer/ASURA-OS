#!/bin/sh
echo "Give each user back the update revisions an earlier sudo asura update left owned by root"

# Before the fix, `sudo asura update` preserved revisions as root inside the
# user's ~/.local/state/asura/revisions, and the user's own update could not
# remove them ("permission denied"). Hand each one to the owner of the home
# it is in, so the user's update keeps and prunes them normally.
for home in /home/*; do
	revisions="$home/.local/state/asura/revisions"
	# Running as root in a folder the user owns: follow no symlink they made.
	[ -d "$revisions" ] && [ ! -L "$revisions" ] || continue
	owner=$(stat -c '%u:%g' "$home") || continue
	[ "${owner%%:*}" != 0 ] || continue
	for revision in "$revisions"/*; do
		[ -e "$revision" ] && [ ! -L "$revision" ] || continue
		if [ -n "$(find "$revision" -user 0 -print -quit)" ]; then
			chown -R -h -P "$owner" "$revision"
			echo "gave $revision back to $owner"
		fi
	done
done
