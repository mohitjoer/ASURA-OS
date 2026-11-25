#!/bin/sh
# Headless QEMU boot smoke test for a built ISO: boot it, watch the serial
# console for boot progress, take a screenshot, and stop the VM.
#
# Usage: installation/ci/qemu-boot-test.sh ISO OUTDIR
#
# Writes OUTDIR/serial.log and, when QMP screendump works, OUTDIR/boot-screen.png.
#
# Without /dev/kvm the boot runs in software emulation and can be too slow to
# reach the markers in time, so a missed boot is only a notice. Set
# ASURA_BOOT_STRICT=1 (release.yml does) to fail in that case too.
set -eu

iso=${1:?usage: qemu-boot-test.sh ISO OUTDIR}
OUT=${2:?usage: qemu-boot-test.sh ISO OUTDIR}
STRICT=${ASURA_BOOT_STRICT:-0}

mkdir -p "$OUT"
SERIAL_LOG="$OUT/serial.log"
QMP_SOCK="$OUT/qmp.sock"
SCREENSHOT="$OUT/boot-screen.ppm"
SCREENSHOT_PNG="$OUT/boot-screen.png"
PIDFILE="$OUT/qemu.pid"

stop_qemu() {
	if [ -f "$PIDFILE" ]; then
		kill "$(cat "$PIDFILE")" 2>/dev/null || true
		rm -f "$PIDFILE"
	fi
}
trap stop_qemu EXIT

if [ -w /dev/kvm ]; then
	echo "Hardware virtualization (/dev/kvm) available; using KVM acceleration"
	ACCEL="-enable-kvm -cpu host"
else
	echo "No /dev/kvm available; running in TCG software emulation mode"
	ACCEL="-cpu qemu64"
fi

echo "Starting QEMU boot test ($ACCEL)..."
# shellcheck disable=SC2086
qemu-system-x86_64 \
	$ACCEL \
	-m 2048 \
	-smp 2 \
	-cdrom "$iso" \
	-boot d \
	-vga std \
	-display vnc=127.0.0.1:0 \
	-serial file:"$SERIAL_LOG" \
	-qmp unix:"$QMP_SOCK",server,nowait \
	-daemonize \
	-pidfile "$PIDFILE"

echo "Waiting for boot sequence (~60-90s in software emulation)..."
BOOT_DETECTED=0
for i in $(seq 1 45); do
	sleep 2
	if [ -f "$SERIAL_LOG" ] && grep -qaE "runit: enter stage: /etc/runit/2|toren-linux login:|toren-os login:|turnstiled: Initializing|Linux version|dracut|ISOLINUX" "$SERIAL_LOG"; then
		echo "Detected boot progress in serial log at ~$((i * 2))s!"
		BOOT_DETECTED=1
		break
	fi
done

# Capture the virtual screen through QMP screendump.
python3 - "$QMP_SOCK" "$SCREENSHOT" <<'PY' || true
import json, socket, sys, time
sock_path, ppm_path = sys.argv[1], sys.argv[2]
time.sleep(1)
try:
    s = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    s.connect(sock_path)
    s.recv(1024)
    s.sendall(b'{"execute": "qmp_capabilities"}\n')
    s.recv(1024)
    cmd = {"execute": "screendump", "arguments": {"filename": ppm_path}}
    s.sendall((json.dumps(cmd) + "\n").encode())
    s.recv(1024)
    s.close()
    print("Captured screenshot to " + ppm_path)
except Exception as e:
    print("Warning: screenshot capture failed: " + str(e))
PY

# screendump returns before the file is fully written; give it a moment.
sleep 1
if [ -f "$SCREENSHOT" ] && command -v pnmtopng >/dev/null 2>&1; then
	pnmtopng "$SCREENSHOT" >"$SCREENSHOT_PNG" 2>/dev/null || true
fi

stop_qemu

echo "== Serial Console Output =="
head -n 40 "$SERIAL_LOG" 2>/dev/null || true
echo "..."
tail -n 40 "$SERIAL_LOG" 2>/dev/null || true
echo "==========================="

if [ "$BOOT_DETECTED" -eq 0 ]; then
	if [ -w /dev/kvm ] || [ "$STRICT" = 1 ]; then
		echo "::error::QEMU boot test did not detect a boot sequence within the timeout" >&2
		exit 1
	fi
	echo "Notice: no boot sequence detected in software emulation without /dev/kvm; not failing (ASURA_BOOT_STRICT=0)."
fi
