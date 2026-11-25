#!/bin/sh
# build-iso.sh is a link to build-iso-void.sh (the Void Linux mklive build), so
# its test is build-iso-void_test.sh. This name is kept for anyone who runs it.
set -eu
REPO_ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
exec sh "$REPO_ROOT/installation/build-iso-void_test.sh" "$@"
