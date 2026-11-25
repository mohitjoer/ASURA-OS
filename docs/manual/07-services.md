# Services, firmware and downloads

## Optional services

Off until you need them, one command each:

```sh
asura enable --list        # each one, on or off
asura enable docker        # also: bluetooth, printing, ssh, libvirt, tailscale
asura disable docker
```

Enabling installs what is missing, starts the service now and at every boot,
and adds you to its group (log out and in once for that).

## Tailscale File Sharing

When Tailscale is enabled (`toren enable tailscale`), share files directly across your devices:

```sh
asura share devices                # list available Tailscale devices
asura share send file.tar.gz phone # send a file via Taildrop
asura share receive                # fetch incoming files into ~/Downloads
```

Incoming files trigger a desktop notification when received.

## User Extension Hooks

You can run custom scripts when Toren lifecycle events occur by placing executable scripts in:

```sh
~/.config/asura/hooks/<event>/
```

Supported events:
- `pre-update`: runs before system and package updates begin
- `post-update`: runs after update completes
- `post-theme`: runs when a new wallpaper or look is applied
- `post-login`: runs when logging in to the graphical desktop session

Hooks run in alphabetical order as your user with a 30s timeout, logged to `~/.local/state/asura/hooks.log`.
A failing hook is reported but never aborts Toren's own operations.
Environment variables passed: `ASURA_EVENT`, `ASURA_HOOK_DIR`, `ASURA_USER`.

## The firewall

Incoming connections are blocked; everything you start (browsing, updates)
works. Services you enable that listen for connections (SSH, for example) may
need a rule in `/etc/nftables.conf`.

## Faster downloads

```sh
asura mirror fastest       # time the mirrors and use the quickest
asura mirror list
asura mirror reset
```
