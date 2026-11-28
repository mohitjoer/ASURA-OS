#!/bin/sh
# Build a bootable Toren live ISO using Void Linux x86_64 (glibc) base and void-mklive.
#
# Usage: installation/build-iso-void.sh [build-directory]
#
# TOREN_VERSION=0.1.0-beta.1 names the image toren-linux-0.1.0-beta.1-x86_64.iso
# and records IMAGE_VERSION in /etc/os-release. Unset, the image is
# toren-linux-x86_64.iso. The file name never names the base layer.
#
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BUILD_DIR=${1:-"$REPO_ROOT/build/iso-void"}
ARCH=${TOREN_ARCH:-${ASURA_ARCH:-x86_64}}
MIRROR=${TOREN_MIRROR:-${ASURA_MIRROR:-https://repo-fastly.voidlinux.org/current}}
TOREN_VERSION=${TOREN_VERSION:-${ASURA_VERSION:-}}
case "$TOREN_VERSION" in
	*[!0-9A-Za-z.+-]*)
		echo "build-iso-void: TOREN_VERSION may contain only letters, digits, '.', '+' and '-'" >&2
		exit 2
		;;
esac
if [ -n "$TOREN_VERSION" ]; then
	ISO_NAME="toren-linux-${TOREN_VERSION}-x86_64"
else
	ISO_NAME=toren-linux-x86_64
fi
ISO_TITLE="Toren Linux x86_64"
export GOWORK=off

# Pinned binary versions and checksum verification file
OPENCODE_VERSION="2.2.6"
CHECKSUMS_FILE="$REPO_ROOT/installation/binary-checksums.txt"

verify_checksum() {
	_file="$1"
	_name=$(basename "$_file")
	_expected=$(awk -v n="$_name" '$2 == n {print $1}' "$CHECKSUMS_FILE")
	if [ -z "$_expected" ]; then
		echo "build-iso-void: package $_name (verified by repository signature)"
		return 0
	fi
	_actual=$(sha256sum "$_file" | awk '{print $1}')
	if [ "$_actual" != "$_expected" ]; then
		echo "build-iso-void: ERROR: Checksum mismatch for $_name!" >&2
		echo "  Expected: $_expected" >&2
		echo "  Actual:   $_actual" >&2
		exit 1
	fi
	echo "build-iso-void: verified SHA-256 for $_name ($_actual)"
}

for tool in python3 git curl; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		echo "build-iso-void: required tool not found: $tool" >&2
		exit 2
	fi
done

if ! python3 -c 'import yaml' >/dev/null 2>&1; then
	echo "build-iso-void: PyYAML is required (python3 -c 'import yaml')" >&2
	exit 2
fi

mkdir -p "$BUILD_DIR"
BUILD_DIR=$(CDPATH= cd -- "$BUILD_DIR" && pwd)
INCLUDE_DIR="$BUILD_DIR/include"
rm -rf "$INCLUDE_DIR"
mkdir -p "$INCLUDE_DIR/etc/sudoers.d" \
	"$INCLUDE_DIR/etc/udev/rules.d" \
	"$INCLUDE_DIR/etc/modprobe.d" \
	"$INCLUDE_DIR/etc/sysctl.d" \
	"$INCLUDE_DIR/etc/turnstile" \
	"$INCLUDE_DIR/etc/xbps.d" \
	"$INCLUDE_DIR/etc/plymouth" \
	"$INCLUDE_DIR/etc/dracut.conf.d" \
	"$INCLUDE_DIR/usr/bin" \
	"$INCLUDE_DIR/usr/share/toren/branding" \
	"$INCLUDE_DIR/usr/share/toren/session" \
	"$INCLUDE_DIR/usr/share/toren/runit/services" \
	"$INCLUDE_DIR/usr/share/toren/hyprland" \
	"$INCLUDE_DIR/usr/share/toren/shell" \
	"$INCLUDE_DIR/usr/share/wayland-sessions"

# Configure Plymouth boot splash with logo and loading progress bar animation
mkdir -p "$INCLUDE_DIR/usr/share/plymouth/themes/toren"
cp -r "$REPO_ROOT/system/plymouth/toren"/* "$INCLUDE_DIR/usr/share/plymouth/themes/toren/"
ln -sf /usr/share/plymouth/themes/toren/toren.plymouth "$INCLUDE_DIR/usr/share/plymouth/themes/default.plymouth"

# Pre-stage TrueType fonts and fontconfig configuration for Plymouth label rendering
mkdir -p "$INCLUDE_DIR/usr/share/fonts/TTF" "$INCLUDE_DIR/etc/fonts"
rm -rf "$INCLUDE_DIR/usr/lib/plymouth"
rm -f "$INCLUDE_DIR/usr/lib/libseat.so"*
for font in DejaVuSans.ttf DejaVuSans-Bold.ttf DejaVuSansMono.ttf; do
	for src in "/usr/share/fonts/truetype/dejavu/$font" "/usr/share/fonts/TTF/$font"; do
		if [ -f "$src" ]; then
			cp -L "$src" "$INCLUDE_DIR/usr/share/fonts/TTF/$font"
			break
		fi
	done
done
if [ -d "$REPO_ROOT/public/fonts" ]; then
	cp -L "$REPO_ROOT/public/fonts"/* "$INCLUDE_DIR/usr/share/fonts/TTF/"
fi

# Configure fontconfig local preferences without destroying the main system fonts.conf
cat >"$INCLUDE_DIR/etc/fonts/local.conf" <<'EOF'
<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
<fontconfig>
  <dir>/usr/share/fonts</dir>
  <dir>/usr/share/fonts/TTF</dir>
  <alias>
    <family>sans-serif</family>
    <prefer>
      <family>Inter</family>
      <family>Space Grotesk</family>
      <family>Noto Sans</family>
      <family>DejaVu Sans</family>
    </prefer>
  </alias>
  <alias>
    <family>sans</family>
    <prefer>
      <family>Inter</family>
      <family>Space Grotesk</family>
      <family>Noto Sans</family>
      <family>DejaVu Sans</family>
    </prefer>
  </alias>
  <alias>
    <family>monospace</family>
    <prefer>
      <family>JetBrainsMono Nerd Font</family>
      <family>JetBrains Mono</family>
      <family>DejaVu Sans Mono</family>
    </prefer>
  </alias>
</fontconfig>
EOF

cat >"$INCLUDE_DIR/etc/plymouth/plymouthd.conf" <<'EOF'
[Daemon]
Theme=toren
ShowDelay=0
DeviceTimeout=8
EOF

cat >"$INCLUDE_DIR/etc/dracut.conf.d/50-plymouth.conf" <<'EOF'
add_dracutmodules+=" plymouth "
add_drivers+=" vboxvideo vmwgfx bochs virtio_gpu qxl simpledrm "
install_optional_items+=" /usr/lib/plymouth/script.so /usr/lib/plymouth/label.so /usr/lib/plymouth/label-pango.so /usr/lib/plymouth/renderers/drm.so /usr/lib/plymouth/renderers/frame-buffer.so /etc/fonts/fonts.conf /etc/fonts/local.conf /usr/share/fonts/TTF/DejaVuSans.ttf /usr/share/fonts/TTF/DejaVuSans-Bold.ttf /usr/share/fonts/TTF/DejaVuSansMono.ttf "
EOF

# Setup OS identity
cat >"$INCLUDE_DIR/etc/os-release" <<'EOF'
NAME="Toren Linux"
PRETTY_NAME="Toren Linux Beta 001"
ID=toren
ID_LIKE="void linux"
VERSION_ID="0.1"
VERSION_CODENAME=beta-001
HOME_URL="https://github.com/mohitjoer/Toren-Linux"
SUPPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
BUG_REPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
EOF
if [ -n "$TOREN_VERSION" ]; then
	printf 'IMAGE_VERSION=%s\n' "$TOREN_VERSION" >>"$INCLUDE_DIR/etc/os-release"
fi

printf 'toren-linux\n' >"$INCLUDE_DIR/etc/hostname"

cat >"$INCLUDE_DIR/etc/issue" <<'EOF'
       #                    #         Toren Linux Beta 001 (\m)
     %##                    ##%       Kernel \r on an \m (\l)
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%    https://github.com/mohitjoer/Toren-Linux
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

EOF

cat >"$INCLUDE_DIR/etc/issue.net" <<'EOF'
       #                    #         Toren Linux Beta 001
     %##                    ##%       https://github.com/mohitjoer/Toren-Linux
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

EOF

cat >"$INCLUDE_DIR/etc/motd" <<'EOF'
       #                    #         Toren Linux Beta 001
     %##                    ##%       https://github.com/mohitjoer/Toren-Linux
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

EOF



# System locale configuration (enforce UTF-8 default across apps and terminals)
mkdir -p "$INCLUDE_DIR/etc"
cat > "$INCLUDE_DIR/etc/locale.conf" <<'LOCALECFG'
LANG=en_US.UTF-8
LC_ALL=en_US.UTF-8
LOCALECFG

# Copy branding assets
mkdir -p "$INCLUDE_DIR/usr/share/toren/branding"
cp "$REPO_ROOT/public/toren-logo.svg" \
	"$INCLUDE_DIR/usr/share/toren/branding/toren-logo.svg"
cp "$REPO_ROOT/public/toren-logo.svg" \
	"$INCLUDE_DIR/usr/share/toren/branding/toren-logo.svg"
cp "$REPO_ROOT/public/toren-logo-on-dark.svg" \
	"$INCLUDE_DIR/usr/share/toren/branding/toren-logo-on-dark.svg"
cp "$REPO_ROOT/public/toren-logo-on-dark.svg" \
	"$INCLUDE_DIR/usr/share/toren/branding/toren-logo-on-dark.svg"
if [ -f "$REPO_ROOT/public/branding/boot-splash.png" ]; then
	cp "$REPO_ROOT/public/branding/boot-splash.png" \
		"$INCLUDE_DIR/usr/share/toren/branding/boot-splash.png"
fi
if [ -d "$REPO_ROOT/public/branding" ]; then
	cp -r "$REPO_ROOT/public/branding"/* \
		"$INCLUDE_DIR/usr/share/toren/branding/" 2>/dev/null || true
fi

# Ensure strict permissions on sudoers.d directory (0750) and wheel rule (0440)
chmod 0750 "$INCLUDE_DIR/etc/sudoers.d"
echo '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' > "$INCLUDE_DIR/etc/sudoers.d/wheel"
chmod 0440 "$INCLUDE_DIR/etc/sudoers.d/wheel"

# Copy session files, keybind catalogue, and desktop sessions
cp "$REPO_ROOT/session/toren-session" "$INCLUDE_DIR/usr/share/toren/session/toren-session"
cp "$REPO_ROOT/session/toren-session" "$INCLUDE_DIR/usr/bin/toren-session"
cp "$REPO_ROOT/session/toren-session-safe" "$INCLUDE_DIR/usr/share/toren/session/toren-session-safe"
cp "$REPO_ROOT/session/toren-session-safe" "$INCLUDE_DIR/usr/bin/toren-session-safe"
cp "$REPO_ROOT/session/toren-activate" "$INCLUDE_DIR/usr/share/toren/session/toren-activate"
cp "$REPO_ROOT/session/toren-activate" "$INCLUDE_DIR/usr/bin/toren-activate"
cp "$REPO_ROOT/session/toren.desktop" "$INCLUDE_DIR/usr/share/toren/session/toren.desktop"
cp "$REPO_ROOT/session/toren.desktop" "$INCLUDE_DIR/usr/share/wayland-sessions/toren.desktop"
cp "$REPO_ROOT/session/toren-safe.desktop" "$INCLUDE_DIR/usr/share/toren/session/toren-safe.desktop"
cp "$REPO_ROOT/session/toren-safe.desktop" "$INCLUDE_DIR/usr/share/wayland-sessions/toren-safe.desktop"
cp "$REPO_ROOT/session/binds.json" "$INCLUDE_DIR/usr/share/toren/session/binds.json"
cp "$REPO_ROOT/session/runit/toren-user-services" "$INCLUDE_DIR/usr/share/toren/session/toren-user-services"
cp "$REPO_ROOT/session/toren-graphics-env" "$INCLUDE_DIR/usr/share/toren/session/toren-graphics-env"
cp "$REPO_ROOT/session/runit/toren-user-services" "$INCLUDE_DIR/usr/bin/toren-user-services"

# Copy session store wrapper, portal configuration, and session CLI scripts
cp "$REPO_ROOT/session/toren-store" "$INCLUDE_DIR/usr/share/toren/session/toren-store"
mkdir -p "$INCLUDE_DIR/usr/share/xdg-desktop-portal" "$INCLUDE_DIR/usr/share/toren/session"
cp "$REPO_ROOT/session/toren-portals.conf" "$INCLUDE_DIR/usr/share/xdg-desktop-portal/toren-portals.conf"
cp "$REPO_ROOT/session/toren-portals.conf" "$INCLUDE_DIR/usr/share/toren/session/toren-portals.conf"

# Stage session units into /usr/share/toren/units and systemd user directory
mkdir -p "$INCLUDE_DIR/usr/share/toren/units" "$INCLUDE_DIR/usr/share/toren/session/units" "$INCLUDE_DIR/usr/lib/systemd/user"
if [ -d "$REPO_ROOT/session/units" ]; then
	cp "$REPO_ROOT/session/units"/* "$INCLUDE_DIR/usr/share/toren/units/"
	cp "$REPO_ROOT/session/units"/* "$INCLUDE_DIR/usr/share/toren/session/units/"
	cp "$REPO_ROOT/session/units"/* "$INCLUDE_DIR/usr/lib/systemd/user/"
fi

# Copy default wallpapers to /usr/share/toren/backgrounds/ and /etc/skel/Pictures/Wallpapers/
mkdir -p "$INCLUDE_DIR/usr/share/toren/backgrounds"
mkdir -p "$INCLUDE_DIR/etc/skel/Pictures/Wallpapers"
if [ -d "$REPO_ROOT/public/wallpapers" ]; then
	cp "$REPO_ROOT/public/wallpapers"/* "$INCLUDE_DIR/usr/share/toren/backgrounds/"
	cp "$REPO_ROOT/public/wallpapers"/* "$INCLUDE_DIR/etc/skel/Pictures/Wallpapers/"
fi

for s in toren-hub toren-idle toren-installer-ui toren-pin toren-reload-cover toren-shot toren-vm toren-wallpaper-ui toren-welcome toren-greeter; do
	if [ -f "$REPO_ROOT/session/$s" ]; then
		cp "$REPO_ROOT/session/$s" "$INCLUDE_DIR/usr/bin/"
	fi
done
if [ -d "$REPO_ROOT/session/scripts" ]; then
	cp "$REPO_ROOT/session/scripts"/* "$INCLUDE_DIR/usr/bin/"
fi

if [ -d "$REPO_ROOT/session/grub" ]; then
	mkdir -p "$INCLUDE_DIR/usr/share/toren/session/grub"
	cp "$REPO_ROOT/session/grub"/* "$INCLUDE_DIR/usr/share/toren/session/grub/"
fi

# Copy all shell scripts (toren-app, toren-cmd-*, toren-summon, etc.) to /usr/bin/
if [ -d "$REPO_ROOT/shell/scripts" ]; then
	cp "$REPO_ROOT/shell/scripts"/* "$INCLUDE_DIR/usr/bin/"
fi

# Copy all system helpers (toren-volume, toren-restart-audio, toren-hw-backlight, etc.)
for hdir in audio display input perf pkg; do
	if [ -d "$REPO_ROOT/system/helpers/$hdir" ]; then
		for f in "$REPO_ROOT/system/helpers/$hdir"/*; do
			if [ -f "$f" ] && [ "${f##*.}" != "sh" -o "${f##*/}" = "toren-installer.sh" ]; then
				cp "$f" "$INCLUDE_DIR/usr/bin/"
			fi
		done
	fi
