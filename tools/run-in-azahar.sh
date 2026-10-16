#!/bin/bash
# Runs a .3dsx in Azahar under Xvfb and prints what the program wrote to the debug log
# (svcOutputDebugString). It stops the emulator as soon as the program prints a final line
# (PASS, FAIL or a line starting with DONE), or after the given seconds. It also warns about
# unmapped memory accesses, which stop the program on the console but not in the emulator.
# Usage: tools/run-in-azahar.sh program.3dsx [seconds]
program=$(realpath "$1"); seconds=${2:-20}
# AZAHAR is the emulator's program (an extracted AppImage's usr/bin/azahar); the debug log must be
# on (Debug.Emulated:Debug and instant_debug_log=true in qt-config.ini).
azahar=${AZAHAR:-azahar}
pgrep -f "Xvfb :91" >/dev/null || (setsid Xvfb :91 -screen 0 1280x960x24 </dev/null >/dev/null 2>&1 &)
sleep 1
log=~/.local/share/azahar-emu/log/azahar_log.txt
rm -f "$log"
DISPLAY=:91 LIBGL_ALWAYS_SOFTWARE=1 setsid "$azahar" "$program" </dev/null >/dev/null 2>&1 &
emulator=$!
output() { grep -a "OutputDebugString" "$log" 2>/dev/null | sed -E 's/.*OutputDebugString:[0-9]+: ?//'; }
for ((second = 0; second < seconds; second++)); do
	sleep 1
	if output | grep -qE '^(PASS|FAIL)$|^DONE'; then
		break
	fi
done
kill -9 $emulator 2>/dev/null
wait $emulator 2>/dev/null
output
# The console stops a program that touches unmapped memory; the emulator only logs it.
unmapped=$(grep -ac "UnmappedAccess" "$log")
if [ "$unmapped" -gt 0 ]; then
	echo "WARNING: $unmapped accesses to unmapped memory; the first:"
	grep -a -m1 "UnmappedAccess" "$log" | sed -E 's/.*UnmappedAccess:[0-9]+: //'
fi
