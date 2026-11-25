#!/usr/bin/env python3
"""Check system/packages-void.yaml against the Void Linux repositories.

A wrong package name must fail loudly when the image is planned, not vanish
quietly when it is built. This resolves every name in the manifest against
the XBPS repositories this machine has configured (the Void mirrors and the
hyprland-void repository the image adds), with `xbps-query -R`, and checks
two ways:

  packages   every `packages:` entry must exist in a repository. A missing
             name is an error.
  exclude    every `exclude:` entry (`- name: reason`) names a package the
             image must not carry. One that no repository has protects
             nothing and is reported, because an exclusion nobody can see the
             effect of is how two implementations of one thing get installed.

  python3 system/verify-packages.py [--manifest PATH] [--strict]

Exit 0 when every package resolves, 1 when one does not (or, with --strict,
when an exclusion protects nothing), 2 when xbps-query is not available: the
check needs a Void system (or a container with xbps and the repositories).
"""
import argparse
import shutil
import subprocess
import sys

DEFAULT_MANIFEST = "system/packages-void.yaml"


def load(path):
    """packages and exclusions per group, without needing PyYAML.

    The manifest's shape is fixed: top-level groups, each with `packages:`
    and optionally `exclude:` lists of `- name` or `- name: reason`.
    """
    groups, group, key = {}, None, None
    with open(path, encoding="utf-8") as f:
        for raw in f:
            line = raw.split(" #", 1)[0].rstrip() if not raw.lstrip().startswith("#") else ""
            if not line.strip():
                continue
            indent = len(line) - len(line.lstrip())
            text = line.strip()
            if indent == 0 and text.endswith(":"):
                group, key = text[:-1], None
                groups.setdefault(group, {"packages": [], "exclude": []})
            elif indent == 2 and text in ("packages:", "exclude:") and group:
                key = text[:-1]
            elif indent == 2:
                key = None
            elif text.startswith("- ") and group and key:
                name = text[2:].split(":", 1)[0].strip()
                if name:
                    groups[group][key].append(name)
    return groups


def in_repository(name):
    """Whether a configured XBPS repository has the package."""
    result = subprocess.run(
        ["xbps-query", "-R", "--property", "pkgver", name],
        capture_output=True, text=True, check=False,
    )
    return result.returncode == 0 and result.stdout.strip() != ""


def main(argv):
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--manifest", default=DEFAULT_MANIFEST)
    parser.add_argument("--strict", action="store_true",
                        help="also fail when an exclusion protects nothing")
    args = parser.parse_args(argv)

    groups = load(args.manifest)
    if not any(g["packages"] for g in groups.values()):
        print(f"verify-packages: no packages found in {args.manifest}", file=sys.stderr)
        return 1
    if shutil.which("xbps-query") is None:
        print("verify-packages: xbps-query is not installed; run this on Void Linux "
              "(or in a container with xbps and the image's repositories)", file=sys.stderr)
        return 2

    missing, unprotected, total = [], [], 0
    for group, entries in groups.items():
        for name in entries["packages"]:
            total += 1
            if not in_repository(name):
                missing.append((group, name))
        for name in entries["exclude"]:
            if not in_repository(name):
                unprotected.append((group, name))

    for group, name in missing:
        print(f"MISSING   {group}: {name} is in no configured repository")
    for group, name in unprotected:
        print(f"NOTHING   {group}: exclusion {name} matches no package")
    print(f"verify-packages: {total - len(missing)}/{total} packages resolve, "
          f"{len(unprotected)} exclusion(s) protect nothing")
    if missing or (args.strict and unprotected):
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