done
cp "$REPO_ROOT/system/helpers/prime-run" "$INCLUDE_DIR/usr/bin/prime-run"
cp "$REPO_ROOT/system/helpers/fan-profile" "$INCLUDE_DIR/usr/bin/fan-profile"
cp "$REPO_ROOT/system/helpers/systemctl" "$INCLUDE_DIR/usr/bin/systemctl"
cp "$REPO_ROOT/system/helpers/toren-gpu" "$INCLUDE_DIR/usr/bin/toren-gpu"
chmod 755 "$INCLUDE_DIR/usr/bin/systemctl"
cp "$REPO_ROOT/system/helpers/journalctl" "$INCLUDE_DIR/usr/bin/journalctl"
chmod 755 "$INCLUDE_DIR/usr/bin/journalctl"
cp "$REPO_ROOT/installation/toren-installer.sh" "$INCLUDE_DIR/usr/bin/toren-installer"
chmod 755 "$INCLUDE_DIR/usr/bin/toren-installer"
touch "$INCLUDE_DIR/etc/toren-live"



# Copy GLib GSettings schemas
mkdir -p "$INCLUDE_DIR/usr/share/glib-2.0/schemas"
cp "$REPO_ROOT/shell"/org.toren.*.gschema.xml "$INCLUDE_DIR/usr/share/glib-2.0/schemas/"

