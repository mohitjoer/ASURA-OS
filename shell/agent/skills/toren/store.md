# Installing apps, themes, bar styles and plugins

`asura-store` is Toren's store, with a window and a command line. It installs
Void packages (XBPS), Flatpaks, and Toren catalogue items (themes, bar styles,
plugins).

## Commands

```sh
asura-store --help
asura-store check                 # which providers (xbps, flatpak) are available
asura-store catalog               # the catalogue as JSON
asura-store search QUERY          # search the catalogue and the package sources
asura-store install ID            # install an item
asura-store remove ID             # remove an item
asura-store open                  # open the Store window
asura-store settings              # the Store's settings
```

Look an item up (`search`, `catalog`) before installing, and use its exact id.

## Bundles

`asura bundle` lists curated app sets; `asura bundle show NAME` says what one
installs; `asura bundle install NAME` installs it (a snapshot first, missing
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
off until switched on; `asura enable NAME` installs the package if needed,
starts the service (and at every boot), and adds the user to its group:

```sh
asura enable --list               # each one, on or off
asura enable docker               # then log out and in for the docker group
asura disable docker              # stop it; the package stays
```

Installing packages through Toren's helper (`asura-pkg-add`) takes a system
snapshot first, so a bad install can be undone with
`sudo asura rollback --system NAME` (`asura snapshot` lists them).

## Afterwards

A new app appears in the launcher at once. A theme or bar style may need
`asura refresh` (or a logout) to take effect. A desktop that misbehaves after
an install: see [`diagnostics.md`](diagnostics.md).
