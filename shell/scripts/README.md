# shell/scripts

The helper scripts the shell and the compositor config call by bare name. The shell
runs under `systemd --user`, whose PATH has no `~/.local/bin`, so each one is installed
into `/usr/bin` (`release/build.d/W13.build`). Every script detects its tools: a missing
one is a notification (these run with no terminal to print to) plus a stderr line naming
the apt package, and exit 127. The toren-hubd, toren-shell and Hyprland sides are
reached directly (`hyprctl`, `toren-shell`), not through a wrapper.

| script                                               | tools it needs                                                             | Ubuntu package                                                        |
| ---------------------------------------------------- | -------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `toren-app`                                          | `jq`, `ghostty` / `nautilus` (any terminal and file manager), `gtk-launch` | `jq`, `libgtk-3-bin`                                                  |
| `toren-summon`                                       | `hyprctl`, `jq`                                                            | `jq`                                                                  |
| `toren-sysinfo`, `toren-profile-stats`               | `lspci`, `hyprctl`, `jq`, `dpkg-query`                                     | `pciutils`, `jq`                                                      |
| `toren-cmd-caffeine`                                 | `systemd-inhibit`, `systemd-run`                                           | base                                                                  |
| `toren-cmd-game-mode`                                | `hyprctl`, `powerprofilesctl`; optional `pkexec` helpers                   | `power-profiles-daemon`                                               |
| `toren-cmd-calc`                                     | `qalc` (python3 fallback)                                                  | `qalc`                                                                |
| `toren-cmd-color-picker`                             | `hyprpicker`, `wl-copy`                                                    | `hyprpicker`, `wl-clipboard`                                          |
| `toren-cmd-ocr`                                      | `grim`, `slurp`, `tesseract`, `wl-copy`                                    | `grim`, `slurp`, `tesseract-ocr`, `tesseract-ocr-eng`, `wl-clipboard` |
| `toren-cmd-qr-scan`                                  | `grim`, `slurp`, `zbarimg`, `wl-copy`, `xdg-open`                          | `zbar-tools`, `xdg-utils`                                             |
| `toren-cmd-google-lens`                              | `grim`, `slurp`, `curl`, `jq`, `xdg-open`                                  | `curl`, `jq`                                                          |
| `toren-cmd-songrec`                                  | `songrec`, `wl-copy`                                                       | not in the archive: `ppa:marin-m/songrec`                             |
| `toren-cmd-folders`                                  | Papirus icons, `gsettings`                                                 | `papirus-icon-theme`                                                  |
| `toren-cmd-radio`                                    | `mpv`, `jq`; `yt-dlp` for the YouTube stations                             | `mpv`, `mpv-mpris`, `yt-dlp`                                          |
| `toren-cmd-screenrecord`                             | `wf-recorder`, `slurp`, `pactl`; `gpu-screen-recorder` if installed        | `wf-recorder`, `slurp`                                                |
| `toren-cmd-studiorecord`, `toren-cmd-edit-recording` | the recorder above, `toren-motion`                                         | not in the archive                                                    |
| `toren-cmd-recording-saved`                          | `notify-send`, `ffprobe`                                                   | `libnotify-bin`, `ffmpeg`                                             |
| `toren-cmd-discord-compress`                         | `ffmpeg`, `ffprobe`                                                        | `ffmpeg`                                                              |
| `stash-install.sh`                                   | `pkexec` + `apt-get` for a `.deb`; `flatpak`, `bsdtar` optional            | `pkexec`, `flatpak`, `libarchive-tools`                               |
| `stash-download.sh`, `stash-cobalt.sh`               | `yt-dlp`, `curl`, `jq`, `ffmpeg`; `docker` for the cobalt engine           | `yt-dlp`, `docker.io`                                                 |
| `stash-compress.sh`                                  | `ffmpeg`, `magick`                                                         | `ffmpeg`, `imagemagick`                                               |
| `localsend.sh`                                       | `python3`, `curl`, `openssl`, `ip`, `file`                                 | base                                                                  |
| `cliphist-thumbs.sh`                                 | `cliphist`, `magick`                                                       | `cliphist`, `imagemagick`                                             |

`../bin/{claude,codex,opencode}-usage` are run by `toren-ai-usage.service` (timer in
`session/units/`), not by name. Checks live in `system/helpers/tools/tests/`.