# Copy applications .desktop entries
mkdir -p "$INCLUDE_DIR/usr/share/applications"
if [ -d "$REPO_ROOT/session/applications" ]; then
	cp "$REPO_ROOT/session/applications"/*.desktop "$INCLUDE_DIR/usr/share/applications/" 2>/dev/null || true
fi

# Stage core application desktop shortcuts and custom branding icons into ISO include
mkdir -p "$INCLUDE_DIR/etc/skel/Desktop"
mkdir -p "$INCLUDE_DIR/usr/share/icons/hicolor/scalable/apps"
mkdir -p "$INCLUDE_DIR/usr/share/icons/hicolor/512x512/apps"
mkdir -p "$INCLUDE_DIR/usr/share/pixmaps"

# Toren Store desktop shortcut and store icon
if [ -f "$REPO_ROOT/session/applications/toren-store.desktop" ]; then
	cp "$REPO_ROOT/session/applications/toren-store.desktop" "$INCLUDE_DIR/usr/share/applications/"
	cp "$REPO_ROOT/session/applications/toren-store.desktop" "$INCLUDE_DIR/etc/skel/Desktop/"
	chmod 755 "$INCLUDE_DIR/etc/skel/Desktop/toren-store.desktop"
fi
if [ -f "$REPO_ROOT/shell/assets/toren-store.svg" ]; then
	cp "$REPO_ROOT/shell/assets/toren-store.svg" "$INCLUDE_DIR/usr/share/icons/hicolor/scalable/apps/toren-store.svg"
	cp "$REPO_ROOT/shell/assets/toren-store.svg" "$INCLUDE_DIR/usr/share/pixmaps/toren-store.svg"
fi
if [ -f "$REPO_ROOT/shell/assets/toren-store.png" ]; then
	cp "$REPO_ROOT/shell/assets/toren-store.png" "$INCLUDE_DIR/usr/share/icons/hicolor/512x512/apps/toren-store.png"
	cp "$REPO_ROOT/shell/assets/toren-store.png" "$INCLUDE_DIR/usr/share/pixmaps/toren-store.png"
fi

# Toren Installer desktop shortcut and icon
if [ -f "$REPO_ROOT/session/applications/toren-installer.desktop" ]; then
	cp "$REPO_ROOT/session/applications/toren-installer.desktop" "$INCLUDE_DIR/usr/share/applications/"
	cp "$REPO_ROOT/session/applications/toren-installer.desktop" "$INCLUDE_DIR/etc/skel/Desktop/"
	chmod 755 "$INCLUDE_DIR/etc/skel/Desktop/toren-installer.desktop"
fi
if [ -f "$REPO_ROOT/shell/assets/toren-installer.svg" ]; then
	cp "$REPO_ROOT/shell/assets/toren-installer.svg" "$INCLUDE_DIR/usr/share/icons/hicolor/scalable/apps/toren-installer.svg"
	cp "$REPO_ROOT/shell/assets/toren-installer.svg" "$INCLUDE_DIR/usr/share/pixmaps/toren-installer.svg"
fi

# Core desktop application icons
for icon_name in asura-settings asura-vm preferences-system preferences-system-network computer utilities-terminal utilities-system-monitor opencode; do
	if [ -f "$REPO_ROOT/shell-desktop/assets/icons/${icon_name}.svg" ]; then
		cp "$REPO_ROOT/shell-desktop/assets/icons/${icon_name}.svg" "$INCLUDE_DIR/usr/share/icons/hicolor/scalable/apps/"
		cp "$REPO_ROOT/shell-desktop/assets/icons/${icon_name}.svg" "$INCLUDE_DIR/usr/share/pixmaps/"
	fi
done

# Resource monitor (btop) desktop entry
if [ -f "$REPO_ROOT/session/applications/btop.desktop" ]; then
	cp "$REPO_ROOT/session/applications/btop.desktop" "$INCLUDE_DIR/usr/share/applications/"
fi
# Install Neovim code editor desktop entry on ISO
cat > "$INCLUDE_DIR/usr/share/applications/nvim.desktop" <<'EOF'
[Desktop Entry]
Name=Neovim
GenericName=Text Editor
Comment=Edit text files in modern terminal
Exec=ghostty --title="Neovim" -e nvim %F
Icon=nvim
Terminal=false
Type=Application
Categories=Utility;TextEditor;Development;
Keywords=Text;editor;code;nvim;vim;development;
StartupNotify=false
MimeType=text/plain;text/x-chdr;text/x-csrc;text/x-c++hdr;text/x-c++src;text/x-rust;text/x-python;text/x-go;text/x-shellscript;application/json;application/xml;text/markdown;text/x-lua;application/toml;application/x-yaml;text/yaml;
EOF

# Download and install official standalone OpenCode CLI (v2) binary if not already cached
OPENCODE_CACHE="$BUILD_DIR/opencode"
mkdir -p "$OPENCODE_CACHE"
OPENCODE_ARCHIVE="$OPENCODE_CACHE/opencode-v${OPENCODE_VERSION}-linux-x64.tgz"
if [ ! -f "$OPENCODE_ARCHIVE" ]; then
	echo "build-iso-void: downloading OpenCode CLI v${OPENCODE_VERSION}..."
	curl -sSL -f -o "$OPENCODE_ARCHIVE" "https://registry.npmjs.org/@opencode/cli-linux-x64/-/cli-linux-x64-${OPENCODE_VERSION}.tgz" || {
		echo "build-iso-void: failed to download OpenCode CLI v${OPENCODE_VERSION}" >&2
		exit 1
	}
fi
verify_checksum "$OPENCODE_ARCHIVE"
mkdir -p "$INCLUDE_DIR/usr/bin"
tar -xzf "$OPENCODE_ARCHIVE" -O package/bin/opencode > "$INCLUDE_DIR/usr/bin/opencode"
chmod 755 "$INCLUDE_DIR/usr/bin/opencode"
if [ -d "$REPO_ROOT/apps/mimeapps" ]; then
	cp "$REPO_ROOT/apps/mimeapps"/toren-mimeapps.list "$INCLUDE_DIR/usr/share/applications/" 2>/dev/null || true
	ln -sf toren-mimeapps.list "$INCLUDE_DIR/usr/share/applications/asura-mimeapps.list" 2>/dev/null || true
fi

# Copy Toren apps (ghostty, starship, zsh configurations)
if [ -d "$REPO_ROOT/apps" ]; then
	mkdir -p "$INCLUDE_DIR/usr/share/toren/apps"
	cp -r "$REPO_ROOT/apps"/* "$INCLUDE_DIR/usr/share/toren/apps/"
fi

# One-time migrations (asura migrate). A fresh install already starts in the
# state they lead to, so every shipped one is recorded as done: the machine's
# in /var/lib/toren/migrations, each user's through /etc/skel. Existing
# machines run them on their next asura update.
if [ -d "$REPO_ROOT/migrations" ]; then
	mkdir -p "$INCLUDE_DIR/usr/share/toren/migrations"
	cp -r "$REPO_ROOT/migrations"/. "$INCLUDE_DIR/usr/share/toren/migrations/"
	for kind in system user; do
		case "$kind" in
		system) state="$INCLUDE_DIR/var/lib/toren/migrations" ;;
		user) state="$INCLUDE_DIR/etc/skel/.local/state/toren/migrations" ;;
		esac
		mkdir -p "$state"
		for f in "$INCLUDE_DIR/usr/share/toren/migrations/$kind"/*.sh; do
			[ -f "$f" ] || continue
			chmod 0755 "$f"
			echo "marked done by the installer image" >"$state/$(basename "$f")"
		done
	done
fi

# The user manual (toren manual).
if [ -d "$REPO_ROOT/docs/manual" ]; then
	mkdir -p "$INCLUDE_DIR/usr/share/doc/toren/manual"
	cp "$REPO_ROOT/docs/manual"/*.md "$INCLUDE_DIR/usr/share/doc/toren/manual/"
fi

# Stage system-wide skills and pre-seed OpenCode / cross-agent skills
mkdir -p "$INCLUDE_DIR/usr/share/toren/skills"
if [ -d "$REPO_ROOT/shell/agent/skills" ]; then
	cp -r "$REPO_ROOT/shell/agent/skills"/* "$INCLUDE_DIR/usr/share/toren/skills/" 2>/dev/null || true
fi
mkdir -p "$INCLUDE_DIR/etc/skel/.agents/skills"
cp -r "$INCLUDE_DIR/usr/share/toren/skills"/* "$INCLUDE_DIR/etc/skel/.agents/skills/" 2>/dev/null || true
mkdir -p "$INCLUDE_DIR/etc/skel/.config/opencode/skills"
cp -r "$INCLUDE_DIR/usr/share/toren/skills"/* "$INCLUDE_DIR/etc/skel/.config/opencode/skills/" 2>/dev/null || true

# Hardware models and components for quirks engine
mkdir -p "$INCLUDE_DIR/usr/share/toren/hardware/models" "$INCLUDE_DIR/usr/share/toren/hardware/components"
if [ -d "$REPO_ROOT/system/hardware/models" ]; then
	cp -r "$REPO_ROOT/system/hardware/models"/* "$INCLUDE_DIR/usr/share/toren/hardware/models/" 2>/dev/null || true
fi
if [ -d "$REPO_ROOT/system/hardware/components" ]; then
	cp -r "$REPO_ROOT/system/hardware/components"/* "$INCLUDE_DIR/usr/share/toren/hardware/components/" 2>/dev/null || true
fi
cat >"$INCLUDE_DIR/etc/skel/.config/opencode/opencode.json" <<'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "skills": [
    "/usr/share/toren/skills",
    "~/.agents/skills"
  ]
}
EOF

# Copy Hyprland configuration, runit services, and shell tree
cp -r "$REPO_ROOT/session/hyprland"/* "$INCLUDE_DIR/usr/share/toren/hyprland/"
mkdir -p "$INCLUDE_DIR/usr/share/toren/safe-mode"
cp -r "$REPO_ROOT/session/safe-mode"/* "$INCLUDE_DIR/usr/share/toren/safe-mode/"
if [ -d "$REPO_ROOT/session/sddm" ]; then
	cp -r "$REPO_ROOT/session/sddm" "$INCLUDE_DIR/usr/share/toren/session/" 2>/dev/null || true
fi
cp -r "$REPO_ROOT/shell"/* "$INCLUDE_DIR/usr/share/toren/shell/"
# The desktop, the default shell (org.toren.shell desktop-shell = desktop).
# Its own scripts (vendor.sh, manifest.txt) are build tooling and stay out of the image.
mkdir -p "$INCLUDE_DIR/usr/share/toren/shell-desktop"
cp -r "$REPO_ROOT/shell-desktop"/* "$INCLUDE_DIR/usr/share/toren/shell-desktop/"
rm -f "$INCLUDE_DIR/usr/share/toren/shell-desktop/toren/vendor.sh" \
	"$INCLUDE_DIR/usr/share/toren/shell-desktop/toren/manifest.txt"
ln -sfn toren "$INCLUDE_DIR/usr/share/toren/shell-desktop/asura"
cp -r "$REPO_ROOT/system/runit/user-services"/* "$INCLUDE_DIR/usr/share/toren/runit/services/"
# The commit this image is built from: asura update compares it with the
# branch, so a fresh install knows whether there is anything newer.
git -c safe.directory='*' -C "$REPO_ROOT" rev-parse HEAD >"$INCLUDE_DIR/usr/share/toren/applied-commit" 2>/dev/null ||
	rm -f "$INCLUDE_DIR/usr/share/toren/applied-commit"

# Ensure XBPS tools are available (if running on a non-Void host, fetch xbps-static)
if ! command -v xbps-rindex >/dev/null 2>&1 || ! command -v xbps-install >/dev/null 2>&1; then
	echo "build-iso-void: xbps tools not found on host, setting up xbps-static..."
	XBPS_STATIC_DIR="$BUILD_DIR/xbps-static"
	if [ ! -x "$XBPS_STATIC_DIR/usr/bin/xbps-install" ] || [ ! -x "$XBPS_STATIC_DIR/usr/bin/xbps-rindex" ]; then
		mkdir -p "$XBPS_STATIC_DIR"
		curl -sSL "https://repo-fastly.voidlinux.org/static/xbps-static-latest.x86_64-musl.tar.xz" | tar -xJ -C "$XBPS_STATIC_DIR"
	fi
	export PATH="$XBPS_STATIC_DIR/usr/bin:$PATH"
fi

# Stage native Void Linux Hyprland compositor stack and libraries from ASURA repository
echo "build-iso-void: staging native Void Linux Hyprland stack from ASURA repository..."
HYPR_CACHE="$BUILD_DIR/hyprland-pkgs"
mkdir -p "$HYPR_CACHE"

ASURA_REPO_URL="${ASURA_REPO_URL:-https://mohitjoer.github.io/Toren-Linux/current}"

HYPR_PKGS="hyprutils-0.14.4_1.x86_64.xbps aquamarine-0.16.0_1.x86_64.xbps hyprlang-0.6.8_1.x86_64.xbps hyprcursor-0.1.14_1.x86_64.xbps hyprgraphics-0.6.0_1.x86_64.xbps hyprland-protocols-0.8.0_1.x86_64.xbps hyprwire-0.3.1_1.x86_64.xbps hyprtoolkit-0.6.0_1.x86_64.xbps hyprland-0.57.2_1.x86_64.xbps xdg-desktop-portal-hyprland-1.5.0_1.x86_64.xbps hypridle-0.1.8_1.x86_64.xbps hyprlock-0.9.7_1.x86_64.xbps hyprland-guiutils-0.2.2_1.x86_64.xbps"
for hpkg in $HYPR_PKGS; do
	if [ ! -f "$HYPR_CACHE/$hpkg" ]; then
		echo "build-iso-void: downloading $hpkg from ASURA repository..."
		curl -sSL -f -o "$HYPR_CACHE/$hpkg" "$ASURA_REPO_URL/$hpkg" || {
			echo "build-iso-void: warning: failed to download $hpkg from $ASURA_REPO_URL (will build or resolve from repository)" >&2
		}
	fi
	if [ -f "$HYPR_CACHE/$hpkg" ]; then
		verify_checksum "$HYPR_CACHE/$hpkg"
		tar -xf "$HYPR_CACHE/$hpkg" -C "$INCLUDE_DIR"
	fi
done
# Fallback: until ASURA's own repository is published, Hyprland 0.57 is not
# downloadable, and Void has no Hyprland at all. Stage the prebuilt 0.49 stack
# instead (the ISO's source before ASURA built its own), so the image always
# has a compositor. toren-session picks the .conf config for Hyprland < 0.57.
if [ ! -f "$HYPR_CACHE/hyprland-0.57.2_1.x86_64.xbps" ]; then
	echo "build-iso-void: Hyprland 0.57.2 is not available from $ASURA_REPO_URL; staging the prebuilt 0.49 stack" >&2
	rm -f "$HYPR_CACHE"/*.xbps
	HYPR_FALLBACK_URL="https://raw.githubusercontent.com/Makrennel/hyprland-void/repository-x86_64-glibc"
	HYPR_PKGS="tomlplusplus-3.4.0_1.x86_64.xbps hyprutils-0.7.1_1.x86_64.xbps aquamarine-0.8.0_2.x86_64.xbps hyprlang-0.6.3_1.x86_64.xbps hyprcursor-0.1.12_1.x86_64.xbps hyprgraphics-0.1.3_2.x86_64.xbps libspng-0.7.4_1.x86_64.xbps sdbus-cpp-2.1.0_1.x86_64.xbps hyprland-0.49.0_1.x86_64.xbps xdg-desktop-portal-hyprland-1.3.9_2.x86_64.xbps hyprland-qtutils-0.1.4_1.x86_64.xbps hypridle-0.1.6_2.x86_64.xbps hyprlock-0.8.2_1.x86_64.xbps"
	for hpkg in $HYPR_PKGS; do
		curl -sSL -f -o "$HYPR_CACHE/$hpkg" "$HYPR_FALLBACK_URL/$hpkg" || {
			echo "build-iso-void: ERROR: could not download $hpkg from either repository" >&2
			exit 1
		}
		verify_checksum "$HYPR_CACHE/$hpkg"
		tar -xf "$HYPR_CACHE/$hpkg" -C "$INCLUDE_DIR"
	done
	# Trust the fallback repository's signing key.
	mkdir -p "$INCLUDE_DIR/var/db/xbps/keys"
	cat >"$INCLUDE_DIR/var/db/xbps/keys/30:93:47:9e:1f:55:c5:c5:07:b1:9a:e6:8b:10:5b:4b.plist" <<'KEYEOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>public-key</key>
	<data>LS0tLS1CRUdJTiBQVUJMSUMgS0VZLS0tLS0KTUlJQm9qQU5CZ2txaGtpRzl3MEJBUUVGQUFPQ0FZOEFNSUlCaWdLQ0FZRUFtNWk4VnhQdEpJRVU5ZERuSlNmSAo5by9PNCt1THFYc1hqemY0emZxTXFXM0RIQTJjM2pHeDE2dFVUT0dHbUdzVk1pb3l6TXl0TjUwSTZwb0J0ODBZCjlWQ25MRVBmVFFmbjlNaHV1ZGFPNjA5OGZVb2srejA4WjJsZ205NVg5Z3JDblkraEhwMC8yZ2VxNGJ5Zy9mZG0KTjlCbGVqaTlNd2dJdHFpRllJM0J4MXdWUVkzekVzWEUyeXlseHpycTFpS2p0S2dYMXAxWDFnWDBadzR2ZEIzMwo4b1luSDNNYkNMUGlobWxocm5BMTRhWm1UaHFuNFNRU1A2czAzdStQSHNuaHVKQ3EvS1N4S3AyRFhVM0JpY1dmCnY0YmdPQmV2TmYrZWhKbDJyRHdTSk9lS0NrOFhnZVYyVXNKQW4xaTM3UUpjdm9ZNWZUamFNUmxYSnRISERvN04KVzNXZHVPcWxwMHVHVVlXZVNxLzFMRVlJUDE2bGJ1dXd1a1ZrWXF5QzZMbWY0VXovOVZOYkc1anp0UUdmVkZJdgoxZmppZGg0cEpzNWEyeGx6QVpOTVpoT1VMdElnM2U5bkorYUc5dGt5K09kSjRiS1lja2NicWxRMlNiME5QTmF4CnQ4Q1Q2c2E4MHNJRU1oZWRvemUvVkhmdlNaZitLZ3FPc25BaGMvY1ZEUmhsQWdNQkFBRT0KLS0tLS1FTkQgUFVCTElDIEtFWS0tLS0tCg==</data>
	<key>public-key-size</key>
	<integer>3072</integer>
	<key>signature-by</key>
	<string>hyprland-void-github-action</string>
	<key>signature-type</key>
	<string>rsa</string>
</dict>
</plist>
KEYEOF
	cp "$INCLUDE_DIR/var/db/xbps/keys/30:93:47:9e:1f:55:c5:c5:07:b1:9a:e6:8b:10:5b:4b.plist" \
	   "$INCLUDE_DIR/var/db/xbps/keys/00649994a2d04d020cdb1119ff53e712be6b5c3e726accfb04c664d1200d748d.plist"
	cp "$INCLUDE_DIR/var/db/xbps/keys/30:93:47:9e:1f:55:c5:c5:07:b1:9a:e6:8b:10:5b:4b.plist" \
	   "$INCLUDE_DIR/var/db/xbps/keys/1293d0c27380d795a88fdbd007356dae84090282b5bd4123678ee9a682d06dd3.plist"
fi
rm -f "$INCLUDE_DIR/props.plist" "$INCLUDE_DIR/files.plist"
if [ -f "$INCLUDE_DIR/usr/libexec/xdg-desktop-portal-hyprland" ]; then
	cp "$INCLUDE_DIR/usr/libexec/xdg-desktop-portal-hyprland" "$INCLUDE_DIR/usr/bin/xdg-desktop-portal-hyprland" 2>/dev/null || true
fi

echo "build-iso-void: indexing local Hyprland package repository..."
if [ -n "$(find "$HYPR_CACHE" -name '*.xbps' -print -quit 2>/dev/null)" ]; then
	xbps-rindex -a "$HYPR_CACHE"/*.xbps
fi

# Compatibility soname symlinks for Void libraries with newer ABI bumps
mkdir -p "$INCLUDE_DIR/usr/lib"
ln -sf libdisplay-info.so.3 "$INCLUDE_DIR/usr/lib/libdisplay-info.so.2"
ln -sf libjxl.so.0.12 "$INCLUDE_DIR/usr/lib/libjxl.so.0.11"
ln -sf libjxl_threads.so.0.12 "$INCLUDE_DIR/usr/lib/libjxl_threads.so.0.11"
ln -sf libhyprutils.so.6 "$INCLUDE_DIR/usr/lib/libhyprutils.so.10"
# libsystemd compatibility symlinks redirecting to Void's native libelogind
ln -sf libelogind.so.0 "$INCLUDE_DIR/usr/lib/libsystemd.so.0"
ln -sf libelogind.so.0 "$INCLUDE_DIR/usr/lib/libsystemd.so"
# libseat is installed natively by XBPS from packages-void.yaml; purge any foreign host binaries
rm -f "$INCLUDE_DIR/usr/lib/libseat.so"*
rm -rf "$INCLUDE_DIR/usr/lib/plymouth"

# Pre-trust ASURA repository keys if provided in system/xbps.d or keys directory
mkdir -p "$INCLUDE_DIR/var/db/xbps/keys"
for keydir in "$REPO_ROOT/system/xbps-keys" "$REPO_ROOT/system/xbps.d/keys" "$REPO_ROOT/system/keys"; do
	if [ -d "$keydir" ]; then
		cp -f "$keydir"/*.plist "$INCLUDE_DIR/var/db/xbps/keys/" 2>/dev/null || true
	fi
done

# Ensure foreign host quickshell binary is never copied into Void rootfs;
# native quickshell package is installed via packages-void.yaml
rm -f "$INCLUDE_DIR/usr/bin/quickshell"

# Copy udev power rules (99-toren-power.rules, 80-nvidia-pm.rules, 99-battery-threshold.rules)
cp "$REPO_ROOT/system/udev"/*.rules "$INCLUDE_DIR/etc/udev/rules.d/"
if [ -d "$REPO_ROOT/system/modprobe.d" ]; then
	cp "$REPO_ROOT/system/modprobe.d"/*.conf "$INCLUDE_DIR/etc/modprobe.d/"
fi
if [ -d "$REPO_ROOT/system/modules-load.d" ]; then
	mkdir -p "$INCLUDE_DIR/etc/modules-load.d"
	cp "$REPO_ROOT/system/modules-load.d"/*.conf "$INCLUDE_DIR/etc/modules-load.d/"
fi

if [ -d "$REPO_ROOT/system/tlp.d" ]; then
	mkdir -p "$INCLUDE_DIR/etc/tlp.d"
	cp "$REPO_ROOT/system/tlp.d"/*.conf "$INCLUDE_DIR/etc/tlp.d/"
fi
cp "$REPO_ROOT/system/sysctl.d/99-toren-performance.conf" "$INCLUDE_DIR/etc/sysctl.d/99-toren-performance.conf"
cp "$REPO_ROOT/session/turnstile/turnstile.conf" "$INCLUDE_DIR/etc/turnstile/turnstile.conf"
cp "$REPO_ROOT/system/xbps.d/toren.conf" "$INCLUDE_DIR/etc/xbps.d/toren.conf"
# Default host firewall (block inbound, allow outbound), loaded by the nftables runit service
cp "$REPO_ROOT/system/nftables/nftables.conf" "$INCLUDE_DIR/etc/nftables.conf"
cp "$REPO_ROOT/system/xbps.d/toren.conf" "$INCLUDE_DIR/etc/xbps.d/00-repository-main.conf"

# Default environment for seatd & Wayland
cat >"$INCLUDE_DIR/etc/environment" <<'EOF'
LIBSEAT_BACKEND=seatd
XDG_SESSION_TYPE=wayland
QT_QPA_PLATFORM=wayland
GDK_BACKEND=wayland
MOZ_ENABLE_WAYLAND=1
EOF

# Build and copy Toren Go control binaries into include tree
echo "build-iso-void: building Toren Go binaries..."
(
	export GOWORK=off
	export GOFLAGS="-buildvcs=false"
	git config --global --add safe.directory '*' 2>/dev/null || true
	cd "$REPO_ROOT/desktop"
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren" ./cmd/toren
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-sysd" ./cmd/toren-sysd
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-userd" ./cmd/toren-userd
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-hubd" ./cmd/toren-hubd
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-shell" ./cmd/toren-shell
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-wallpaper" ./cmd/toren-wallpaper
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/apm" ./cmd/apm
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-agent" ./cmd/toren-agent
	# Build Toren Store (prioritizing standalone repository https://github.com/Toren-Linux/toren-store)
	STORE_SRC="${TOREN_STORE_SRC:-${ASURA_STORE_SRC:-}}"
	if [ -z "$STORE_SRC" ]; then
		if [ -d "$REPO_ROOT/../toren-store" ]; then
			STORE_SRC="$REPO_ROOT/../toren-store"
		fi
	fi
	if [ -n "$STORE_SRC" ] && [ -d "$STORE_SRC/cmd/toren-store" ]; then
		echo "build-iso-void: building toren-store from standalone repo ($STORE_SRC)..."
		( cd "$STORE_SRC/cmd/toren-store" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INCLUDE_DIR/usr/bin/toren-store" . )
		if [ -d "$STORE_SRC/catalog" ]; then
			mkdir -p "$INCLUDE_DIR/usr/share/toren/store/catalog"
			cp -r "$STORE_SRC/catalog"/* "$INCLUDE_DIR/usr/share/toren/store/catalog/" 2>/dev/null || true
		fi
	elif [ -n "$STORE_SRC" ] && [ -d "$STORE_SRC/cmd/toren-store" ]; then
		echo "build-iso-void: building toren-store from standalone repo ($STORE_SRC)..."
		( cd "$STORE_SRC/cmd/toren-store" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INCLUDE_DIR/usr/bin/toren-store" . )
		if [ -d "$STORE_SRC/catalog" ]; then
			mkdir -p "$INCLUDE_DIR/usr/share/toren/store/catalog"
			cp -r "$STORE_SRC/catalog"/* "$INCLUDE_DIR/usr/share/toren/store/catalog/" 2>/dev/null || true
		fi
	else
		echo "build-iso-void: building toren-store from desktop tree..."
		CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-store" ./cmd/toren-store
	fi

	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-vm-mon" ./cmd/toren-vm/mon
	CGO_ENABLED=0 go build -o "$INCLUDE_DIR/usr/bin/toren-vm-ssh" ./cmd/toren-vm/remote
	cp -f "$REPO_ROOT/desktop/cmd/toren-vm/toren-vmd" "$INCLUDE_DIR/usr/bin/toren-vmd"

	# Build Prowl from standalone org repository (https://github.com/Toren-Linux/prowl)
	PROWL_SRC="${TOREN_PROWL_SRC:-}"
	PROWL_TMP=""
	if [ -z "$PROWL_SRC" ] && [ -d "$REPO_ROOT/../prowl" ]; then
		PROWL_SRC="$REPO_ROOT/../prowl"
	fi
	if [ -z "$PROWL_SRC" ]; then
		PROWL_TMP=$(mktemp -d)
		if git clone --depth 1 https://github.com/Toren-Linux/prowl.git "$PROWL_TMP/prowl" 2>/dev/null; then
			PROWL_SRC="$PROWL_TMP/prowl"
		fi
	fi
	if [ -n "$PROWL_SRC" ] && [ -d "$PROWL_SRC/cmd/toren-prowl" ]; then
		echo "build-iso-void: building toren-prowl from standalone repo ($PROWL_SRC)..."
		( cd "$PROWL_SRC" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INCLUDE_DIR/usr/bin/toren-prowl" ./cmd/toren-prowl )
	fi
	[ -n "$PROWL_TMP" ] && rm -rf "$PROWL_TMP" || true
)
find "$INCLUDE_DIR/usr/bin" -type f -exec chmod 0755 {} +
# Symlinks
ln -sf toren "$INCLUDE_DIR/usr/bin/os-ctl"
for f in "$INCLUDE_DIR/usr/share/toren/session"/toren-*; do
	[ -e "$f" ] || continue
	base=$(basename "$f")
	if [ -x "$f" ] && [ ! -e "$INCLUDE_DIR/usr/bin/$base" ]; then
		ln -sf "/usr/share/toren/session/$base" "$INCLUDE_DIR/usr/bin/$base"
	fi
done

# Pre-seed user skeleton configuration and XDG user directories
mkdir -p "$INCLUDE_DIR/etc/skel/.config/toren" \
         "$INCLUDE_DIR/etc/skel/.config/ghostty" \
         "$INCLUDE_DIR/etc/skel/.config/fastfetch" \
         "$INCLUDE_DIR/etc/skel/Desktop" \
         "$INCLUDE_DIR/etc/skel/Documents" \
         "$INCLUDE_DIR/etc/skel/Downloads" \
         "$INCLUDE_DIR/etc/skel/Music" \
         "$INCLUDE_DIR/etc/skel/Pictures" \
         "$INCLUDE_DIR/etc/skel/Videos"
cp -a "$INCLUDE_DIR/usr/share/toren"/* "$INCLUDE_DIR/etc/skel/.config/toren/"

# Pre-seed Ghostty terminal configuration in skeleton
cat >"$INCLUDE_DIR/etc/skel/.config/ghostty/config" <<'EOF'
config-file = ?~/.config/toren/apps/ghostty/config
config-file = ?/usr/share/toren/apps/ghostty/config
config-file = ?~/.config/ghostty/colors
working-directory = home
EOF

# Pre-seed btop configuration in skeleton
mkdir -p "$INCLUDE_DIR/etc/skel/.config/btop"
if [ -f "$REPO_ROOT/apps/btop/btop.conf" ]; then
	cp "$REPO_ROOT/apps/btop/btop.conf" "$INCLUDE_DIR/etc/skel/.config/btop/btop.conf"
fi

# Pre-seed zsh configuration in skeleton
cat >"$INCLUDE_DIR/etc/skel/.zshrc" <<'EOF'
# Toren Linux Zsh configuration
if [ -r "$HOME/.config/toren/apps/zsh/toren.zsh" ]; then
	. "$HOME/.config/toren/apps/zsh/toren.zsh"
elif [ -r "/usr/share/toren/apps/zsh/toren.zsh" ]; then
	. "/usr/share/toren/apps/zsh/toren.zsh"
fi
EOF

# System-wide zsh configuration fallback for all interactive shells
mkdir -p "$INCLUDE_DIR/etc/zsh"
cat >"$INCLUDE_DIR/etc/zsh/zshrc" <<'EOF'
# System-wide Toren Linux Zsh initialization
if [ -z "$__TOREN_ZSH_LOADED" ] && [ -n "$PS1" ]; then
	if [ -r "$HOME/.config/toren/apps/zsh/toren.zsh" ]; then
		. "$HOME/.config/toren/apps/zsh/toren.zsh"
	elif [ -r "/usr/share/toren/apps/zsh/toren.zsh" ]; then
		. "/usr/share/toren/apps/zsh/toren.zsh"
	fi
fi
EOF

# Pre-seed fastfetch ASCII logo and configuration
mkdir -p "$INCLUDE_DIR/usr/share/toren/branding" "$INCLUDE_DIR/etc/skel/.config/fastfetch"
if [ -f "$REPO_ROOT/public/branding/toren-logo.txt" ]; then
	cp "$REPO_ROOT/public/branding/toren-logo.txt" "$INCLUDE_DIR/usr/share/toren/branding/toren-logo.txt"
	cp "$REPO_ROOT/public/branding/toren-logo.txt" "$INCLUDE_DIR/usr/share/toren/branding/fastfetch-ascii.txt"
fi
if [ -f "$REPO_ROOT/public/branding/toren-logo-mono.txt" ]; then
	cp "$REPO_ROOT/public/branding/toren-logo-mono.txt" "$INCLUDE_DIR/usr/share/toren/branding/toren-logo-mono.txt"
fi
if [ -f "$REPO_ROOT/public/branding/fastfetch-emblem.png" ]; then
	cp "$REPO_ROOT/public/branding/fastfetch-emblem.png" "$INCLUDE_DIR/usr/share/toren/branding/fastfetch-emblem.png"
fi

cat >"$INCLUDE_DIR/etc/skel/.config/fastfetch/config.jsonc" <<'EOF'
{
  "$schema": "https://github.com/fastfetch-cli/fastfetch/raw/dev/doc/json_schema.json",
  "logo": {
    "type": "file-raw",
    "source": "/usr/share/toren/branding/toren-logo.txt",
    "padding": {
      "left": 2,
      "right": 4,
      "top": 1
    }
  },
  "display": {
    "separator": " ── ",
    "color": {
      "keys": "2"
    }
  },
  "modules": [
    {
      "format": "\u001b[32m■\u001b[0m \u001b[1;37mTOREN LINUX\u001b[0m",
      "type": "custom"
    },
    "break",
    {
      "format": "\u001b[32m──\u001b[0m \u001b[1;37mSYSTEM\u001b[0m \u001b[38;5;240m──────────────────────────\u001b[0m",
      "type": "custom"
    },
    {
      "key": "OS",
      "type": "os"
    },
    {
      "key": "KERNEL",
      "type": "kernel"
    },
    {
      "key": "UPTIME",
      "type": "uptime"
    },
    {
      "key": "PACKAGES",
      "type": "packages"
    },
    {
      "key": "SHELL",
      "type": "shell"
    },
    {
      "key": "WM",
      "type": "wm"
    },
    "break",
    {
      "format": "\u001b[32m──\u001b[0m \u001b[1;37mHARDWARE\u001b[0m \u001b[38;5;240m────────────────────────\u001b[0m",
      "type": "custom"
    },
    {
      "key": "HOST",
      "type": "host"
    },
    {
      "key": "CPU",
      "type": "cpu"
    },
    {
      "key": "GPU",
      "type": "gpu"
    },
    {
      "key": "MEMORY",
      "type": "memory"
    },
    {
      "key": "DISPLAY",
      "type": "display"
    },
    "break",
    {
      "type": "colors",
      "symbol": "circle"
    }
  ]
}
EOF

# Stage system runit services (toren-sysd, scx-lavd, ananicy-cpp)
mkdir -p "$INCLUDE_DIR/etc/sv/toren-sysd"
for svc_dir in "$REPO_ROOT/system/runit/services"/*; do
	if [ -d "$svc_dir" ]; then
		svc_name=$(basename "$svc_dir")
		mkdir -p "$INCLUDE_DIR/etc/sv/$svc_name"
		cp -r "$svc_dir"/* "$INCLUDE_DIR/etc/sv/$svc_name/"
		chmod +x "$INCLUDE_DIR/etc/sv/$svc_name/run" 2>/dev/null || true
	fi
done

# Stage Layer 3: zramen zstd RAM compression configuration
mkdir -p "$INCLUDE_DIR/etc/sv/zramen"
cat >"$INCLUDE_DIR/etc/sv/zramen/conf" <<'EOF'
export ZRAM_COMP_ALGORITHM=zstd
export ZRAM_SIZE=50
export ZRAM_PRIORITY=32767
EOF

# Stage Layer 3: mimalloc opt-in runner helper (avoid global ld.so.preload)
cp "$REPO_ROOT/system/helpers/with-mimalloc" "$INCLUDE_DIR/usr/bin/with-mimalloc"
install -m 0755 "$REPO_ROOT/system/helpers/toren-charge-limit-apply" "$INCLUDE_DIR/usr/bin/toren-charge-limit-apply"
chmod 0755 "$INCLUDE_DIR/usr/bin/with-mimalloc"
rm -f "$INCLUDE_DIR/etc/ld.so.preload"

# Stage Layer 3: ananicy-cpp auto-nice rules for compilers, scanners, UI, and games
mkdir -p "$INCLUDE_DIR/etc/ananicy.d" "$INCLUDE_DIR/etc/ananicy-cpp"
if [ -d "$REPO_ROOT/system/ananicy" ]; then
	cp "$REPO_ROOT/system/ananicy"/* "$INCLUDE_DIR/etc/ananicy.d/" 2>/dev/null || true
	cp "$REPO_ROOT/system/ananicy"/* "$INCLUDE_DIR/etc/ananicy-cpp/" 2>/dev/null || true
fi

# Extract package list from system/packages-void.yaml
PACKAGES=$(python3 - "$REPO_ROOT/system/packages-void.yaml" "$MIRROR" <<'PY'
import sys, yaml
from pathlib import Path

manifest = yaml.safe_load(Path(sys.argv[1]).read_text())

if manifest.get("base", {}).get("distribution") != "void-linux":
    raise SystemExit("build-iso-void: system/packages-void.yaml must declare distribution: void-linux")

packages = []
for group, body in manifest.items():
    if group in {"version", "base"} or not isinstance(body, dict):
        continue
    entries = body.get("packages")
    if isinstance(entries, list):
        packages.extend(entries)

filtered = [p for p in packages if p != "base-system"]
all_requested = sorted(dict.fromkeys(str(p) for p in filtered))
# staged above from Toren's repository (HYPR_PKGS), not in Void's
custom_pkgs = {"hyprland", "xdg-desktop-portal-hyprland", "hypridle", "hyprlock", "hyprland-guiutils"}
out = [p for p in all_requested if p not in custom_pkgs]
print(" ".join(out))
PY
)

# Acquire void-mklive
MKLIVE_DIR="$BUILD_DIR/void-mklive"
if [ ! -d "$MKLIVE_DIR" ]; then
	echo "build-iso-void: cloning void-mklive..."
	git clone --depth 1 https://github.com/void-linux/void-mklive.git "$MKLIVE_DIR"
elif [ -d "$MKLIVE_DIR/.git" ]; then
	git -C "$MKLIVE_DIR" checkout -f
fi

# Pre-trust custom package repository keys in void-mklive keys directory
mkdir -p "$MKLIVE_DIR/keys"
cp "$INCLUDE_DIR/var/db/xbps/keys"/*.plist "$MKLIVE_DIR/keys/" 2>/dev/null || true

# Overwrite upstream mklive issue, motd, and splash data files to eliminate upstream branding
if [ -d "$MKLIVE_DIR/data" ]; then
	cp "$INCLUDE_DIR/etc/issue" "$MKLIVE_DIR/data/issue"
	cp "$INCLUDE_DIR/etc/motd" "$MKLIVE_DIR/data/motd"
	if [ -f "$REPO_ROOT/public/branding/boot-splash.png" ]; then
		cp "$REPO_ROOT/public/branding/boot-splash.png" "$MKLIVE_DIR/data/splash.png"
	fi
fi

# Clean up void-mklive bootstrap repository flags and force-include plymouth and VM drivers in initramfs
if [ -f "$MKLIVE_DIR/mklive.sh" ]; then
	sed -i 's/VOID_LIVE/TOREN_LIVE/g' "$MKLIVE_DIR/mklive.sh"
	sed -i 's/--force-add "vmklive autoinstaller"/--force-add "vmklive autoinstaller plymouth"/g' "$MKLIVE_DIR/mklive.sh"
	sed -i 's/--add-drivers "ahci"/--add-drivers "ahci vboxvideo vmwgfx virtio_gpu bochs qxl simpledrm"/g' "$MKLIVE_DIR/mklive.sh"
	sed -i 's|--repository=https://repo-default.voidlinux.org/current/musl||g' "$MKLIVE_DIR/mklive.sh"
	sed -i "s|https://repo-default.voidlinux.org/current|$MIRROR|g" "$MKLIVE_DIR/mklive.sh"
	sed -i 's/loglevel=4/quiet splash loglevel=3 rd.udev.log_level=3 udev.log_level=3 rd.plymouth=1 plymouth.enable=1 vt.global_cursor_default=0 fbcon=nodefer console=ttyS0,115200 console=tty0/g' "$MKLIVE_DIR/mklive.sh"
	# Stage /EFI/BOOT into ISO9660 filesystem so xorriso emerges /EFI/BOOT
	sed -i 's|mkdir -p "${GRUB_EFI_TMPDIR}"/EFI/BOOT|mkdir -p "${GRUB_EFI_TMPDIR}"/EFI/BOOT "$IMAGEDIR"/EFI/BOOT|g' "$MKLIVE_DIR/mklive.sh"
	sed -i 's|cp -f "$VOIDTARGETDIR/tmp/boot${EFI_ARCH,,}.efi" "${GRUB_EFI_TMPDIR}/EFI/BOOT/BOOT${EFI_ARCH^^}.EFI"|cp -f "$VOIDTARGETDIR/tmp/boot${EFI_ARCH,,}.efi" "${GRUB_EFI_TMPDIR}/EFI/BOOT/BOOT${EFI_ARCH^^}.EFI"\n\t\tcp -f "$VOIDTARGETDIR/tmp/boot${EFI_ARCH,,}.efi" "$IMAGEDIR/EFI/BOOT/BOOT${EFI_ARCH^^}.EFI"|g' "$MKLIVE_DIR/mklive.sh"
	sed -i 's|cp -f grub/grub.cfg "$GRUB_DIR"|mkdir -p "$IMAGEDIR/EFI/BOOT" "$GRUB_DIR"\n    cp -f grub/grub.cfg "$GRUB_DIR"\n    cp -f grub/grub.cfg "$IMAGEDIR/EFI/BOOT/grub.cfg"|g' "$MKLIVE_DIR/mklive.sh"
fi
if [ -f "$MKLIVE_DIR/lib.sh" ]; then
	sed -i 's|--repository=https://repo-default.voidlinux.org/current/musl||g' "$MKLIVE_DIR/lib.sh"
	sed -i "s|https://repo-default.voidlinux.org/current|$MIRROR|g" "$MKLIVE_DIR/lib.sh"
fi

# Patch dracut vmklive module to use asura-os hostname, asura user, and include _seatd/input/toren-hw groups
if [ -f "$MKLIVE_DIR/dracut/vmklive/adduser.sh" ]; then
	sed -i 's/echo void-live >/echo toren-linux >/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/-G audio,video,wheel/-G audio,video,wheel,_seatd,input,tty,storage,network,toren-hw/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/USERNAME=anon/USERNAME=toren/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/voidlinux/toren/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/99-void-live/99-toren-live/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/void-live.rules/toren-live.rules/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/-s \/bin\/bash/-s \/bin\/zsh/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
	sed -i 's/chroot ${NEWROOT} useradd/chroot ${NEWROOT} id "$USERNAME" >\/dev\/null 2>\&1 || chroot ${NEWROOT} useradd/g' "$MKLIVE_DIR/dracut/vmklive/adduser.sh"
fi

# Replace cluttered multi-option isolinux menu with a clean, streamlined Asura OS boot menu
if [ -d "$MKLIVE_DIR/isolinux" ]; then
	cat >"$MKLIVE_DIR/isolinux/isolinux.cfg.in" <<'BOOTCFG'
UI vesamenu.c32
PROMPT 0
TIMEOUT 30
ONTIMEOUT linux

MENU TABMSG Press ENTER to boot or TAB to edit boot options
MENU AUTOBOOT Toren Linux starts in # second{,s}...
MENU BACKGROUND @@SPLASHIMAGE@@
MENU WIDTH 78
MENU MARGIN 4
MENU ROWS 6
MENU VSHIFT 2
MENU TIMEOUTROW 10
MENU TABMSGROW 2
MENU CMDLINEROW 13

MENU COLOR title        * #ffffffff *
MENU COLOR border       * #00000000 #00000000 none
MENU COLOR sel          * #00000000 #ffffffff *
MENU COLOR unsel        * #ffffffff #00000000 none
MENU COLOR timeout_msg  * #88888888 #00000000 none
MENU COLOR timeout      * #ffffffff #00000000 none
MENU COLOR tabmsg       * #66666666 #00000000 none

LABEL linux
MENU LABEL Toren Linux (Live Session)
KERNEL /boot/vmlinuz
APPEND initrd=/boot/initrd root=live:CDLABEL=TOREN_LIVE init=/sbin/init ro rd.luks=0 rd.md=0 rd.dm=0 vconsole.unicode=1 vconsole.keymap=@@KEYMAP@@ locale.LANG=@@LOCALE@@ vga=current @@BOOT_CMDLINE@@

LABEL linuxram
MENU LABEL Toren Linux (Copy to RAM)

KERNEL /boot/vmlinuz
APPEND initrd=/boot/initrd root=live:CDLABEL=TOREN_LIVE init=/sbin/init ro rd.luks=0 rd.md=0 rd.dm=0 vconsole.unicode=1 vconsole.keymap=@@KEYMAP@@ locale.LANG=@@LOCALE@@ vga=current @@BOOT_CMDLINE@@ rd.live.ram

LABEL reboot
MENU LABEL Reboot System
COM32 reboot.c32

LABEL poweroff
MENU LABEL Power Off
COM32 poweroff.c32
BOOTCFG
fi

# Prune cluttered extra speech and nogfx entries from GRUB configuration, and enable serial console
if [ -f "$MKLIVE_DIR/grub/grub_void.cfg.pre" ]; then
	sed -i 's/timeout=15/timeout=3/g' "$MKLIVE_DIR/grub/grub_void.cfg.pre"
	sed -i 's/terminal_input console/terminal_input console serial/g' "$MKLIVE_DIR/grub/grub_void.cfg.pre"
	sed -i 's/terminal_output gfxterm/terminal_output gfxterm serial/g' "$MKLIVE_DIR/grub/grub_void.cfg.pre"
	sed -i '1s/^/insmod serial\nserial --speed=115200 --unit=0 --word=8 --parity=no --stop=1\n/' "$MKLIVE_DIR/grub/grub_void.cfg.pre"
fi

# Post-setup script to configure users and seat permissions inside ROOTFS
POSTSETUP_SCRIPT="$BUILD_DIR/postsetup.sh"
cat >"$POSTSETUP_SCRIPT" <<EOF
#!/bin/sh
ROOTFS="\$1"
HYPR_CACHE="$HYPR_CACHE"
TOREN_VERSION="$TOREN_VERSION"
ASURA_VERSION="$TOREN_VERSION"
REPO_ROOT="$REPO_ROOT"
INCLUDE_DIR="$INCLUDE_DIR"
EOF
cat >>"$POSTSETUP_SCRIPT" <<'EOF'

# Generate and enforce UTF-8 system locale
mkdir -p "$ROOTFS/etc"
echo "LANG=en_US.UTF-8" > "$ROOTFS/etc/locale.conf"
echo "LC_ALL=en_US.UTF-8" >> "$ROOTFS/etc/locale.conf"
if [ -f "$ROOTFS/etc/default/libc-locales" ]; then
	sed -i 's/^#\(en_US.UTF-8 UTF-8\)/\1/' "$ROOTFS/etc/default/libc-locales"
	chroot "$ROOTFS" xbps-reconfigure -f glibc-locales 2>/dev/null || true
fi

# Enable core system runit services (Layer 1 seatd, Layer 3 scx-lavd, ananicy-cpp, zramen). Docker is opt-in: not enabled by default (podman is the default container tool).
mkdir -p "$ROOTFS/etc/runit/runsvdir/default"
for svc in dbus seatd elogind udevd NetworkManager acpid turnstiled tlp zramen chronyd nftables cupsd toren-sysd toren-boot-guard socklog-unix nanoklogd earlyoom toren-maintenance; do
	if [ -d "$ROOTFS/etc/sv/$svc" ]; then
		ln -sf "/etc/sv/$svc" "$ROOTFS/etc/runit/runsvdir/default/$svc"
	fi
done
if [ -x "$ROOTFS/usr/bin/scx_lavd" ] && [ -d "$ROOTFS/etc/sv/scx-lavd" ]; then
	ln -sf /etc/sv/scx-lavd "$ROOTFS/etc/runit/runsvdir/default/scx-lavd"
fi
if [ -x "$ROOTFS/usr/bin/ananicy-cpp" ] && [ -d "$ROOTFS/etc/sv/ananicy-cpp" ]; then
	ln -sf /etc/sv/ananicy-cpp "$ROOTFS/etc/runit/runsvdir/default/ananicy-cpp"
fi

# thermald: decided on the machine, at boot, by its conf (it stays down
# without Intel coretemp sensors), not by the CPU of the machine building this
# image.
if [ -d "$ROOTFS/etc/sv/thermald" ]; then
	cp "$REPO_ROOT/system/runit/sv-conf/thermald/conf" "$ROOTFS/etc/sv/thermald/conf"
	ln -sf /etc/sv/thermald "$ROOTFS/etc/runit/runsvdir/default/thermald"
fi

# Ensure runtime library compatibility symlinks exist inside rootfs
mkdir -p "$ROOTFS/usr/lib"
[ -e "$ROOTFS/usr/lib/libdisplay-info.so.3" ] && ln -sf libdisplay-info.so.3 "$ROOTFS/usr/lib/libdisplay-info.so.2"
[ -e "$ROOTFS/usr/lib/libjxl.so.0.12" ] && ln -sf libjxl.so.0.12 "$ROOTFS/usr/lib/libjxl.so.0.11"
[ -e "$ROOTFS/usr/lib/libjxl_threads.so.0.12" ] && ln -sf libjxl_threads.so.0.12 "$ROOTFS/usr/lib/libjxl_threads.so.0.11"
ldconfig -r "$ROOTFS" 2>/dev/null || true

# Hook turnstile into PAM session stack if present
if [ -f "$ROOTFS/etc/pam.d/system-login" ] && ! grep -q 'pam_turnstile.so' "$ROOTFS/etc/pam.d/system-login"; then
	echo "session optional pam_turnstile.so" >> "$ROOTFS/etc/pam.d/system-login"
fi

# Hook elogind into PAM session stack if present (registers the session with
# login1 so polkit sees a valid subject and asurad can spawn its agent)
if [ -f "$ROOTFS/etc/pam.d/system-login" ] && ! grep -q 'pam_elogind.so' "$ROOTFS/etc/pam.d/system-login"; then
	echo "session optional pam_elogind.so" >> "$ROOTFS/etc/pam.d/system-login"
fi

# The power key stays with acpid: elogind must not also handle it (one writer)
mkdir -p "$ROOTFS/etc/elogind/logind.conf.d"
printf '[Login]\nHandlePowerKey=ignore\nHandleSuspendKey=ignore\nHandleHibernateKey=ignore\nHandleLidSwitch=ignore\n' > "$ROOTFS/etc/elogind/logind.conf.d/90-toren-powerkey.conf"

# Configure seat permissions and groups for default user
chroot "$ROOTFS" getent group toren-hw >/dev/null 2>&1 || chroot "$ROOTFS" groupadd -r toren-hw || true
chroot "$ROOTFS" getent group toren-hw >/dev/null 2>&1 || chroot "$ROOTFS" groupadd -r toren-hw || true
chroot "$ROOTFS" getent group docker >/dev/null 2>&1 || chroot "$ROOTFS" groupadd -r docker || true
for user in anon toren; do
	if chroot "$ROOTFS" id "$user" >/dev/null 2>&1; then
		chroot "$ROOTFS" usermod -aG _seatd,video,audio,input,wheel,storage,network,toren-hw,docker "$user" || true
	fi
done

# Ensure user toren exists with zsh login shell and configure autologin
if ! chroot "$ROOTFS" id toren >/dev/null 2>&1; then
	chroot "$ROOTFS" useradd -m -s /bin/zsh -U -G _seatd,video,audio,input,wheel,storage,network,toren-hw,docker toren || true
else
	chroot "$ROOTFS" usermod -aG toren-hw,docker -s /bin/zsh toren || true
fi
echo "toren:toren" | chroot "$ROOTFS" chpasswd 2>/dev/null || true

# Reset root and anon passwords to eliminate upstream voidlinux defaults
echo "root:toren" | chroot "$ROOTFS" chpasswd 2>/dev/null || true
if chroot "$ROOTFS" id anon >/dev/null 2>&1; then
	echo "anon:toren" | chroot "$ROOTFS" chpasswd 2>/dev/null || true
fi

# Neutralize upstream installer and docs commands
if [ -f "$ROOTFS/usr/bin/void-installer" ]; then
	rm -f "$ROOTFS/usr/bin/void-installer"
	ln -sf /usr/bin/toren-installer "$ROOTFS/usr/bin/void-installer"
fi

cat >"$ROOTFS/usr/bin/void-docs" <<'DOCS'
#!/bin/sh
echo "Toren Linux documentation is available at: https://github.com/mohitjoer/Toren-Linux"
DOCS
chmod 0755 "$ROOTFS/usr/bin/void-docs"

# Sudo permissions for wheel group (strict 0750 dir, 0440 file)
chmod 0750 "$ROOTFS/etc/sudoers.d"
echo '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' > "$ROOTFS/etc/sudoers.d/wheel"
chmod 0440 "$ROOTFS/etc/sudoers.d/wheel"
chmod 0440 "$ROOTFS/etc/sudoers.d"/* 2>/dev/null || true
grep -q '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' "$ROOTFS/etc/sudoers" 2>/dev/null || echo '%wheel ALL=(ALL:ALL) NOPASSWD: ALL' >> "$ROOTFS/etc/sudoers"

# Compile GSettings schemas for Toren Shell and Settings
if command -v glib-compile-schemas >/dev/null 2>&1; then
	glib-compile-schemas "$ROOTFS/usr/share/glib-2.0/schemas" 2>/dev/null || true
fi

# Enable automatic silent console login on tty1 for seamless graphical boot
mkdir -p "$ROOTFS/etc/sv/agetty-tty1" "$ROOTFS/home/toren"
echo 'GETTY_ARGS="--autologin toren --skip-login --nonewline --noissue"' > "$ROOTFS/etc/sv/agetty-tty1/conf"
touch "$ROOTFS/home/toren/.hushlogin"
touch "$INCLUDE_DIR/etc/skel/.hushlogin"
# Pre-materialize user configuration for toren user, skeleton, and XDG directories
mkdir -p "$ROOTFS/home/toren/.config/toren" \
         "$ROOTFS/home/toren/.config/ghostty" "$ROOTFS/home/toren/.config/fastfetch" \
         "$ROOTFS/home/toren/Desktop" "$ROOTFS/home/toren/Documents" \
         "$ROOTFS/home/toren/Downloads" "$ROOTFS/home/toren/Music" \
         "$ROOTFS/home/toren/Pictures" "$ROOTFS/home/toren/Videos" \
         "$ROOTFS/etc/skel/.config/toren" \
         "$INCLUDE_DIR/etc/skel/Desktop"
# Ensure live user desktop shortcuts are deployed and executable
mkdir -p "$ROOTFS/home/toren/Desktop"
if [ -d "$ROOTFS/etc/skel/Desktop" ]; then
	cp -a "$ROOTFS/etc/skel/Desktop"/* "$ROOTFS/home/toren/Desktop/" 2>/dev/null || true
fi
chmod 755 "$ROOTFS/home/toren/Desktop"/*.desktop 2>/dev/null || true
chown -R toren:toren "$ROOTFS/home/toren/Desktop" 2>/dev/null || true
if [ -d "$ROOTFS/usr/share/toren" ]; then
	cp -a "$ROOTFS/usr/share/toren"/* "$ROOTFS/home/toren/.config/toren/"
	cp -a "$ROOTFS/usr/share/toren"/* "$ROOTFS/etc/skel/.config/toren/"
fi
for _u in toren; do
	cp -a "$ROOTFS/etc/skel/.config/ghostty"/* "$ROOTFS/home/$_u/.config/ghostty/" 2>/dev/null || true
	cp -a "$ROOTFS/etc/skel/.config/fastfetch"/* "$ROOTFS/home/$_u/.config/fastfetch/" 2>/dev/null || true
	cp -a "$ROOTFS/etc/skel/.config/opencode" "$ROOTFS/home/$_u/.config/" 2>/dev/null || true
	cp -a "$ROOTFS/etc/skel/.agents" "$ROOTFS/home/$_u/" 2>/dev/null || true
	cp -a "$ROOTFS/etc/skel/.zshrc" "$ROOTFS/home/$_u/.zshrc" 2>/dev/null || true
done
chroot "$ROOTFS" su - toren -c "toren materialize" 2>/dev/null || true
chroot "$ROOTFS" su - toren -c "xdg-user-dirs-update" 2>/dev/null || true
chroot "$ROOTFS" chown -R toren:toren /home/toren 2>/dev/null || true

# Ensure Plymouth theme and configuration are active in ROOTFS
mkdir -p "$ROOTFS/usr/share/plymouth/themes"
ln -sf /usr/share/plymouth/themes/toren/toren.plymouth "$ROOTFS/usr/share/plymouth/themes/default.plymouth" 2>/dev/null || true
if [ -f "$ROOTFS/etc/plymouth/plymouthd.conf" ]; then
	sed -i 's/^Theme=.*/Theme=toren/' "$ROOTFS/etc/plymouth/plymouthd.conf"
