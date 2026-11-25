# Packaging & XBPS Repository Architecture

Toren Linux uses the native **XBPS (X Binary Package System)** package manager from Void Linux.
System components, desktop shells, and tools are built and distributed as XBPS binary packages.

The packaging infrastructure resides in `packaging/xbps/`.

---

## Package Templates

Package build instructions are stored as templates in `packaging/xbps/`:

| Package Template | Description |
|---|---|
| `asura-desktop.template` | Core desktop control binary (`toren`), configuration schemas, session launchers, and runit service scripts. |
| `toren-shell.template` | Quickshell UI surfaces, theme tokens, widgets, icons, wallpapers, and Hub backend. |
| `toren-installer.template` | Guided system installer script, desktop launcher, and partition utilities. |
| `toren-store.template` | Software store CLI data plane (`toren-store`) and catalog manifests. |

---

## Building Binary Packages & Repository (`build-repo.sh`)

Binary packages are created using `xbps-create` and indexed with `xbps-rindex`.
The build workflow is driven by `packaging/xbps/build-repo.sh`.

### Package Build Pipeline

1. **Compilation**: Binaries are compiled with Go (`go build`) and stripped of debug symbols.
2. **Directory Staging**: Files are placed into target staging roots following Void filesystem hierarchy (`/usr/bin/`, `/usr/share/asura/`, `/etc/sv/`).
3. **Metadata Specification**:
   - `architecture`: `x86_64`
   - `maintainer`: `Toren Linux Team <team@asura-os.org>`
   - `license`: `GPL-3.0-only`
   - `pkgname` and `version`
4. **Package Creation**:
   ```sh
   xbps-create -A x86_64 -m "Toren Linux Team" -l "GPL-3.0-only" -s "Short summary" -V "0.1.0_1" destdir
   ```
5. **Repository Indexing**:
   ```sh
   xbps-rindex -a *.xbps
   xbps-rindex --sign --signedby "Toren Linux Signing Key" *.xbps
   ```

---

## Target System Configuration

The published repository, how it is signed, and how machines are pointed at it
are in [`package-repo.md`](package-repo.md). The package is one, `toren`:

```sh
sudo xbps-install -Su          # once the repository is configured
```
