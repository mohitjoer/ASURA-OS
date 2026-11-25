# NOTICE

The desktop shell in `shell/` is a port of the upstream project (https://github.com/upstream/shell),
which is GPL-3.0. Toren is GPL-3.0 as well, so the licence is compatible, and this
notice is the honest ledge for where the code came from.

- Source: upstream, under GPL-3.0.
- What came across: the whole quickshell shell tree (the bar's frame, the dock,
  the launcher, the desktop, the wallpaper engine, the services, the ui kit).
- What did not come across: upstream name, its redistributable installers, its
  usage dashboards, and its welcome/pin/screenshot client configs.
- Branding in the running desktop says "Toren". The library names inside the
  tree are the ones the imports resolve through.
- First-party audio and image assets were not copied. The unlicensed third-party
  display fonts in `shell/assets/fonts/` have been replaced with open-licensed equivalents
  under the SIL Open Font License (Space Grotesk and JetBrains Mono Nerd Font).

## The second port: upstream 0.75.5-beta.20

The paragraphs above describe the first port, the shell tree. A second port brought
across the processes, windows and configuration around that tree. Where the two
differ, this section is current: the welcome tour, the pin, the reload cover, the
screenshot tool (without its beautify editor) and the AI-usage collectors did come
across, so the "did not come across" line above no longer holds for them.

- Taken from: upstream commit `27b4e706b3eba9196734fa362392cdcf3f551ef3`, `VERSION`
  0.75.5-beta.20, codename "Onogoro". Upstream is GPL-3.0, and so is everything below
  that came from it, except the portions named under "Third-party code".
- How: `tools/port.py SRC DST` copies text files from a upstream checkout, applies the
  rename rules (upstream names and QML module names to Toren's), never overwrites a
  file, skips binary assets, leaves any line that carries a copyright, SPDX or credit
  marker untouched, and prints every mention of the old name it left. Go imports and
  everything Arch-specific were then fixed by hand. A re-port is that one command
  against a newer checkout, followed by the same hand repair.
- Attribution: files keep the copyright, SPDX and "ported from" lines they came
  with or were given. Where a port differs from upstream on purpose, the file says so,
  and `desktop/cmd/toren-shell/CONTRACT.md` and `desktop/cmd/toren-hubd/CONTRACT.md`
  list every difference in the wire and the verbs.

### What came across

| In this tree                                                                                                                                                                                                                | From upstream                                                                                                             | Notes                                                                                                                                          |
| --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `shell/shell.qml`, `Services/`, `Ui/`, `components/`, `utils/`, `Barkit/`, `FrameBars/`, `modules/` (bar, capture, clipboard, confirm, desktop, dock, launcher, notifications, osd, overview, stage, visualizer, wallpaper) | `upstream/shell/quickshell/shell/`, `upstream/ui/`, `upstream/shell/framebars/`                                           | The first port, adapted since                                                                                                                  |
| `shell/Blobs-src/`                                                                                                                                                                                                          | `upstream/shell/plugin/`                                                                                                  | The frame-blob renderer: C++ sources and shaders                                                                                               |
| `shell/PluginKit/`, `shell/plugins/`, `shell/scripts/toren-plugins-place`                                                                                                                                                   | `upstream/shell/quickshell/plugins/`                                                                                      | The plugin kit, runtime and placement script                                                                                                   |
| `shell/hub.qml`, `shell/hub/`                                                                                                                                                                                               | `upstream/hub/quickshell/`                                                                                                | Toren Settings. `Toren.qml`, `TorenPage.qml`, `PageView.qml`, `Editor.qml`, `GVariant.qml` and `Controls/` are Toren's own                     |
| `shell/shot.qml`, `shell/shot/`                                                                                                                                                                                             | `upstream/shell/quickshell/toren-shot/`                                                                                   | The screenshot tool, without the beautify editor                                                                                               |
| `shell/pin.qml`, `shell/welcome.qml`, `shell/reload-cover.qml` and the directories of the same names                                                                                                                        | `upstream/shell/quickshell/toren-pin/`, `welcome/`, `reload-cover/`                                                       | Pinned shots, the first-run tour, the restart curtain                                                                                          |
| `shell/store.qml`, `shell/store/`                                                                                                                                                                                           | `upstream/apps/toren-store/quickshell/`                                                                                   | The store window                                                                                                                               |
| `shell/vm.qml`, `shell/vm/`                                                                                                                                                                                                 | `upstream/apps/toren-vm/quickshell/`                                                                                      | The VM manager window                                                                                                                          |
| `shell/wallpaper-ui.qml`, `shell/wallpaper-ui/`                                                                                                                                                                             | `upstream/shell/toren-wallpaper/wall-ui/`                                                                                 | The wallpaper picker (skwd-wall's, below)                                                                                                      |
| `shell/agent/skills/`                                                                                                                                                                                                       | `upstream/rashin/skills/upstream/`                                                                                        | Skills shipped to the agent's engine                                                                                                           |
| `shell/matugen/`                                                                                                                                                                                                            | `upstream/shell/matugen/`                                                                                                 | Templates only; those for apps Toren does not ship are gone                                                                                    |
| `shell/scripts/`, `shell/bin/`                                                                                                                                                                                              | `upstream/shell/scripts/`, `upstream/shell/bin/`                                                                          | Helper scripts and the three usage collectors                                                                                                  |
| `shell/i18n/data/`, `shell/i18n/overrides/`                                                                                                                                                                                 | `upstream/i18n/catalog/`, `catalog/overrides/`, `langs.json`                                                              | upstream machine translations (34 languages besides English), rebranded. Its extractor, translator and runtimes did not come across            |
| `desktop/cmd/toren-shell/`, `desktop/internal/wm/`                                                                                                                                                                          | `upstream/shell/ipc/`, `upstream/wm/`                                                                                     | The shell daemon and its CLI (`toren-shell`), and the compositor seam, Hyprland only                                                           |
| `desktop/cmd/toren-hubd/`                                                                                                                                                                                                   | `upstream/hub/backend/`, and the Hyprland provider and `settings.lua` generator of `upstream/wm/`                         | The Hub's data plane (`toren-hubd`); compositor settings are written as hyprlang                                                               |
| `desktop/cmd/toren-wallpaper/`                                                                                                                                                                                              | `upstream/shell/toren-wallpaper/daemon/`                                                                                  | The wallpaper daemon (`toren-wallpaper`)                                                                                                       |
| `desktop/cmd/toren-agent/`                                                                                                                                                                                                  | `upstream/rashin/backend/`                                                                                                | Toren Agent (Rashin). Its resident engine is not part of it; the engine is the user's own (`docs/agent.md`)                                    |
| `desktop/cmd/toren-store/`                                                                                                                                                                                                  | `upstream/apps/toren-store/backend/`                                                                                      | apt and Void's xbps; the local `catalog/` instead of the registry. The 1.0.7 sync below adds the revision, `check` and the window-manager gate |
| `desktop/cmd/toren-vm/`                                                                                                                                                                                                     | `upstream/apps/toren-vm/` (`bin/toren-vm`, `mon/`, `remote/`)                                                             | `toren-vmd`, `toren-vm-mon`, `toren-vm-ssh`                                                                                                    |
| `desktop/cmd/toren/plugin*.go`, `plugin_template/`, `wm.go`                                                                                                                                                                 | `upstream/cli/`                                                                                                           | `toren plugin` and `toren wm`. `icons.go`, `status_live.go` and `reconcile.go` follow upstream lookup order, field names and doctor pattern    |
| `desktop/internal/theme/scheme.go`                                                                                                                                                                                          | `upstream/shell/ipc/matugen.go`                                                                                           | `shellRoleKeys` and `baseSixteen`. The generator itself is Toren's: it replaces the matugen binary and is not matugen's arithmetic             |
| `desktop/internal/backlight/` tests                                                                                                                                                                                         | `system/hardware/display/toren-hw-backlight`                                                                              | The panel-selection cases                                                                                                                      |
| `session/hyprland/*.conf`                                                                                                                                                                                                   | `upstream/hyprland/` (`hyprland.lua`, `modules/`, `keyboard.lua`, `monitors.lua`, `shaders/`)                             | Translated from Lua to hyprlang for Hyprland 0.53; the shaders are copied                                                                      |
| `session/scripts/toren-workspace`, `toren-monitor`                                                                                                                                                                          | `upstream/hyprland/scripts/toren-workspace`, `toren-monitor`                                                              |                                                                                                                                                |
| `session/units/torend.service`, `toren-wallpaper.service`, `toren-idle.service`, `toren-agent.service`, `toren-ai-usage.*`                                                                                                  | `upstream/shell/systemd/user/`, `upstream/rashin/systemd/`                                                                | From `toren-shell`, `toren-wallpaper`, `toren-idle`, `toren-agent` and `toren-ai-usage`                                                        |
| `apps/`                                                                                                                                                                                                                     | `upstream/apps/` (`ghostty`, `zsh`, `starship`, `wireplumber`, `mimeapps.list`)                                           | Adapted to ghostty, zsh and starship as the defaults                                                                                           |
| `system/helpers/`                                                                                                                                                                                                           | `system/hardware/audio/`, `system/hardware/display/`, `system/extras/toren-pkg-add`                                       | `toren-volume`, `toren-restart-audio`, `toren-hw-backlight`, and `toren-pkg-add`/`-remove` rewritten for Void and polkit.                      |
| `system/helpers/tools/tests/`, `shell/reload-cover/reload_cover_test.sh`, `session/scripts/monitor_test.sh`                                                                                                                 | `tests/` (`calc-eval.sh`, `radio-engine.sh`, `game-mode.sh`, `stash-install.sh`, `reload-cover.sh`, `monitor-persist.sh`) | Hermetic script tests, rewritten where they assumed Arch. The other tests in those directories are Toren's own                                 |
| `docs/plugins.md`                                                                                                                                                                                                           | `docs/plugins.md`                                                                                                         | The plugin author's guide                                                                                                                      |

The vendored `shell/modules/common/`, `shell/modules/ii/` and `shell/services/` and the
shapes kit are end-4's, not upstream. They are recorded in `shell/vendor/manifest.yaml`,
`shell/vendor/notices.md` and `LICENSING.md`.

### What did not come across

- upstream name, brand art, decors and wallpapers (`upstream/assets/`), and its Hub and
  dashboard fonts (`upstream/hub/quickshell/fonts/`, `upstream/rashin/backend/web/fonts/`).
- upstream installer (`installation/` and `shell-installer/` in its repository)
  and the Arch, AUR and pacman paths.
- niri: the provider, `upstream/niri/`, and the Lua configuration itself, which was
  translated rather than copied.
- The Hyprland C++ plugins (the key-sounds plugin among them), the lock skins
  (`upstream/lockscreen/`, qylock), the browser extension and the palette bridge.
- `system/hardware/` tuning: the udev rules, NetworkManager, Bluetooth, GPU, power and
  input drivers and rules. This is deliberate. The hardware profile system owns
  machine-specific tuning, and the base-layer boundary forbids those paths.
- Features with no backend in the Ubuntu archive, or that Toren decided against:
  wallpaper upscaling (`waifu2x-ncnn-vulkan`), the C live-wallpaper player, rice
  publishing, voxtype dictation, GPU mux and passthrough, QMK lighting, the store's
  remote catalogue, the VM parallel downloader, the screenshot beautify editor and
  Rashin's resident engine.
- Legacy unlicensed third-party display fonts in `shell/assets/fonts/` have been removed
  and replaced with open-licensed fonts (Space Grotesk and JetBrains Mono Nerd Font).
  Other image files from the first port: the SVGs in `shell/Ui/art/`,
  `shell/modules/launcher/shared/art/hands-adam.png` and the symbolic icons in
  `shell/components/icons/`. Check each before redistributing.

### Third-party code

upstream own NOTICE credits these, and the files named still carry their credit.

- Caelestia, https://github.com/caelestia-dots/shell, GPL-3.0: the frame-blob
  renderer (`shell/Blobs-src/`) and parts of the animation and motion curves
  (`shell/Ui/Anim.qml`, `shell/Ui/Singletons/Tokens.qml`).
- OkShell, https://github.com/JohnOberhauser/OkShell/tree/87a9d09d923163cd03f396a395be3bb02b335b6e,
  by John Oberhauser, GPL-3.0: the configurable frame-bar layout, the rail and widget
  catalogue organisation, the Bar Studio workflow (`shell/hub/barstudio/`) and the
  OkShell launcher variant (`shell/modules/launcher/variants/okshell/`).
- DankMaterialShell, https://github.com/AvengeMedia/DankMaterialShell, MIT, Copyright
  (c) 2025 Avenge Media LLC: portions of the display configuration UI (the Hub's
  Displays page, `shell/hub/pages/DisplaysPage.qml`).
- skwd-wall, https://github.com/liixini/skwd-wall, MIT, Copyright (c) 2026 liixini:
  the wallpaper picker (`shell/wallpaper-ui/`, which carries its own `LICENSE`), the 39
  shader transitions in `shell/modules/wallpaper/skwd/`, and the Go port of its
  daemon's modules in `desktop/cmd/toren-wallpaper/`.
- qsbar, by HANCORE-linux, MIT, Copyright (c) 2026 HANCORE-linux:
  `shell/modules/bar/barstyles/qsbar/`, which carries its own `LICENSE`. Parts of it
  are adapted from other projects and say so in comments (Shibumi-Shell's
  `SearchEngine.js`).

The Hub's Credits page lists credited projects, including those whose code
did not come across.

The MIT-licensed portions above are under this licence, each with its own copyright line:

```text
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## The store backport: upstream 1.0.7

The store — backend, window and local catalogue — was itself synced against a
newer upstream. Nothing else in this tree changed hands in that pass.

- Taken from: upstream commit `f7197d72de8a21e44b0bea922bdd915f28f6b327`
  (`release/1.0.7`, `VERSION` 1.0.7). Upstream is GPL-3.0, and so is everything
  that came across here, except the portions named under "Third-party code".
- What came across: the catalogue revision and the `check` verb (the refresh
  dot lights only when upstream advances), the window-manager gate (the
  `toren wm status` seam, and installs refused for items authored for another
  provider), the metadata fields a catalogue item may carry (initials, kind,
  licence, stars, verification, listed-at, required package managers,
  unavailable edition), the 20 second search timeout, and the store window's
  helpers for all of that (the UPDATE action, the A-to-Z category sort, status
  tags, initials, the sidebar's plugin kind and the search index).
- What did not come across: the remote registry and its catalogue, the extras
  actuator and the receipt transactions, the Flatpak provider, the
  sub-shell categories, and the snapshot cache. A plugin is a folder installed
  by `toren plugin`, and the store's plugins category delegates to it; the
  catalogue is the local `catalog/` JSON, not a registry.
- Where: `desktop/cmd/toren-store/` and `shell/store/` here, and the standalone
  `toren-store` repository, which carries the same backend and window files and
  is the copy the ISO builder prefers.
- `desktop/catalog/` in this tree carries the same local catalogue JSON so a
  desktop-tree build finds a shop where the standalone checkout is absent.