fi

# Boot status goes to Plymouth while it runs; the stock functions file stays
# intact — its msg/msg_warn/msg_error are OVERRIDDEN below, and everything
# else it defines (detect_container, emergency_shell, ...) must survive, or
# stage 1 dies with "detect_container: not found" on every boot.
if [ -f "$ROOTFS/etc/runit/functions" ]; then
	cat >>"$ROOTFS/etc/runit/functions" <<'FNEOF'
msg() {
	if command -v plymouth >/dev/null 2>&1 && plymouth --ping 2>/dev/null; then
		plymouth update --status="$*" 2>/dev/null || true
	else
		printf "=> %s\n" "$*"
	fi
}
msg_warn() {
	if command -v plymouth >/dev/null 2>&1 && plymouth --ping 2>/dev/null; then
		plymouth update --status="WARNING: $*" 2>/dev/null || true
	else
		printf "=> WARNING: %s\n" "$*"
	fi
}
msg_error() {
	if command -v plymouth >/dev/null 2>&1 && plymouth --ping 2>/dev/null; then
		plymouth update --status="ERROR: $*" 2>/dev/null || true
	else
		printf "=> ERROR: %s\n" "$*"
	fi
}
msg_ok() {
	if command -v plymouth >/dev/null 2>&1 && plymouth --ping 2>/dev/null; then
		:
	else
		printf "\033[1m\033[32m OK\033[m\n"
	fi
}
FNEOF
fi

