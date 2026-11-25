# Toren Store

Toren Store is the package and customization showroom window and CLI data plane for Toren Linux. It provides browsing, discovery, and management of native packages, plugins, themes, and UI customizations, with privilege escalation safely handled through polkit.

The canonical standalone repository is hosted at **[github.com/Toren-Linux/asura-store](https://github.com/Toren-Linux/asura-store)**.
The tree in this repository keeps byte-identical mirrors of the standalone store:
- `desktop/cmd/toren-store/` mirrors `cmd/toren-store/`
- `shell/store/` and `shell/store.qml` mirror `ui/`
- `desktop/catalog/` mirrors `catalog/`

CI parity is automatically enforced on every push and PR via `.github/workflows/store-sync.yml`.

| Component | Path | Function |
|---|---|---|
| `toren-store` (`toren-store`) | `desktop/cmd/toren-store` (binary `/usr/bin/toren-store`) | The Go data plane: catalogue, search, check, install, remove, open, settings. |
| Store Window | `shell/store.qml`, `shell/store/` (`/usr/share/asura/store/ui`) | Quickshell showroom window: discover stage, search shelf, product overlays. |
| Catalog | `desktop/catalog/` (`/usr/share/asura/store/catalog/`) | Local catalogue JSON (`plugins/`, `themes/`, `barstyles/`). Updated without recompilation. |
| Launcher | `session/toren-store` (`/usr/share/asura/session/toren-store`) | Starts the store window under a single-instance lock. |

## How an Install Happens

The store UI executes `toren-store install CATEGORY ID`. The data plane outputs JSON progress events on stdout (`{"stage":"...", "percent":...}`) and status or error messages on stderr.

### 1. Applications (`apps`)
Applications are native packages.
1. The store uses native package management tools (`xbps-query`, `xbps-install`, `xbps-remove`).
2. Elevated transactions run through `pkexec`, querying the active desktop authentication agent (`toren-shell` polkit agent) for authorization.
3. Progress transitions through defined stages: `AUTHORIZING`, `DOWNLOADING`, and `INSTALLING` (or `REMOVING`).
4. If authentication is canceled or denied, `pkexec` returns cleanly and the store presents a clear authorization message.
5. `toren-store` ignores `SIGPIPE` and `SIGHUP` to ensure background package transactions finish cleanly even if the window closes.

### 2. Plugins (`plugins`)
Plugins extend Toren Linux capabilities:
1. The catalogue entry references the canonical upstream repository.
2. The store runs `toren plugin install <upstream> --yes` (removal runs `toren plugin remove <id>`), matching the Hub Add-ons manager.
3. Installed plugins reside in `~/.local/share/asura/plugins`.

### 3. Themes (`themes`)
Themes modify desktop styling and colors:
1. Installing an existing theme writes its identifier into `~/.config/asura/current/theme.name`, picked up live by the shell.
2. Removal protects the currently active theme, refusing to delete active configurations.

### 4. Bar Styles (`barstyles`)
Bar styles (`top`, `side`) are built directly into Toren Linux shell. The store provides style discovery, preview, and selection via Bar Studio.

## CLI Commands

```text
toren-store catalog [--refresh] [--category ID]    JSON: categories and items
toren-store check                                  JSON: {"revision","updateAvailable","offline"}
toren-store search QUERY                           JSON: {"items": [...]} from repository
toren-store install CATEGORY ID                    Streamed progress lines on stdout, status on stderr
toren-store remove CATEGORY ID                     Streamed progress lines on stdout, status on stderr
toren-store open SECTION                           discover | library | window section
toren-store settings CATEGORY [ID]                 Navigates Settings Hub (toren-shell hub open)
```

### Catalog Categories
- `apps`: Curated desktop applications with summaries, descriptions, and install status queryable via `xbps-query`.
- `plugins`: Antigravity and Toren Linux plugins from verified upstreams.
- `themes`: Local and curated desktop themes.
- `barstyles`: Built-in bar styles for desktop customization.

## Revisions and Update Checks

Every freshly built catalog computes a deterministic revision hash (SHA-256 over item attributes, versions, manifests, and declared links).
- Viewing the catalog acknowledges the revision at `~/.cache/toren-store/seen-revision`.
- `toren-store check` inspects the active catalog and reports `{"revision","updateAvailable","offline"}`.
- When an unseen catalog update is available, the store UI illuminates the refresh indicator.

## Search

`toren-store search <query>` queries native repository indices using `xbps-query -Rs`, ranking exact matches first, filtering system libraries and debug packages, and combining results with installed state.

## Deep Links & Settings Navigation

- `toren-store open [section]` activates or raises the single-instance store window and navigates to the target section (`discover`, `library`, `apps`, etc.).
- `toren-store settings CATEGORY [ID]` directly opens the relevant section in Toren Settings Hub (`toren-shell hub open`).

## Test Suite

```sh
# Data plane verification
cd desktop && gofmt -l cmd/toren-store && go vet ./cmd/toren-store && go test ./cmd/toren-store

# UI logic verification
node --test shell/store/lib/store.test.mjs
```
