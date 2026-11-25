# Updates and going back

```sh
asura update            # Toren and the desktop
sudo asura update       # also every system package, with a snapshot first
asura update --check    # is there anything new?
```

Toren only installs versions that passed its automatic tests, so an update
that is still being tested is not offered yet. The desktop picks up an update
without logging out. A notice appears when an update is available, and when a
reboot would finish one (a new kernel).

## Channels

```sh
asura channel get
sudo asura channel set stable     # releases (main)
sudo asura channel set unstable   # every change as soon as it passes (dev-unstable, the default)
```

## Firmware

`asura firmware` lists firmware updates for your laptop, SSD and devices;
`asura firmware update` installs them.

## Going back

- `toren rollback`: your desktop settings as they were before the last change.
- `asura snapshot`: the system snapshots, with the date, why each was taken
  (before an update, before installing apps, by hand) and which Toren it holds.
- `sudo asura rollback --system NAME`: boot the system as it was in that
  snapshot (from the next restart). Your files in /home are not touched.
- `sudo asura snapshot create`: take one by hand before you try something.