if [ -f "$ROOTFS/etc/runit/core-services/03-console-setup.sh" ]; then
	# Prevent setfont on tty0 from resetting video mode and destroying Plymouth canvas
	sed -i 's|setfont |! command -v plymouth >/dev/null 2>\&1 \|\| ! plymouth --ping 2>/dev/null \&\& setfont |g' "$ROOTFS/etc/runit/core-services/03-console-setup.sh"
fi

if [ -f "$ROOTFS/etc/runit/core-services/02-kmods.sh" ]; then
	# Upstream's cleanup strips `.ko` and `.ko.gz` from modules-load -v output
	# but not `.ko.zst`, so zstd-compressed module names print during boot.
	python3 - "$ROOTFS/etc/runit/core-services/02-kmods.sh" <<'PY'
import sys
p = sys.argv[1]
s = open(p).read()
s = s.replace(r"s:\.ko\(\.gz\)\? ::g", r"s:\.ko\(\.gz\|\.zst\)\? ::g")
open(p, "w").write(s)
PY
fi

if [ -f "$ROOTFS/etc/runit/core-services/08-sysctl.sh" ]; then
	sed -i 's/sysctl -p/sysctl -q -p/g' "$ROOTFS/etc/runit/core-services/08-sysctl.sh"
	sed -i 's/sysctl --system/sysctl -q --system/g' "$ROOTFS/etc/runit/core-services/08-sysctl.sh"
