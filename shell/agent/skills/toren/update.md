# Updating Toren

`asura update` brings the machine to the newest Toren. "The newest" means the
newest commit on the `dev-unstable` branch **that passed CI**, not simply the
newest commit.

## Commands

```sh
asura update --check        # report only: "up to date" or "update available: old -> new"
asura update                # Toren (no root needed; sudo is asked for when it installs)
sudo asura update           # also the Void system packages (XBPS), with a Btrfs snapshot
asura update --wait         # wait for CI still running on the newest commit
asura update --force        # take the newest commit without checking CI (risky)
asura version --commit      # the commit this machine runs
```

`--check` exits 0 when up to date and 3 when an update is available.

## What an update does, in order

1. **System packages** (only `sudo asura update` or `--system`): `xbps-install
   -Su`. When a third-party repository lags a library change ("broken,
   unresolvable shlib"), the affected packages are held for that run and
   released afterwards; everything else updates.
2. **Toren**: fetch `dev-unstable` (30 s limit; offline it uses the last
   fetch), choose the newest commit whose GitHub checks all passed, and apply
   it with `installation/apply-repo.sh` (files, Go tools, new packages). A
   commit older than the one applied is never applied. What changed is listed.
3. **Flatpaks**.
4. **Refresh**: the running desktop picks up the change without a logout
   (`asura refresh`: user services re-synced, Hyprland reloaded, the shell and
   Toren's daemons restarted).

## Reading the output

- "checks for the newest commit X are still running; taking the last commit
  that passed, Y" — normal; run again later or `--wait`.
- "the newest commit X failed CI" — the branch is broken upstream; the machine
  stays on the last good commit. `--force` would install the broken one.
- "could not read CI results from GitHub" — offline or rate-limited; nothing is
  applied. `GITHUB_TOKEN` raises the rate limit.
- "the installed Toren (X) is newer than Y" — nothing to do.

## Channels, firmware, cleanup

- `asura channel get`; `sudo asura channel set stable` follows releases
  (main), `unstable` follows dev-unstable (the default). Both only take
  commits that passed CI.
- `asura firmware` lists firmware updates (fwupd); `asura firmware update`
  installs them.
- `asura clean` removes orphaned packages and the package cache (XBPS asks
  first).
- A desktop notice says when an update is available or a reboot would finish
  one (the `toren-updates` user service runs `asura notify-updates`).

## Going back

- `asura rollback` — the previous config revision (Toren keeps the three newest
  in `~/.local/state/asura/revisions/`).
- `asura snapshot` — list system snapshots: when, why (before an update, before
  installing packages, by hand) and which Toren each holds.
  `asura snapshot create` takes one by hand.
- `sudo asura rollback --system <snapshot>` — make a pre-update Btrfs snapshot
  the root at the next reboot (`asura rollback --system` lists them; the five
  newest pre-update snapshots are kept). The old root is kept until deleted.

## Downloads are slow

`asura mirror fastest` times the Void mirrors and uses the quickest;
`asura mirror list` shows them, `asura mirror set NAME` picks one,
`asura mirror reset` goes back to the default. Updates keep the choice.
