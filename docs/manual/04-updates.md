# Updates and going back

```sh
toren update            # Toren and the desktop
sudo toren update       # also every system package, with a snapshot first
toren update --check    # is there anything new?
```

Toren only installs versions that passed its automatic tests, so an update
that is still being tested is not offered yet. The desktop picks up an update
without logging out. A notice appears when an update is available, and when a
reboot would finish one (a new kernel).

## Channels

```sh
toren channel get
sudo toren channel set stable     # releases (main)
sudo toren channel set unstable   # every change as soon as it passes (dev-unstable, the default)
```

## Firmware

`toren firmware` lists firmware updates for your laptop, SSD and devices;
`toren firmware update` installs them.

## Going back

- `toren rollback`: your desktop settings as they were before the last change.
- `toren snapshot`: the system snapshots, with the date, why each was taken
  (before an update, before installing apps, by hand) and which Toren it holds.
- `sudo toren rollback --system NAME`: boot the system as it was in that
  snapshot (from the next restart). Your files in /home are not touched.
- `sudo toren snapshot create`: take one by hand before you try something.
