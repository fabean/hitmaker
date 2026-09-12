#!/bin/sh
# Run from the repository root. Recording never reads or saves a personal song.
set -eu
unset NO_COLOR
export TERM=xterm-256color COLORTERM=truecolor
demo_dir=$(mktemp -d "${TMPDIR:-/tmp}/hitmaker-demo.XXXXXX")
trap 'rm -rf "$demo_dir"' EXIT HUP INT TERM
./hitmaker --demo --data-dir "$demo_dir/data" --config-dir "$demo_dir/config"
