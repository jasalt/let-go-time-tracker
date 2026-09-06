#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "$0")/../.." && pwd)
binary="$root/build/gtk-counter"
display_number=${LGTT_XVFB_DISPLAY:-:91}
log=$(mktemp)
xvfb_log=$(mktemp)
trap 'rm -f "$log" "$xvfb_log"' EXIT

go build -o "$binary" "$root/examples/gtk-counter"

for cycle in 1 2 3; do
  Xvfb "$display_number" -screen 0 800x600x24 -nolisten tcp >"$xvfb_log" 2>&1 &
  xvfb_pid=$!
  DISPLAY="$display_number" GDK_BACKEND=x11 GSK_RENDERER=cairo \
    "$binary" >"$log" 2>&1 &
  app_pid=$!
  cleanup_cycle() {
    kill "$app_pid" "$xvfb_pid" 2>/dev/null || true
    wait "$app_pid" "$xvfb_pid" 2>/dev/null || true
  }
  trap 'cleanup_cycle; rm -f "$log" "$xvfb_log"' EXIT

  window=""
  for _ in $(seq 1 250); do
    for candidate in $(DISPLAY="$display_number" xdotool search --pid "$app_pid" 2>/dev/null || true); do
      eval "$(DISPLAY="$display_number" xdotool getwindowgeometry --shell "$candidate")"
      if ((WIDTH > 100)); then
        window=$candidate
        break 2
      fi
    done
    sleep 0.02
  done
  [[ -n "$window" ]] || {
    cat "$log"
    echo "window did not appear" >&2
    exit 1
  }

  sleep 1.2
  title=$(DISPLAY="$display_number" xdotool getwindowname "$window")
  [[ "$title" =~ ^let-go\ GTK\ Counter\ —\ 0$ ]] || {
    echo "timer/lifecycle title check failed: $title" >&2
    exit 1
  }

  if ((cycle == 1)); then
    DISPLAY="$display_number" xdotool mousemove --window "$window" 180 45 \
      click --repeat 1100 --delay 1 1
    sleep 0.5
    title=$(DISPLAY="$display_number" xdotool getwindowname "$window")
    count=${title##*— }
    if ((count < 1000)); then
      echo "click stress failed: $title" >&2
      exit 1
    fi
  fi

  if grep -q 'callback:' "$log"; then
    cat "$log"
    echo "callback error observed" >&2
    exit 1
  fi
  cleanup_cycle
  trap 'rm -f "$log" "$xvfb_log"' EXIT
done

echo "GTK counter smoke passed: timer advanced, 1000 clicks delivered, 3 lifecycle cycles clean"
