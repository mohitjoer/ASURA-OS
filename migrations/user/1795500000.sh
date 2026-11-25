#!/bin/sh
echo "Migrate user config directory for Toren Linux transition"

CFG="${XDG_CONFIG_HOME:-$HOME/.config}"

if [ -d "$CFG/asura" ] && [ ! -e "$CFG/toren" ]; then
    ln -s asura "$CFG/toren"
elif [ -d "$CFG/toren" ] && [ ! -e "$CFG/asura" ]; then
    ln -s toren "$CFG/asura"
fi

exit 0
