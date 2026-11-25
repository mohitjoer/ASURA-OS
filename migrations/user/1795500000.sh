#!/bin/sh
echo "Migrate user config directory for Toren Linux transition"

CFG="${XDG_CONFIG_HOME:-$HOME/.config}"

if [ -d "$CFG/toren" ] && [ ! -e "$CFG/toren" ]; then
    ln -s toren "$CFG/toren"
elif [ -d "$CFG/toren" ] && [ ! -e "$CFG/toren" ]; then
    ln -s toren "$CFG/toren"
fi

exit 0
