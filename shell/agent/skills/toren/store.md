# Installing apps, themes, bar styles and plugins

`toren-store` is Toren's store, with a window and a command line. It installs
Void packages (XBPS), Flatpaks, and Toren catalogue items (themes, bar styles,
plugins).

## Commands

```sh
toren-store --help
toren-store check                 # which providers (xbps, flatpak) are available
toren-store catalog               # the catalogue as JSON
toren-store search QUERY          # search the catalogue and the package sources
toren-store install ID            # install an item
toren-store remove ID             # remove an item
toren-store open                  # open the Store window
toren-store settings              # the Store's settings
```

Look an item up (`search`, `catalog`) before installing, and use its exact id.

## Bundles

`toren bundle` lists curated app sets; `toren bundle show NAME` says what one
installs; `toren bundle install NAME` installs it (a snapshot first, missing
package names skipped). Bundles: gaming, dev, editors, apps, creative, office.

## Plain packages

For a package the catalogue does not list, the system's package manager works
directly; installing needs root, so say what you would run first:

```sh
xbps-query -Rs NAME               # search the repositories
sudo xbps-install -S NAME         # install
sudo xbps-remove NAME             # remove
```

`apm` is Toren's shorthand for the same verbs (`apm install NAME`).

## Optional services

Docker, Bluetooth, printing, SSH, virtual machines (libvirt) and Tailscale are
off until switched on; `toren enable NAME` installs the package if needed,
starts the service (and at every boot), and adds the user to its group:

```sh
toren enable --list               # each one, on or off
toren enable docker               # then log out and in for the docker group
toren disable docker              # stop it; the package stays
```

Installing packages through Toren's helper (`toren-pkg-add`) takes a system
snapshot first, so a bad install can be undone with
`sudo toren rollback --system NAME` (`toren snapshot` lists them).

## Afterwards

A new app appears in the launcher at once. A theme or bar style may need
`toren refresh` (or a logout) to take effect. A desktop that misbehaves after
an install: see [`diagnostics.md`](diagnostics.md).
