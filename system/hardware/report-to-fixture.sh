#!/bin/sh
# report-to-fixture.sh: converts a redacted hardware report into a test fixture
# and outputs the matched layers.
# Usage: ./report-to-fixture.sh <report-file> [fixture-name]

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <report-file> [fixture-name]" >&2
    exit 1
fi

REPORT_FILE="$1"
NAME="$2"

if [ ! -f "$REPORT_FILE" ]; then
    echo "Error: report file '$REPORT_FILE' not found" >&2
    exit 1
fi

# Determine base dir
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Run python helper to parse report and create fixture
python3 - << EOF
import sys, os, json, re

report_path = "$REPORT_FILE"
name = "$NAME"
repo_root = "$REPO_ROOT"

with open(report_path, "r") as f:
    content = f.read()

# Simple DMI parser from report output
dmi = {}
for line in content.splitlines():
    line = line.strip()
    if ":" in line:
        k, v = line.split(":", 1)
        k = k.strip()
        v = v.strip().strip('"').strip("'")
        if k in ["sys_vendor", "product_name", "product_family", "board_name", "chassis_type", "bios_vendor", "bios_version", "bios_date"]:
            dmi[k] = v

if not name:
    vendor = dmi.get("sys_vendor", "unknown").lower().replace(" ", "-")
    model = dmi.get("product_name", dmi.get("product_family", "machine")).lower().replace(" ", "-")
    name = re.sub(r'[^a-z0-9_-]', '', f"{vendor}-{model}")

fixture_dir = os.path.join(repo_root, "system", "hardware", "testdata", "machines", name)
dmi_dir = os.path.join(fixture_dir, "class", "dmi", "id")
os.makedirs(dmi_dir, exist_ok=True)

for k, v in dmi.items():
    with open(os.path.join(dmi_dir, k), "w") as f:
        f.write(v + "\n")

# Redacted serial numbers
for s in ["product_serial", "chassis_serial", "board_serial"]:
    with open(os.path.join(dmi_dir, s), "w") as f:
        f.write("[REDACTED]\n")

for b in ["bios_vendor", "bios_version", "bios_date"]:
    if b not in dmi:
        with open(os.path.join(dmi_dir, b), "w") as f:
            f.write("Vendor\n")

for empty in ["lspci.txt", "lsusb.txt", "modules.txt"]:
    path = os.path.join(fixture_dir, empty)
    if not os.path.exists(path):
        with open(path, "w") as f:
            f.write("\n")

print(f"Fixture created at: {fixture_dir}")
print(f"Fixture name: {name}")
EOF

echo "Running match check to compute matched layers..."
cd "$REPO_ROOT/system/hardware"
go test -count=1 ./quirks -run TestMachineFixtures -v || true
