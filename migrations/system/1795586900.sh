#!/bin/sh
echo "Enable beta system services (socklog, earlyoom, maintenance, thermald)"

enable_svc() {
	svc="$1"
	if [ -d "/etc/sv/$svc" ]; then
		mkdir -p /var/service /etc/runit/runsvdir/default
		if [ ! -L "/var/service/$svc" ]; then
			ln -sf "/etc/sv/$svc" "/var/service/$svc" 2>/dev/null || true
		fi
		if [ ! -L "/etc/runit/runsvdir/default/$svc" ]; then
			ln -sf "/etc/sv/$svc" "/etc/runit/runsvdir/default/$svc" 2>/dev/null || true
		fi
	fi
}

enable_svc socklog-unix
enable_svc nanoklogd
enable_svc earlyoom
enable_svc toren-maintenance

# thermald only on Intel CPUs
if grep -qi "GenuineIntel\|Intel" /proc/cpuinfo 2>/dev/null; then
	enable_svc thermald
fi

exit 0
