#!/usr/bin/env bash

QUICKSHELL_CONFIG_NAME="ii"
XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
XDG_CACHE_HOME="${XDG_CACHE_HOME:-$HOME/.cache}"
XDG_STATE_HOME="${XDG_STATE_HOME:-$HOME/.local/state}"
CONFIG_DIR="$XDG_CONFIG_HOME/quickshell/$QUICKSHELL_CONFIG_NAME"
CACHE_DIR="$XDG_CACHE_HOME/quickshell"
STATE_DIR="$XDG_STATE_HOME/quickshell"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

term_alpha=100 # Set this to < 100 make all your terminals transparent
if [ ! -d "$STATE_DIR"/user/generated ]; then
  mkdir -p "$STATE_DIR"/user/generated
fi
cd "$CONFIG_DIR" 2>/dev/null || true

# If material_colors.scss is empty or missing, derive it directly from colors.json using jq
if [ ! -s "$STATE_DIR/user/generated/material_colors.scss" ]; then
  json_src=""
  if [ -s "$STATE_DIR/user/generated/colors.json" ]; then
    json_src="$STATE_DIR/user/generated/colors.json"
  elif [ -s "$XDG_CACHE_HOME/asura/colors.json" ]; then
    json_src="$XDG_CACHE_HOME/asura/colors.json"
  fi
  if [ -n "$json_src" ] && command -v jq >/dev/null 2>&1; then
    jq -r 'to_entries | .[] | "$\(.key): \(.value);"' "$json_src" > "$STATE_DIR/user/generated/material_colors.scss" 2>/dev/null || true
  fi
fi

colornames=''
colorstrings=''
colorlist=()
colorvalues=()

if [ -s "$STATE_DIR/user/generated/material_colors.scss" ]; then
  colornames=$(cat "$STATE_DIR/user/generated/material_colors.scss" | cut -d: -f1)
  colorstrings=$(cat "$STATE_DIR/user/generated/material_colors.scss" | cut -d: -f2 | cut -d ' ' -f2 | cut -d ";" -f1)
  IFS=$'\n'
  colorlist=($colornames)     # Array of color names
  colorvalues=($colorstrings) # Array of color values
fi