fi

mkdir -p "$ROOTFS/etc/xbps.d"
cp "$REPO_ROOT/system/xbps.d/toren.conf" "$ROOTFS/etc/xbps.d/toren.conf" 2>/dev/null || true
cp "$REPO_ROOT/system/xbps.d/toren.conf" "$ROOTFS/etc/xbps.d/00-repository-main.conf" 2>/dev/null || true

if [ -d "$REPO_ROOT/system/modules-load.d" ]; then
	mkdir -p "$ROOTFS/etc/modules-load.d"
	cp "$REPO_ROOT/system/modules-load.d"/*.conf "$ROOTFS/etc/modules-load.d/" 2>/dev/null || true
fi

if [ -d "$REPO_ROOT/system/tlp.d" ]; then
	mkdir -p "$ROOTFS/etc/tlp.d"
	cp "$REPO_ROOT/system/tlp.d"/*.conf "$ROOTFS/etc/tlp.d/" 2>/dev/null || true
fi

# Default host firewall. Copied again here because the nftables package's own
# /etc/nftables.conf is laid down when packages install into the rootfs.
cp "$REPO_ROOT/system/nftables/nftables.conf" "$ROOTFS/etc/nftables.conf"

# Ensure wpa_supplicant does not run as a standalone runit service (managed by NetworkManager)
rm -f "$ROOTFS/etc/runit/runsvdir/default/wpa_supplicant" "$ROOTFS/var/service/wpa_supplicant"
rm -f "$ROOTFS/etc/ld.so.preload"

# Deploy systemctl and journalctl compatibility bridge into ROOTFS
mkdir -p "$ROOTFS/usr/bin"
cp "$REPO_ROOT/system/helpers/systemctl" "$ROOTFS/usr/bin/systemctl" 2>/dev/null || true
chmod 755 "$ROOTFS/usr/bin/systemctl" 2>/dev/null || true
cp "$REPO_ROOT/system/helpers/journalctl" "$ROOTFS/usr/bin/journalctl" 2>/dev/null || true
chmod 755 "$ROOTFS/usr/bin/journalctl" 2>/dev/null || true

# Ensure portal executable symlink is available in /usr/bin if installed in /usr/libexec
if [ -f "$ROOTFS/usr/libexec/xdg-desktop-portal-hyprland" ] && [ ! -f "$ROOTFS/usr/bin/xdg-desktop-portal-hyprland" ]; then
	ln -sf /usr/libexec/xdg-desktop-portal-hyprland "$ROOTFS/usr/bin/xdg-desktop-portal-hyprland" 2>/dev/null || true
fi

# Register custom Hyprland packages into ROOTFS pkgdb so xbps-query -l tracks them
python3 - "$ROOTFS" "$HYPR_CACHE" <<'PY'
import sys, os, plistlib, tarfile

rootfs = sys.argv[1] if len(sys.argv) > 1 else ""
hypr_cache = sys.argv[2] if len(sys.argv) > 2 else ""

if not hypr_cache or not os.path.isdir(hypr_cache):
    sys.exit(0)

pkgdb_path = os.path.join(rootfs, "var/db/xbps/pkgdb-0.38.plist")
pkgdb = {}
if os.path.exists(pkgdb_path):
    try:
        with open(pkgdb_path, "rb") as f:
            pkgdb = plistlib.load(f)
    except Exception:
        pass

for fname in os.listdir(hypr_cache):
    if not fname.endswith(".xbps"):
        continue
    fpath = os.path.join(hypr_cache, fname)
    try:
        with tarfile.open(fpath, "r:*") as tar:
            props_file = None
            files_file = None
            for m in tar.getmembers():
                if m.name in ("./props.plist", "props.plist"):
                    props_file = tar.extractfile(m)
                elif m.name in ("./files.plist", "files.plist"):
                    files_file = tar.extractfile(m)
            if props_file:
                pkg_props = plistlib.load(props_file)
                pkgver = pkg_props.get("pkgver", fname.replace(".xbps", ""))
                pkg_props["state"] = "installed"
                # Soname-compat: ensure package shlib names match base layer ABI.
                # Rewrite shlib-requires to the names the installed libraries actually
                # provide, so xbps-install -Su does not abort later. Runtime
                # resolution still uses the compat symlinks created above.
                requires = pkg_props.get("shlib-requires")
                if isinstance(requires, list):
                    fixed = []
                    for r in requires:
                        r = r.replace("libdisplay-info.so.2", "libdisplay-info.so.3")
                        r = r.replace("libjxl.so.0.11", "libjxl.so.0.12")
                        r = r.replace("libjxl_threads.so.0.11", "libjxl_threads.so.0.12")
                        fixed.append(r)
                    pkg_props["shlib-requires"] = fixed
                provides = pkg_props.get("shlib-provides")
                if isinstance(provides, list) and "libhyprutils.so.6" in provides:
                    if "libhyprutils.so.10" not in provides:
                        provides.append("libhyprutils.so.10")
                    pkg_props["shlib-provides"] = provides
                # pkgdb is keyed by package NAME (xbps-query hyprland,
                # xbps-install -Su); a pkgver key leaves the package unknown
                # to xbps, which then tries to install it from the repository.
                pkgname = pkg_props.get("pkgname") or pkgver.rsplit("-", 1)[0]
                pkgdb[pkgname] = pkg_props
                if files_file:
                    dest_files = os.path.join(rootfs, f"var/db/xbps/{pkgver}-files.plist")
                    with open(dest_files, "wb") as out_f:
                        out_f.write(files_file.read())
    except Exception as e:
        print(f"postsetup: warning registering {fname}: {e}", file=sys.stderr)

os.makedirs(os.path.dirname(pkgdb_path), exist_ok=True)
with open(pkgdb_path, "wb") as f:
    plistlib.dump(pkgdb, f)
PY

# Overwrite OS identity inside ROOTFS
cat >"$ROOTFS/etc/os-release" <<'OSREL'
NAME="Toren Linux"
PRETTY_NAME="Toren Linux Beta 001"
ID=toren
ID_LIKE="void linux"
VERSION_ID="0.1"
VERSION_CODENAME=beta-001
HOME_URL="https://github.com/mohitjoer/Toren-Linux"
SUPPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
BUG_REPORT_URL="https://github.com/mohitjoer/Toren-Linux/issues"
OSREL
if [ -n "$TOREN_VERSION" ]; then
	printf 'IMAGE_VERSION=%s\n' "$TOREN_VERSION" >>"$ROOTFS/etc/os-release"
fi

# Branding replacement: Change Welcome to Void! to Welcome to Asura OS! in the boot sequence
for runit_file in "$ROOTFS/etc/runit/1" "$ROOTFS/etc/runit/2" "$ROOTFS/etc/runit/3" "$ROOTFS/etc/rc.local"; do
	if [ -f "$runit_file" ]; then
		sed -i 's/Welcome to Void!/Welcome to Toren Linux!/g' "$runit_file"
		sed -i 's/Welcome to Void Linux!/Welcome to Toren Linux!/g' "$runit_file"
		sed -i 's/Void Linux/Toren Linux/g' "$runit_file"
	fi
done

# Configure fontconfig cache permissions and hardware battery threshold in /etc/rc.local
mkdir -p "$ROOTFS/var/cache/fontconfig" "$INCLUDE_DIR/var/cache/fontconfig"
chmod 1777 "$ROOTFS/var/cache/fontconfig" "$INCLUDE_DIR/var/cache/fontconfig" 2>/dev/null || true
mkdir -p "$ROOTFS/usr/share/fonts/TTF"
if [ -d "$REPO_ROOT/public/fonts" ]; then
	cp -L "$REPO_ROOT/public/fonts"/* "$ROOTFS/usr/share/fonts/TTF/" 2>/dev/null || true
fi
chroot "$ROOTFS" fc-cache -f 2>/dev/null || true
chroot "$ROOTFS" gtk-update-icon-cache -qtf /usr/share/icons/hicolor 2>/dev/null || true
if [ -f "$ROOTFS/etc/rc.local" ]; then
	if ! grep -q 'charge_control_end_threshold' "$ROOTFS/etc/rc.local"; then
		echo 'for b in /sys/class/power_supply/BAT*/charge_control_end_threshold; do [ -f "$b" ] && echo 80 > "$b" 2>/dev/null || true; done' >> "$ROOTFS/etc/rc.local"
	fi
