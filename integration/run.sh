#!/bin/sh
# Integration checks for yup-uniq, run inside a Debian (GNU coreutils) container.
#
# parity ARGS...  — yup-uniq reading stdin must be byte-identical to GNU `uniq`.
# parity_file ARGS... — yup-uniq reading a file operand must match GNU `uniq`.
#
# uniq collapses only ADJACENT duplicate lines, so the sample is pre-grouped.
set -eu

fails=0
sample='apple
apple
Apple
banana
banana
banana
cherry'

parity() {
	ours=$(printf '%s\n' "$sample" | yup-uniq "$@" 2>/dev/null || true)
	gnu=$(printf '%s\n' "$sample" | uniq "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  uniq %s < stdin\n' "$*"
	else
		printf 'FAIL  parity  uniq %s < stdin\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

parity_file() {
	printf '%s\n' "$sample" > /tmp/in.txt
	ours=$(yup-uniq "$@" /tmp/in.txt 2>/dev/null || true)
	gnu=$(uniq "$@" /tmp/in.txt 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  uniq %s /tmp/in.txt\n' "$*"
	else
		printf 'FAIL  parity  uniq %s /tmp/in.txt\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

# Default: collapse each run of adjacent duplicates to one line.
parity
# -c / --count: prefix each line with its width-7 occurrence count (GNU's %7d).
parity -c
# -d / --repeated: only groups that repeat (count > 1).
parity -d
# -u / --unique: only groups that do not repeat (count == 1).
parity -u
# -i / --ignore-case: fold case when comparing adjacent lines.
parity -i
# -c with -i: count groups formed under case-insensitive comparison.
parity -c -i
# File operand instead of stdin.
parity_file
parity_file -c

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
