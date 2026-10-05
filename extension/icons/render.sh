#!/usr/bin/env bash
# Renders the icons' PNGs from icon.html (in headless Edge, for its
# transparent screenshots): ./render.sh, or ./render.sh 汉 for another
# character. Needs python3 with PIL.
set -euo pipefail
cd "$(dirname "$0")"
c=${1:-中}
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
microsoft-edge --headless --disable-gpu --no-first-run --user-data-dir="$tmp/profile" \
  --allow-file-access-from-files --default-background-color=00000000 --hide-scrollbars \
  --window-size=512,512 --virtual-time-budget=5000 \
  --screenshot="$tmp/big.png" "file://$PWD/icon.html?c=$c" >/dev/null 2>&1
python3 - "$tmp" <<'PY'
import sys
from PIL import Image
big = Image.open(sys.argv[1] + "/big.png").convert("RGBA")
for n in (16, 32, 48, 128):
    big.resize((n, n), Image.LANCZOS).save(f"icon{n}.png")
PY