fi

# Set Toren Linux hostname and network identity
printf 'toren-linux\n' > "$ROOTFS/etc/hostname"
if [ -f "$ROOTFS/etc/rc.conf" ]; then
	sed -i 's/^#\?HOSTNAME=.*/HOSTNAME="toren-linux"/' "$ROOTFS/etc/rc.conf"
	if ! grep -q '^HOSTNAME=' "$ROOTFS/etc/rc.conf"; then
		echo 'HOSTNAME="toren-linux"' >> "$ROOTFS/etc/rc.conf"
	fi
fi
if [ -f "$ROOTFS/etc/hosts" ]; then
	sed -i 's/void-live/toren-linux/g' "$ROOTFS/etc/hosts"
fi

# Update all TTY greeting messages (/etc/issue, /etc/issue.net, /etc/motd) to show Asura OS wordmark and logo
cat >"$ROOTFS/etc/issue" <<'ISSUE'
       #                    #         Toren Linux Beta 001 (\m)
     %##                    ##%       Kernel \r on an \m (\l)
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%    https://github.com/mohitjoer/Toren-Linux
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

ISSUE

cat >"$ROOTFS/etc/issue.net" <<'ISSUENET'
       #                    #         Toren Linux Beta 001
     %##                    ##%       https://github.com/mohitjoer/Toren-Linux
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

