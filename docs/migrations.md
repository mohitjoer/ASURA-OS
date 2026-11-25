# Migrations

Most changes reach an installed machine by themselves: `toren update` copies
the new files and rebuilds the tools. Some do not. When an older machine has to
**do** something once to reach the new state (re-key a package database entry,
fix the ownership of old files, move a setting to a new name), that is a
migration.

A migration is a small shell script in `migrations/`. `toren update` runs every
migration the machine has not run yet, records each one that succeeds, and
never runs it again. A fresh install records all of them as done, because it
already starts in the new state.

## When to write one

| The change                                                       | Where it goes                                       |
| ---------------------------------------------------------------- | --------------------------------------------------- |
| A new or changed file ASURA ships                                | Just change it; `apply-repo.sh` installs it         |
| A new package                                                    | `system/packages-void.yaml`; the update installs it |
| A fix existing machines need once (state left by an older ASURA) | A migration                                         |

Do not add one-time fixes to `apply-repo.sh` or `toren update`: they would run
on every update forever, and a fresh install would run them for nothing.

## Writing one

```sh
touch "migrations/system/$(date +%s).sh"     # or migrations/user/
```

- **The name** is the Unix time it was written, then `.sh`. Migrations run in
  that order.
- **Line 1** is `#!/bin/sh`. **Line 2** is `echo "what it does"`, one plain
  sentence; `toren migrate --list` and the update show it.
- **Safe to run twice.** Write it as "if the old state is there, fix it". If
  it fails part way, it runs again on the next update.
- **Exit 0** when done (including "nothing to do here"). Any other exit marks
  it failed: the run stops there, `toren doctor` reports it, and the next update
  tries it again.
- **Never edit a merged migration.** Machines that ran it will not run it
  again. Fix a wrong one with a new one.

`migrations/system/` runs once per machine, as root. Do not follow symlinks in
users' folders, which they control. `migrations/user/` runs once per user, as
that user, with their `HOME`.

## Where things are on a machine

| What          | Where                                        |
| ------------- | -------------------------------------------- |
| The scripts   | `/usr/share/toren/migrations/{system,user}/` |
| Done (system) | `/var/lib/toren/migrations/<name>`           |
| Done (user)   | `~/.local/state/toren/migrations/<name>`     |
| A failure     | `<name>.failed` beside the done files        |
| Output        | `migrate.log` beside the state folder        |

```sh
toren migrate --list       # every migration: done, pending or FAILED
toren migrate --pending    # what the next update would run
toren migrate              # run them now (system ones through sudo)
```

## Testing one

1. Reproduce the old state in a temporary folder (or a VM), and point the
   script at it.
2. Run it twice: the first run fixes, the second changes nothing.
3. Run it where the old state is absent: it must exit 0 and change nothing.
4. `go test ./cmd/toren` (in `desktop/`) checks every migration's name, first
   two lines, unique time and syntax. CI also runs
   `shellcheck -S warning migrations/*/*.sh`.
