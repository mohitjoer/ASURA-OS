package main

// The per-subcommand usage texts. They live in one file so that a flag added to
// a subcommand and the text describing it are edited in the same place.
const reportUsage = `Usage: asura hardware report [flags]

Prints this machine's hardware, redacted. The output is safe to paste into an
issue: serials appear as stable tokens, never in the clear.

  --format text|json|profile   output shape (default text)
  --profile                    shorthand for --format profile
  --deep                       include ACPI-derived facts
  --submit                     prepare redacted report and open prefilled GitHub issue in browser
  --raw                        unredacted dump, local only; refuses under CI
  --out FILE                   write to a file instead of stdout
  --root DIR                   read this sysfs tree instead of /sys
  --profiles DIR               profile database directory
`

const matchUsage = `Usage: asura hardware match [flags]

Reports the profile that claims this machine, its support tier, and the
candidates that matched but lost. Exit code 3 means no profile matches, or two
profiles tie and disagree — an answer, not a failure.

  --root DIR        read this sysfs tree instead of /sys
  --profiles DIR    profile database directory
`

const lintUsage = `Usage: asura hardware lint-profiles [flags]

Checks every profile against the schema rules in system/hardware/README.md, and
warns about pairs of profiles that can select the same machine — the shape that
becomes a conflict once a third arrives.

  --strict          exit non-zero on a warning as well as an error
  --profiles DIR    profile database directory
`

const captureUsage = `Usage: asura hardware capture --out DIR [flags]

Runs the same readers a live machine does and writes the files they consume, so
a fixture is a machine. Captures are redacted as they are written, so a committed
fixture cannot contain a serial.

  --out DIR     required; the directory to write
  --root DIR    read this sysfs tree instead of /sys
`

const updateUsage = `Usage: asura update [flags]
       sudo asura update [flags]

Updates Toren desktop packages and configuration. When run as a normal user,
it updates Toren packages, overlays, and git code. When run as root (or with sudo),
it also updates all base system packages (XBPS) and creates a Btrfs snapshot.

  --root DIR       operate on a tree under this directory
  --dry-run        do the snapshot and preserve steps, fetch nothing
  --no-snapshot    skip the config snapshot
  --unstable       pull latest code changes from the dev-unstable branch
  --branch NAME    specify a git branch to pull code updates from
  --channel NAME   update channel (stable, unstable)
  --system         update all system packages (XBPS) even when unprivileged
  --asura-only     only update Toren packages and components (skip XBPS)
  --check          only report whether a newer Toren is available (exit 3 if so)
  --force          take the tip of the branch without checking CI, and reapply
                   even when it is already the applied commit
  --wait           wait for CI still running on the newest commit instead of
                   taking the last one that passed

Order: system packages (as root), then Toren from its branch (dev-unstable by
default), Flatpaks, and a refresh of the running desktop.

Toren follows the branch safely: the fetch is bounded and never prompts (offline,
the copy fetched before is used); the target is the newest commit whose GitHub
checks all passed, not the tip; a target older than the applied commit is never
applied; the update's own clone resets to the target, while a developer's
checkout is only fast-forwarded and keeps local work. What changed is listed and
the checkout is applied by installation/apply-repo.sh, which records the commit
and the checkout it came from. Nothing new means nothing is rebuilt.
`

const rollbackUsage = `Usage: asura rollback [flags] [snapshot-name]

Restores configuration overlay (--config) or restores Btrfs root snapshots (--system).

Modes:
  --config             revert ~/.config overlay to previous revision (default)
  --system             list or restore Btrfs root snapshots
  --list               list available config revisions and system snapshots

Flags:
  --root DIR           operate on a tree under this directory
  --force              (--system) restore even when the snapshot has no modules
                       for the installed kernel

A system restore renames the current root subvolume (@) aside as
@pre-rollback-<time>, makes a writable copy of @snapshots/<name> the new @,
and takes effect at the next reboot; fstab and GRUB already mount @. The old
root is kept until you delete it. /boot is its own partition and is not
rolled back.
`

const materializeUsage = `Usage: asura materialize [flags]

Re-lays the whole base into the config tree, clobbering shipped content. User
edits are preserved. For one file, use ` + "`asura reset <path>`" + `.

  --root DIR   operate on a tree under this directory
`

const resetUsage = `Usage: asura reset <path> [flags]
       asura reset --desktop [flags]
       asura reset --restore BACKUP [flags]

Drops user edits and restores defaults.

  <path>              drop one user edit and restore that file from the base
  --desktop           reset desktop configuration back to Toren defaults (desktop config.json,
                      Hyprland user files, shell state), backing up current files to a dated folder
  --restore BACKUP    restore a previous desktop configuration backup byte-for-byte
  --root DIR          operate on a tree under this directory
`

const reinstallUsage = `Usage: asura reinstall [flags]

Re-apply the Toren files for the commit already installed, to repair deleted
or edited shipped files without updating to anything newer.

  --root DIR   operate on a tree under this directory
`

const recoveryUsage = `Usage: asura recovery

A guided text menu that works on a TTY (no desktop needed) to diagnose and
repair system issues: run doctor, reset desktop configuration, reinstall
Toren files, list and restore snapshots, launch safe-mode session, or inspect logs.
`

const doctorUsage = `Usage: asura doctor [flags]

Reports drift between the base and the config tree. It changes nothing unless
asked to: --fix renames the values Toren has renamed in your shell.json.

  --root DIR   operate on a tree under this directory
  --fix        apply the renames the Config section offers
  --json       print the report as JSON: {"ok", "sections": [{"section", "ok", "lines"}]}
`

const statusUsage = `Usage: asura status [flags]

The current revision, the paths in use, how much drift there is, and whether
the shell and the daemon are up.

  --root DIR   operate on a tree under this directory
  --json       the update report the shell reads, as one line of JSON
`

const bindsUsage = `Usage: asura binds [flags]

Prints the keybinding catalogue as JSON: every chord, with a label and a
category. It is the same catalogue the compositor's configuration is checked
against, and the shell draws it as its cheat sheet.

  --root DIR   operate on a tree under this directory
`
