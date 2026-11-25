<p align="center">
  <img src="public/toren-logo.svg" alt="Toren Linux" width="180">
</p>

# Toren Linux

<p align="center">
  <img src="public/toren-logo-on-dark.svg" alt="" width="96">
</p>

**untouchable · dominating**

*Ten* is heaven, the highest point of the sky. An asura is that same point. The
mark and the name say one thing, and they say it about the architecture as much as the
project: this is the top of the stack, and it sits above everything it is built on.

Nothing reaches up to touch it. That is the standard the project holds itself to, and it is
why the base layer is treated as substrate and never as identity — see
[the base-layer boundary](docs/architecture.md#the-base-layer-boundary).

Toren is a Linux distribution. Its **primary base layer** is Void Linux x86_64 (glibc). Everything above
that line is Toren's own — its compositor configuration, its shell drawn in Quickshell, its
control tools, its hardware matrix, its updater, and its supervised session daemons.

Toren is not a Void spin. Void supplies the minimalist, lightning-fast substrate and nothing else. Toren is the
distribution; Void Linux is what Toren is built on top of.

> **Status: Beta.** Live bootable ISO images (`toren-linux-<version>-x86_64.iso`, published on the [Releases](https://github.com/mohitjoer/Toren-Linux/releases) page) build cleanly in CI and run a fully functional live environment with automated Wayland desktop login. The guided desktop installer (`toren install` / `toren-installer.desktop`) partitions disks, installs rootfs, and configures GRUB EFI.

---

## Install

Toren is in beta. Install it on a machine you can afford to wipe, and back up anything you
care about first.

### What you need

| | |
|---|---|
| CPU | 64-bit x86 (x86_64), Intel or AMD |
| Memory | 4 GB minimum, 8 GB or more recommended (the live session boots in 2 GB) |
| Storage | A whole disk for Toren (the installer erases it), 32 GB or more recommended, SSD preferred |
| Firmware | UEFI or legacy BIOS. **Secure Boot must be turned off** in the firmware settings; it is [not supported yet](docs/system-services.md#secure-boot) |
| USB stick | 4 GB or larger, for the installer image |
| Network | Optional for the install, needed for updates |

Check your laptop in the [hardware compatibility matrix](docs/hardware.md) first. Machines without a
profile get conservative defaults (support tier 3) and may need manual fixes.

### 1. Download

Get these files from the latest release on the
[Releases page](https://github.com/mohitjoer/Toren-Linux/releases):

- `toren-linux-<version>-x86_64.iso`: the installer image
- `SHA256SUMS`: its checksum
- `SHA256SUMS.sig`: a minisign signature of the checksums, when the release is signed

### 2. Verify the download

In the folder holding both files:

```console
$ sha256sum -c SHA256SUMS
toren-linux-<version>-x86_64.iso: OK
```

Anything other than `OK` means the download is corrupt or has been tampered with. Download it
again, and don't write it to a USB stick.

If the release includes `SHA256SUMS.sig`, also check the signature with
[minisign](https://jedisct1.github.io/minisign/) and the project's public key, as published
on the release page:

```bash
minisign -Vm SHA256SUMS -x SHA256SUMS.sig -P <public key>
```

### 3. Write the image to a USB stick

Any of these work:

- **Linux, `dd`.** Find the stick with `lsblk` and be certain it is the right device: `dd`
  overwrites whatever you point it at.

  ```bash
  sudo dd if=toren-linux-<version>-x86_64.iso of=/dev/sdX bs=4M status=progress oflag=sync
  ```

- **[Ventoy](https://www.ventoy.net/)**: copy the ISO onto a Ventoy stick and pick it from
  Ventoy's menu.
- **[Fedora Media Writer](https://github.com/FedoraQt/MediaWriter)** (Windows, macOS, Linux):
  choose "Select .iso file" and pick the Toren ISO.

### 4. Boot the live session

1. Turn off Secure Boot in the firmware settings.
2. Boot from the USB stick (usually F12, F11, F8 or Esc at power-on opens the boot menu).
3. The live session logs in automatically as `toren` and starts the desktop. Nothing is
   written to your disk yet, so check that display, audio, Wi-Fi and the keyboard work.

### 5. Install

Open **Install Toren Linux** from the launcher, or run this in a terminal:

```bash
asura install
```

The installer walks you through choosing the disk, the filesystem (Btrfs with snapshots, or
ext4), full-disk encryption (LUKS2, on by default), your user account and hostname. It shows a
summary and asks you to type `YES` before it erases anything. `toren install --dry-run` shows
what it would do without touching the disk.

When it finishes, reboot and remove the USB stick. If you chose encryption, you'll be asked for
the disk passphrase at boot.

After installing, `toren hardware report --profile` produces a profile of your machine that you
can [contribute](CONTRIBUTING.md), so the next person with the same laptop gets a verified setup.

---

## What Toren is

Most Linux desktops are a distribution's defaults plus whatever you can find. Toren is a
single coherent opinion: a curated package set, one design language applied everywhere, a
hardware profile for your specific machine, and an updater that can roll back.

| Layer | Choice |
|---|---|
| Base layer | Void Linux x86_64 (glibc) — the substrate only |
| Init System | runit (system daemons) + user-level runit supervisor (`asura-user-services`) |
| Seat Management | seatd (`LIBSEAT_BACKEND=seatd`) / elogind |
| Compositor | Hyprland (Wayland) |
| Desktop shell | Quickshell (QML) — bar, dock, launcher, popouts, widgets. The engine is built by Toren; the surfaces are vendored open-source components, re-themed to Toren's monochrome design language |
| Control tools | Go — `toren`, `toren-sysd` (root daemon at `/run/asura/sysd.sock`), `toren-userd` (session daemon), `toren-shell`, `toren-hubd`, `toren-wallpaper`, and `apm` |
| Settings store | GSettings schemas + dconf, one writer per setting |
| GPU & Power | `toren gpu` (PCI inspection & `prime-run` offload), `toren power` (profiles & NVIDIA RTX 3050 D3cold management) |
| Terminal | Ghostty — strictly supported, configured by Toren, never kitty |
| User shell | zsh + starship |
| Editor | Neovim |
| Boot chain | GRUB EFI |
| Installer | `toren install` / `toren-installer.desktop` guided system installer |
| Desktop services | pipewire + wireplumber, tlp, NetworkManager, seatd |
| Portals | `xdg-desktop-portal-hyprland` + `-gtk`, and `-lxqt` for Secret interface |
| Package manager | XBPS, with first-party repository `asura-xbps-repo`, plus `apm` package manager frontend |
| Software Store | [Toren Store](https://github.com/Toren-Linux/asura-store) — Graphical showroom window & Go data plane (`toren-store`), multi-provider XBPS/plugin/theme catalogue |
| Resource Monitor | `btop` — Real-time telemetry, process manager & task manager with Toren adaptive color theming |
| Agent layer | OpenCode with the Toren skills wired in (`toren-agent`) |

### Why Void Linux is the base layer

Void Linux provides a clean, bloat-free foundation:

1. **glibc x86_64**: Complete binary compatibility with NVIDIA proprietary drivers, Steam, Electron, and modern developer toolchains.
2. **runit init system**: Fast, predictable service supervision without systemd overhead, paired with Toren's user-level session supervisor (`asura-user-services`).
3. **XBPS package manager**: Fast package installation and native binary package generation (`xbps-create`, `xbps-rindex`).
4. **Hardware reach**: Direct support for hybrid graphics offloading (`prime-run`) and fine-grained power management.

### No GNOME

Toren contains no GNOME — not GNOME Shell, not the GNOME daemons, not the GNOME portal backend.
It draws its own shell on Hyprland and owns the surfaces a stock desktop used to provide:
notifications, global shortcuts, locking, idle handling, and quick settings.

What this costs: Toren writes those surfaces itself. What it simplifies: there is no
`xdg-desktop-portal-gnome` to convince, and no GNOME daemon stack to fight. See
[`docs/architecture.md`](docs/architecture.md) for the portal arrangement, and
[`agents/agents.md`](agents/agents.md) for the archived GNOME Shell notes.

---

## Hardware support

Toren publishes support tiers rather than claiming universal support, because no Linux
distribution supports every laptop. Every machine falls into one of these.

| Tier | Meaning |
|---|---|
| **1 — Verified** | Installs, boots, suspend/resume, GPU acceleration, audio and peripherals confirmed by a maintainer. |
| **2 — Community confirmed** | Working on a contributor's machine, with a profile submitted. Not independently verified. |
| **3 — Best effort** | No matching profile. Installs conservative defaults: the `generic` kernel, full firmware bundle. May need manual intervention. |

The `generic` in that row is not prose. It is the same value as `default_for_unknown` in
[`system/kernel/fleet.yaml`](system/kernel/fleet.yaml), which is the authority for which kernel
a machine gets, and where the two disagree that file is right.

The matrix grows through [`system/hardware/`](system/hardware/) — a database of verified
machine profiles in this repository. Nobody can test every laptop, but every user can test
one. A machine that boots successfully emits a candidate profile via `toren hardware report`;
contributors attach it to a pull request; merged profiles ship in the next ISO and are
selected automatically at install time.

This is the single most important mechanism in Toren. It turns an unbounded hardware problem
into a union of working setups that grows without anyone owning a hardware lab.

---

## Repository layout

```
apps/                built-in desktop applications (Hub, Store, Sysmon, VM)
desktop/             the control tools: cmd/toren, toren-sysd, toren-userd, toren-shell, toren-hubd, apm
system/hardware/     detection, matching, redaction, reports, profiles and tiers
system/kernel/       pinned kernel fleet
installation/        installer backend (toren-installer.sh) and ISO builder (build-iso-void.sh)
packaging/           packaging infrastructure: packaging/xbps/ with binary repo generator
session/             session startup, Hyprland configs, binds, SDDM greeter theme, GRUB snapshots
shell/               Quickshell QML desktop shell (bar, dock, launcher, popouts, Hub)
system/              system services (runit), hardware detection, kernel fleet, packages manifest
docs/                guides, roadmap, and architecture records
agents/              the agent contract and the design system
tests/               contract tests and CI suites
public/              branding assets, wallpapers, and logos
```

The desktop source lives under `desktop/`, never in a directory named after one of the
control tools — a `asura/` directory collides with the built binary of the same name and
silently defeats the repository's own ignore rules.

The four design documents, and what each one decides:

| Document | Decides |
|---|---|
| [`docs/architecture.md`](docs/architecture.md) | The stack and the reasoning behind it — settled decisions, rejected options, the base-layer boundary, the portal seam |
| [`docs/hardware.md`](docs/hardware.md) | Hardware compatibility matrix, supported laptop families, support levels, and submission workflow |
| [`docs/hardware-detection.md`](docs/hardware-detection.md) | The subsystem that decides whether Toren works on someone else's laptop — probes, matching, PII stripping, tier assignment |
| [`docs/session.md`](docs/session.md) | The boot and login chain — GRUB through runit to Hyprland, the session entry point, what starts in what order, and who owns logout, suspend and shutdown |
| [`docs/menu.md`](docs/menu.md) | `toren menu` — the unified "do everything" menu on Super+Alt+Space coordinating apps, setup, themes, updates and power |
| [`docs/adopting-toren.md`](docs/adopting-toren.md) | `toren adopt` (Legacy — Ubuntu/APT design) — historical in-place adoption design; current installation uses the live ISO and `toren install` |

[`docs/roadmap.md`](docs/roadmap.md) orders the work.

Detection is code, and it lives beside its data. `system/hardware/` holds both: the tiers and
profiles the probes resolve against, and the Go that reads them —

```
system/hardware/
  probe/               five sysfs readers, each taking a filesystem root
  match/               pure resolution: candidates, tie-breaks, conflicts
  redact/              the report redaction policy, two layers
  tiers.yaml           the three support tiers, the authority
  profiles/            the machine database, one YAML per machine
```

It lives there rather than under `desktop/` because every byte of hardware-dependent code is
behind a filesystem root, which is what makes the whole subsystem testable without a hardware
lab: `probe` readers take a root, so a captured directory tree is a machine. `TestNoAbsoluteSysfsPath`
walks the AST of package `probe` and fails on any hard-coded `/sys` or `/proc` path, so the
property cannot decay into a review comment. `system/hardware/` is its own Go module and depends
on nothing outside the standard library; `desktop/` is a second module that depends on it.

The commands that reach it:

```console
$ toren menu                         # open the unified "do everything" menu
$ toren install                      # launch the guided disk installer
$ toren hardware report              # this machine, redacted, safe to paste
$ toren hardware report --profile    # a candidate profile to review and submit
$ toren hardware match               # which profile claims this machine, and at what tier
$ toren hardware lint-profiles       # check the database
$ toren hardware capture --out DIR   # write this machine as a test fixture
```

And the overlay model, which is what makes an update safe to automate:

```console
$ toren status          # revision, paths, how much drift
$ toren doctor          # drift, plus whether the portal config is really being read
$ toren materialize     # re-lay the base, clobbering shipped content
$ toren reset PATH      # drop one of your edits back to the base
$ toren update          # snapshot, fetch, re-materialize; your edits survive
$ toren rollback        # re-materialize the previous revision
```

Every command takes `--root DIR`, so the whole surface runs against a temporary tree or a
captured machine as happily as against the live one. Exit code 3 from `match` means no profile
matched — an answer, not a failure.

The one rule the overlay model exists to keep: **a user's override wins at its mirrored path,
and no update can clobber it.**

The legacy in-place adoption design (`toren adopt`) is documented in [`docs/adopting-toren.md`](docs/adopting-toren.md); current installation is driven by the bootable ISO and `toren install`.

## Branches

| Branch | Holds |
|---|---|
| `main` | Documentation. Design, decisions, hardware framework, project rules. Always reviewable. |
| `dev-unstable` | Code. The working implementation, merged forward when a milestone completes. |

Documentation changes land on `main` directly. Implementation lands on `dev-unstable`. A
milestone is done when its code has merged forward, not when it is written.

`dev-unstable` contains the complete working implementation, including the bootable Void-based Live ISO, the guided installer (`toren install`), hardware detection, and the Quickshell desktop shell. See [`docs/roadmap.md`](docs/roadmap.md) for milestone progress.

---

## Design system

Toren's visual language is monochrome by construction: black, grey and white, with accent
colour never carrying meaning. The rules live in [`agents/agents.md`](agents/agents.md) and
are binding, not advisory:

- Strict monochrome. Traffic lights and destructive/warning/error states are the only
  exceptions, because they carry meaning a grey cannot.
- A 6/7/8/10/12/14/18/22px radius scale, a 4px spacing grid, and two-stop elevation. One
  shadow reads as a flat sticker.
- Dark and light variants declare the same properties in the same order and differ only in
  token values.
- Contrast floor: 4.5:1 for body ink, 3:1 for UI shapes and indicators.
- **Typography hierarchy**: `Space Grotesk` for display headings and section marks, `Inter` for primary UI text, and `JetBrainsMono Nerd Font` for monospace code and terminal environments.
- **Strict white mark**: `public/toren-logo-on-dark.svg` (`#ffffff`) is exclusively used across all dark UI surfaces; colored or red logos are forbidden to preserve the clean monochrome identity.

Aesthetic character comes from material, shape and rhythm — never from hue.

---

## Contributing

Toren is open source and contributions are the point. **Hardware profiles are the most valuable
contribution and the easiest for a new person to make** — you need one machine and no code.

- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how to contribute, including hardware profiles
- [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) — the behaviour expected of everyone
- [`SECURITY.md`](SECURITY.md) — reporting a vulnerability privately
- [`agents/agents.md`](agents/agents.md) — the project rules and the design system

Contributions are made under the Developer Certificate of Origin.

---

## Licence

Toren is free software, but not under one licence, and not entirely free of proprietary code on
disk. Both facts matter, so both are documented:

- Toren's own code and Toren's own repository are **GPL-3.0** and entirely free.
- The shell's **engine** is Quickshell, **LGPL-3.0-only**, shipped unmodified. The shell's
  **surfaces** are a vendored open-source Quickshell configuration —
  [`end-4/dots-hyprland`](https://github.com/end-4/dots-hyprland)'s "ii" config under
  GPL-3.0 and [`end-4/rounded-polygon-qmljs`](https://github.com/end-4/rounded-polygon-qmljs)
  under Apache-2.0 — kept byte-identical to their pinned commits and re-themed to Toren's
  design language. The composed shell is GPL-3.0.
- Some laptop hardware needs proprietary firmware. Toren references the base layer's
  `linux-firmware-*` packages rather than redistributing firmware itself, and declares per
  machine what is needed.

[`LICENSING.md`](LICENSING.md) has the full split. [`LICENSE`](LICENSE) is the GNU General
Public License v3.0.