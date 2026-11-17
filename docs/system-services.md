# System services

The system-level services every ASURA install ships with and starts at boot. Packages are
listed in [`system/packages-void.yaml`](../system/packages-void.yaml). Services are linked
into `/etc/runit/runsvdir/default` by the same loop in
[`installation/build-iso-void.sh`](../installation/build-iso-void.sh) (live ISO) and
[`installation/asura-installer.sh`](../installation/asura-installer.sh) (installed system).
`installation/build-iso-void_test.sh` fails if either list drops one of them.

| Need | Package | Runit service | Notes |
|---|---|---|---|
| Clock sync | `chrony` | `chronyd` | Without it the clock drifts until TLS, `asura update` and package signature checks fail |
| Inbound firewall | `nftables` | `nftables` | Loads `/etc/nftables.conf`, ASURA's ruleset below |
| Removable media | `udisks2` | none (D-Bus activated) | Lets Thunar and gvfs mount USB drives and SD cards without root |
| Firmware updates | `fwupd` | none (D-Bus activated) | `fwupdmgr refresh && fwupdmgr update` updates BIOS, SSD and dock firmware from LVFS |
| Printing | `cups`, `cups-filters` | `cupsd` | Add printers at <http://localhost:631> |
| Device power saving | `tlp` | `tlp` | USB autosuspend, Wi-Fi power save, SATA and PCIe runtime PM. CPU governor, EPP and platform profile stay with `asura-sysd`; [`system/tlp.d/01-asura-cpu.conf`](../system/tlp.d/01-asura-cpu.conf) disables them in tlp |

## Firewall

The ruleset lives in [`system/nftables/nftables.conf`](../system/nftables/nftables.conf) and is
installed as `/etc/nftables.conf`.

- **Inbound:** dropped unless it belongs to a connection the machine started. Loopback,
  ICMP/ICMPv6 (needed for IPv6 and path MTU discovery) and DHCPv6 replies are accepted.
- **Outbound:** everything is allowed.
- **Forwarding:** not touched. Docker and Podman manage their own NAT and forwarding rules.

The file replaces only its own table (`inet asura_filter`) and never runs `flush ruleset`.
To apply a change, reload the file with `sudo nft -f /etc/nftables.conf`; that leaves
Docker's and Podman's rules alone. **Don't** use `sv restart nftables` for this: the Void
service's `finish` script runs `nft flush ruleset` when it stops, which also removes the
container tools' rules until they restart.

`openssh` is installed but `sshd` is not enabled, and the firewall doesn't accept port 22. To
accept a service from the network, enable it and add an `accept` rule for its port to the input
chain.

The `nftables` package ships no `/etc/nftables.conf` of its own, so package updates never
replace ASURA's file.

## Secure Boot

**Not supported yet.** The ISO and installed systems boot through an unsigned GRUB, and no
shim or Machine Owner Key enrolment is set up. On a machine with Secure Boot enabled, the
firmware refuses to boot the ISO.

Until it is implemented, disable Secure Boot in the firmware settings before booting the ISO.
Implementing it means either shipping a signed shim with MOK enrolment in the installer, or
signing GRUB and the kernel with machine-local keys (`sbctl`); the decision is open in
[`roadmap.md`](roadmap.md).