apply_ghostty() {
  local json_file=""
  if [ -s "$STATE_DIR/user/generated/colors.json" ]; then
    json_file="$STATE_DIR/user/generated/colors.json"
  elif [ -s "$XDG_CACHE_HOME/asura/colors.json" ]; then
    json_file="$XDG_CACHE_HOME/asura/colors.json"
  fi

  [ -z "$json_file" ] && return 0
  command -v jq >/dev/null 2>&1 || return 0

  # Sync to ~/.cache/asura/colors.json with Material roles + Scheme.qml aliases + camelCase aliases
  mkdir -p "$XDG_CACHE_HOME/asura"
  jq '. + {
    "color1": (.error // "#f7768e"),
    "color2": (.primary // "#9ece6a"),
    "color3": (.tertiary // "#e0af68"),
    "color4": (.secondary // "#7aa2f7"),
    "color5": (.primary_fixed // "#bb9af7"),
    "color6": (.secondary_fixed // "#7dcfff"),
    "onSurface": (.on_surface // .onSurface),
    "onSurfaceVariant": (.on_surface_variant // .onSurfaceVariant),
    "surfaceContainerLow": (.surface_container_low // .surfaceContainerLow),
    "surfaceContainer": (.surface_container // .surfaceContainer),
    "surfaceContainerHigh": (.surface_container_high // .surfaceContainerHigh),
    "surfaceContainerHighest": (.surface_container_highest // .surfaceContainerHighest),
    "surfaceContainerLowest": (.surface_container_lowest // .surfaceContainerLowest),
    "surfaceVariant": (.surface_variant // .surfaceVariant),
    "surfaceTint": (.surface_tint // .surfaceTint),
    "inverseSurface": (.inverse_surface // .inverseSurface),
    "inverseOnSurface": (.inverse_on_surface // .inverseOnSurface),
    "primaryContainer": (.primary_container // .primaryContainer),
    "secondaryContainer": (.secondary_container // .secondaryContainer),
    "tertiaryContainer": (.tertiary_container // .tertiaryContainer),
    "errorContainer": (.error_container // .errorContainer),
    "outlineVariant": (.outline_variant // .outlineVariant)
  }' "$json_file" > "$XDG_CACHE_HOME/asura/colors.json.tmp" 2>/dev/null && mv "$XDG_CACHE_HOME/asura/colors.json.tmp" "$XDG_CACHE_HOME/asura/colors.json"

  local bg fg cursor cur_text sel_bg sel_fg p0 p1 p2 p3 p4 p5 p6 p7 p8 p9 p10 p11 p12 p13 p14 p15
  bg=$(jq -r '.background // "#0f0c07"' "$json_file")
  fg=$(jq -r '.on_surface // "#f3ede1"' "$json_file")
  cursor=$(jq -r '.primary // "#e2342a"' "$json_file")
  cur_text=$(jq -r '.on_primary // "#ffffff"' "$json_file")
  sel_bg=$(jq -r '.secondary_container // "#3b4460"' "$json_file")
  sel_fg=$(jq -r '.on_secondary_container // "#ffffff"' "$json_file")

  p0=$(jq -r '.surface_container_high // "#282a2f"' "$json_file")
  p1=$(jq -r '.error // "#ffb4ab"' "$json_file")
  p2=$(jq -r '.primary // "#b8c4ff"' "$json_file")
  p3=$(jq -r '.tertiary // "#e0af68"' "$json_file")
  p4=$(jq -r '.secondary // "#c3c5dd"' "$json_file")
  p5=$(jq -r '.primary_fixed // "#dfe0ff"' "$json_file")
  p6=$(jq -r '.secondary_fixed // "#e2e2f9"' "$json_file")
  p7=$(jq -r '.on_surface_variant // "#c6c6cf"' "$json_file")
  p8=$(jq -r '.outline // "#90909a"' "$json_file")
  p9=$(jq -r '.error // "#ffb4ab"' "$json_file")
  p10=$(jq -r '.primary_fixed // "#dfe0ff"' "$json_file")
  p11=$(jq -r '.tertiary_fixed // "#ffdcc1"' "$json_file")
  p12=$(jq -r '.secondary_fixed // "#e2e2f9"' "$json_file")
  p13=$(jq -r '.primary_fixed_dim // "#b8c4ff"' "$json_file")
  p14=$(jq -r '.outline // "#90909a"' "$json_file")
  p15=$(jq -r '.on_surface // "#e2e2e9"' "$json_file")

  mkdir -p "$XDG_CONFIG_HOME/asura/apps/ghostty" "$XDG_CONFIG_HOME/ghostty"
  cat >"$XDG_CONFIG_HOME/asura/apps/ghostty/colors" <<EOF
# Asura palette -> ghostty. Rendered by asura theme; do not edit.
background = $bg
foreground = $fg
cursor-color = $cursor
cursor-text = $cur_text
selection-background = $sel_bg
selection-foreground = $sel_fg
palette = 0=$p0
palette = 1=$p1
palette = 2=$p2
palette = 3=$p3
palette = 4=$p4
palette = 5=$p5
palette = 6=$p6
palette = 7=$p7
palette = 8=$p8
palette = 9=$p9
palette = 10=$p10
palette = 11=$p11
palette = 12=$p12
palette = 13=$p13
palette = 14=$p14
palette = 15=$p15
EOF

  cp "$XDG_CONFIG_HOME/asura/apps/ghostty/colors" "$XDG_CONFIG_HOME/ghostty/colors" 2>/dev/null || true

  # Broadcast live OSC sequences to active PTYs so open terminals retint immediately
  local osc_seq=""
  osc_seq+="\033]10;${fg}\007"
  osc_seq+="\033]11;${bg}\007"
  osc_seq+="\033]12;${cursor}\007"
  osc_seq+="\033]4;0;${p0}\007\033]4;1;${p1}\007\033]4;2;${p2}\007\033]4;3;${p3}\007"
  osc_seq+="\033]4;4;${p4}\007\033]4;5;${p5}\007\033]4;6;${p6}\007\033]4;7;${p7}\007"
  osc_seq+="\033]4;8;${p8}\007\033]4;9;${p9}\007\033]4;10;${p10}\007\033]4;11;${p11}\007"
  osc_seq+="\033]4;12;${p12}\007\033]4;13;${p13}\007\033]4;14;${p14}\007\033]4;15;${p15}\007"

  for file in /dev/pts/*; do
    if [[ $file =~ ^/dev/pts/[0-9]+$ ]]; then
      printf "%b" "$osc_seq" >"$file" 2>/dev/null & disown || true
    fi
  done

  # Signal ghostty instances to reload configuration file
  pkill -SIGUSR1 ghostty 2>/dev/null || true
  pkill -SIGUSR2 ghostty 2>/dev/null || true
}

apply_kitty() {  
  # Check if terminal escape sequence template exists
  if [ ! -f "$SCRIPT_DIR/terminal/kitty-theme.conf" ]; then
    return
  fi
  # Copy template
  mkdir -p "$STATE_DIR"/user/generated/terminal
  cp "$SCRIPT_DIR/terminal/kitty-theme.conf" "$STATE_DIR"/user/generated/terminal/kitty-theme.conf
  # Apply colors
  for i in "${!colorlist[@]}"; do
    sed -i "s/${colorlist[$i]} #/${colorvalues[$i]#\#}/g" "$STATE_DIR"/user/generated/terminal/kitty-theme.conf 2>/dev/null || true
  done

  # Reload
  if pgrep -f kitty >/dev/null 2>&1; then
    kill -SIGUSR1 $(pidof kitty) 2>/dev/null || true
  fi
}

apply_anyterm() {
  # Check if terminal escape sequence template exists
  if [ ! -f "$SCRIPT_DIR/terminal/sequences.txt" ]; then
    return
  fi
  # Copy template
  mkdir -p "$STATE_DIR"/user/generated/terminal
  cp "$SCRIPT_DIR/terminal/sequences.txt" "$STATE_DIR"/user/generated/terminal/sequences.txt
  # Apply colors
  for i in "${!colorlist[@]}"; do
    sed -i "s/${colorlist[$i]} #/${colorvalues[$i]#\#}/g" "$STATE_DIR"/user/generated/terminal/sequences.txt 2>/dev/null || true
  done

  sed -i "s/\$alpha/$term_alpha/g" "$STATE_DIR/user/generated/terminal/sequences.txt" 2>/dev/null || true

  for file in /dev/pts/*; do
    if [[ $file =~ ^/dev/pts/[0-9]+$ ]]; then
      {
      cat "$STATE_DIR"/user/generated/terminal/sequences.txt >"$file"
      } & disown || true
    fi
  done
}

apply_term() {
  apply_ghostty &
  apply_anyterm &
  apply_kitty &
}

# Check if terminal theming is enabled in config
CONFIG_FILE="$XDG_CONFIG_HOME/illogical-impulse/config.json"
if [ -f "$CONFIG_FILE" ]; then
  enable_terminal=$(jq -r '.appearance.wallpaperTheming.enableTerminal' "$CONFIG_FILE" 2>/dev/null || echo "true")
  if [ "$enable_terminal" = "true" ]; then
    apply_term &
  fi
else
  apply_term &
fi
