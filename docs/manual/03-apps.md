# Installing apps

## The Store

Open **Store** from the launcher to browse and install apps, themes and
plugins. From a terminal: `toren-store search NAME`, `toren-store install ID`.

## Bundles: many apps in one command

```sh
asura bundle                 # the list
asura bundle show gaming     # what one installs
asura bundle install gaming  # install it
```

| Bundle | Installs |
|---|---|
| gaming | Steam, Lutris, Heroic, RetroArch, MangoHud, GameMode, Gamescope, Wine, 32-bit graphics drivers for your GPU |
| dev | C/C++, Go, Rust, Node.js, Python, Docker with Compose, git tools |
| editors | Helix, Emacs, Visual Studio Code |
| apps | Signal, Discord, Spotify, Obsidian, Dropbox, Bitwarden, Tailscale |
| creative | GIMP, Inkscape, Krita, Blender, OBS Studio, Kdenlive |
| office | LibreOffice, Xournal++, Simple Scan, printing |

A snapshot is taken first, so `sudo asura rollback --system` can undo it.

## Single packages

```sh
xbps-query -Rs NAME          # search
sudo xbps-install -S NAME    # install
sudo xbps-remove NAME        # remove
flatpak install flathub ID   # an app from Flathub
```

`asura clean` removes packages nothing needs any more and the download cache.