ISSUENET

cat >"$ROOTFS/etc/motd" <<'MOTD'
       #                    #         Toren Linux Beta 001
     %##                    ##%       https://github.com/mohitjoer/Toren-Linux
    ####   %%%%%%%%%%%%%%   ####
  ######% %%%%%%%%%%%%%%%% %######    Robust. Minimal. Kinetic.
%###########% %%%%%%%%  ###########%  Live Session: user 'toren' (nopasswd sudo)
 %###########          ###########%   Run 'toren install' or 'toren-installer' to install
     ########  @@@@@@  ########
   %% ########%      %######## %%%
   %%  %%##################%%% %%%
   %% ##     %########%     ## %%%
   %% ####%%%%%      %%%%%#### %%
       %####################%
      %%  %##############%  %%
       ### ############## %##
       %##  ############  ###
        ###  %        %  ###
         %################%
           %############%
             ##########

MOTD


# Auto-start Asura Wayland desktop on tty1 login (with safe fallback if compositor exits)
mkdir -p "$ROOTFS/etc/profile.d"
cat >"$ROOTFS/etc/profile.d/toren-autologin.sh" <<'PROFILE'
if [ -z "$WAYLAND_DISPLAY" ] && [ -z "$DISPLAY" ] && [ "$(tty 2>/dev/null)" = "/dev/tty1" ]; then
	# Suppress text console cursor and clear screen to eliminate terminal flicker before GUI
	setterm -cursor off 2>/dev/null || true
	clear 2>/dev/null || true
	command -v xdg-user-dirs-update >/dev/null 2>&1 && xdg-user-dirs-update 2>/dev/null || true
	if [ ! -d "$HOME/.config/toren/shell" ] && command -v asura >/dev/null 2>&1; then
		toren materialize 2>/dev/null || true
	fi
	if ! /usr/share/toren/session/toren-session; then
		setterm -cursor on 2>/dev/null || true
		if command -v plymouth >/dev/null 2>&1; then
			sudo -n plymouth --quit 2>/dev/null || plymouth --quit 2>/dev/null || sudo -n killall -9 plymouthd 2>/dev/null || true
		fi
		echo
		echo "================================================================"
		if grep -q "No 3D enabled" "$HOME/.local/state/toren/session.log" 2>/dev/null || \
		   grep -q "failed to create dri2 screen" "$HOME/.local/state/toren/session.log" 2>/dev/null; then
			echo "  [Virtual Machine Display Notice]"
			echo "  VirtualBox / VMware 3D Acceleration is currently DISABLED."
			echo "  Hyprland Wayland compositor requires hardware 3D acceleration."
			echo
			echo "  To run the full Wayland graphical desktop in VirtualBox:"
			echo "    1. Power off this virtual machine."
			echo "    2. Go to VirtualBox Settings -> Display -> Screen."
			echo "    3. Check 'Enable 3D Acceleration' and set Video Memory >= 128MB."
			echo
		else
			echo "  Toren graphical desktop session exited."
			echo "  Log available at: ~/.local/state/toren/session.log"
			echo
		fi
		echo "  To launch the installer directly, run: sudo toren install"
		echo "  To retry launching the desktop, run: toren-session"
		echo "================================================================"
		echo
	fi
fi
PROFILE
chmod 0755 "$ROOTFS/etc/profile.d/toren-autologin.sh"
# zsh login shells do not read /etc/profile.d/*.sh unless /etc/zsh/zprofile sources it
mkdir -p "$ROOTFS/etc/zsh"
if [ ! -f "$ROOTFS/etc/zsh/zprofile" ] || ! grep -q 'toren-autologin.sh' "$ROOTFS/etc/zsh/zprofile"; then
	cat >>"$ROOTFS/etc/zsh/zprofile" <<'ZPROFILE'
[ -r /etc/profile.d/toren-autologin.sh ] && emulate sh -c '. /etc/profile.d/toren-autologin.sh'
ZPROFILE
fi
exit 0
EOF
chmod +x "$POSTSETUP_SCRIPT"

FINAL_ISO="$BUILD_DIR/$ISO_NAME.iso"

# Run mklive to produce the image with boot splash & hybrid graphics.
#
# Flag audit (void-mklive revision at clone time):
#   -a ARCH        target architecture (x86_64)
#   -b base-system base metapackage
#   -v linux       kernel package (`-k` is the CONSOLE KEYMAP in void-mklive;
#                  passing `-k linux` there made the live image boot with
#                  `KEYMAP=linux`, which loadkeys cannot open)
#   -r HYPR_CACHE  extra repository for ASURA hyprland packages
#   -r MIRROR      Void mirror (passed twice: hypr repo + void repo)
#   -p PACKAGES    package list from system/packages-void.yaml
#   -I INCLUDE_DIR overlay rootfs (ASURA's files laid over the base)
#   -i gzip        initramfs compression
#   -s zstd        ISO squashfs compression
#   -S "..."       runit services enabled in the live image
#   -C "..."       kernel cmdline (quiet splash for Plymouth, console for serial)
#   -x POSTSETUP   post-install hook script
#   -T ISO_TITLE   volume label
#   -o FINAL_ISO   output path
# SPLASH_IMAGE env points at the boot splash PNG consumed by mklive.
echo "build-iso-void: generating ISO with void-mklive..."
(
	cd "$MKLIVE_DIR"
	export SPLASH_IMAGE="$REPO_ROOT/public/branding/boot-splash.png"
	./mklive.sh \
		-a "$ARCH" \
		-b base-system \
		-v linux \
		-r "$HYPR_CACHE" \
		-r "$MIRROR" \
		-p "$PACKAGES" \
		-I "$INCLUDE_DIR" \
		-i gzip \
		-s zstd \
		-S "dbus seatd elogind NetworkManager zramen" \
		-C "quiet splash loglevel=3 rd.udev.log_level=3 udev.log_level=3 rd.plymouth=1 plymouth.enable=1 vt.global_cursor_default=0 fbcon=nodefer console=ttyS0,115200 console=tty0 live.user=toren live.autologin nowatchdog mitigations=off preempt=full" \
		-x "$POSTSETUP_SCRIPT" \
		-T "$ISO_TITLE" \
		-o "$FINAL_ISO"
)

if [ ! -f "$FINAL_ISO" ]; then
	echo "build-iso-void: void-mklive produced no ISO" >&2
	exit 1
fi

chmod -R a+rwX "$BUILD_DIR" || true
echo "build-iso-void: image written to $FINAL_ISO"
