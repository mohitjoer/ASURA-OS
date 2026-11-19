# One-Time Migrations in ASURA OS

Some updates require a state transition on an *already-installed* machine to move from an older ASURA state to a newer one (for example, re-keying a package database entry or repairing file ownership after an update).

On a fresh install, the machine starts in the current state and does not need these fixes. Migrations solve this: each one-time fix is an independent shell script in `migrations/`, executed by `asura update` (via `asura migrate`) for any migration the machine has not run yet. Once a migration succeeds, its completion is recorded and it is never executed again.

---

## When to write a migration

Write a migration when:
- An already-installed machine needs a one-time adjustment that a fresh install does not need (for example, re-indexing an existing database, repairing stale ownership, or converting legacy local state).
- The fix should run exactly once per machine or once per user account.

Do **not** write a migration when:
- The change is a persistent file or configuration deployed with every update (put those in `installation/apply-repo.sh` or the session configuration).
- The operation must happen on every update cycle (put those in `asura update` or `apply-repo.sh`).

---

## Structure and Locations

Migrations live in two directories in the repository:

| Directory | Scope | Execution | State marker directory |
|---|---|---|---|
| `migrations/system/` | Machine-wide | Runs once per machine as `root` (via `sudo`) | `/var/lib/asura/migrations/<name>` |
| `migrations/user/` | Per-user | Runs once per user as that user (`SUDO_USER` when under `sudo`) | `~/.local/state/asura/migrations/<name>` |

At install time, migrations are copied to `/usr/share/asura/migrations/{system,user}`.

### Fresh Installs

On a fresh ISO install, `build-iso-void.sh` marks every shipped migration as done by writing marker files into `/var/lib/asura/migrations/` (system) and `/etc/skel/.local/state/asura/migrations/` (user), as well as the initial live user directory. Fresh machines therefore start with all shipped migrations recorded as completed and will only run migrations created in subsequent updates.

---

## File Naming and Header Rules

1. **File name**: A Unix timestamp with 10 or more digits followed by `.sh` (e.g. `1795180000.sh`). Scripts run in strict numeric order. Generate a new timestamp using `date +%s`.
2. **Line 1**: Must be exactly `#!/bin/sh`.
3. **Line 2**: Must be `echo "One sentence: what it does"`. The runner (`asura migrate`) displays this sentence when running migrations and in `--list` / `--pending` reports.
4. **Permissions**: Scripts must be executable (`chmod 0755`).

Example:
```sh
#!/bin/sh
echo "Fix ownership of user state directories"

for homedir in /home/*; do
    [ -d "$homedir" ] || continue
    u=$(basename "$homedir")
    id -u "$u" >/dev/null 2>&1 || continue
    if [ -d "$homedir/.local/state/asura" ]; then
        chown -R "$u:$u" "$homedir/.local/state/asura"
    fi
done
```

---

## Core Principles

### Safe to run twice (Idempotency)

Every migration must be written defensively: "if the old state is present, fix it; otherwise do nothing". A migration that encounters an error or gets interrupted will exit non-zero, leaving a `<name>.failed` marker. On the next update, it will be executed again from the start.

### Never edit a merged migration (Immutability)

Once a migration is merged into `dev-unstable`, it must **never** be edited. Machines that already ran the original migration have marked it as done and will never run an edited version. Any correction or follow-up must be introduced as a new migration with a newer timestamp.

---

## CLI Usage

The `asura migrate` command manages migrations:

```bash
asura migrate                     # Run pending system migrations (as root) then user migrations
asura migrate --system            # Run only system migrations (prompts for sudo if unprivileged)
asura migrate --user              # Run only user migrations
asura migrate --pending           # Show pending migrations that would run without executing them
asura migrate --list              # List all migrations with status (done, pending, or FAILED)
asura migrate --mark-done         # Record all migrations as done without executing them
```

Logs are written to `migrate.log` beside the respective state directory (`/var/lib/asura/migrate.log` for system migrations and `~/.local/state/asura/migrate.log` for user migrations).

System health checks via `asura doctor` inspect migrations and report pending migrations (informational) or FAILED migrations (causing doctor to fail).

---

## Testing Migrations

When developing a migration, test it using environment overrides without touching real system paths or running as root:

```bash
# Point to a temporary migrations tree and state directory
export ASURA_MIGRATIONS_DIR="/path/to/test/migrations"
export ASURA_MIGRATION_STATE="/path/to/test/sys_state"
export XDG_STATE_HOME="/path/to/test/user_state"

# Run tests
asura migrate --pending
asura migrate --user
asura migrate --list
```

The Go test suite in `desktop/cmd/asura/migrate_test.go` automatically verifies:
- Numeric sorting and filtering of ignored files
- Successful runs and marker creation
- Stop-on-failure behavior, `.failed` marker generation, and recovery
- Flag behavior (`--pending`, `--list`, `--mark-done`)
- Syntax validation (`sh -n`) and naming rules for all repository migrations
