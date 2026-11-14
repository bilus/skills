#!/bin/sh
# Counts changed lines for a step or a stage: insertions plus deletions,
# per file and in total, against the budget.
#
#   scripts/step-size.sh            working tree and index against HEAD,
#                                   plus untracked files
#   scripts/step-size.sh <commit>   that commit against its parent
#   scripts/step-size.sh <a>..<b>   the commits after <a> up to <b>, for
#                                   a stage
#
# Exit status is 0 when the total is within the budget, 1 when over,
# 2 on a usage error.
#
# Environment:
#   STEP_SIZE_LIMIT    budget, default 100 for a step and 399 for a stage
#   STEP_SIZE_EXCLUDE  extra exclusion patterns (extended regexes on the
#                      path, space separated), added to the defaults
#
# Excluded by default: lockfiles, vendored and generated code, the plan
# folder, diagram sources and renders, binary files. A file whose first
# three lines contain "DO NOT EDIT" is treated as generated.

set -eu

defaults='(^|/)(go\.sum|package-lock\.json|yarn\.lock|pnpm-lock\.yaml|Cargo\.lock|poetry\.lock|Gemfile\.lock|composer\.lock|flake\.lock|mix\.lock|Package\.resolved)$
\.lock$
(^|/)(vendor|node_modules|third_party)/
\.pb\.go$
_generated\.go$
_gen\.go$
\.gen\.[A-Za-z0-9]+$
\.generated\.[A-Za-z0-9]+$
^docs/plans/
\.dfd$
\.mmd$
\.dot$
\.svg$'

patterns="$defaults"
if [ -n "${STEP_SIZE_EXCLUDE:-}" ]; then
  for p in $STEP_SIZE_EXCLUDE; do
    patterns="$patterns
$p"
  done
fi

excluded() {
  path="$1"
  printf '%s\n' "$patterns" | while IFS= read -r p; do
    [ -z "$p" ] && continue
    if printf '%s\n' "$path" | grep -Eq -- "$p"; then
      echo yes
      return
    fi
  done
  if [ -f "$path" ] && head -n 3 -- "$path" 2>/dev/null | grep -q 'DO NOT EDIT'; then
    echo yes
  fi
}

unit=step
default=100
case "$#" in
  0)
    diff_cmd="git diff --numstat HEAD --"
    untracked=$(git ls-files --others --exclude-standard)
    ;;
  1)
    case "$1" in
      *..*)
        diff_cmd="git diff --numstat $1 --"
        unit=stage
        default=399
        ;;
      *)
        diff_cmd="git diff --numstat $1^ $1 --"
        ;;
    esac
    untracked=""
    ;;
  *)
    echo "usage: $0 [commit | from..to]" >&2
    exit 2
    ;;
esac
limit="${STEP_SIZE_LIMIT:-$default}"

tmp=$(mktemp)
trap 'rm -f "$tmp"' EXIT

# Tracked changes.
$diff_cmd | while IFS="$(printf '\t')" read -r ins del path; do
  [ -z "$path" ] && continue
  if [ "$ins" = "-" ] || [ "$del" = "-" ]; then
    printf '%8s  %s\n' 'binary' "$path"
    continue
  fi
  if [ -n "$(excluded "$path")" ]; then
    printf '%8s  %s\n' 'excluded' "$path"
    continue
  fi
  printf '%8d  %s\n' "$((ins + del))" "$path"
  printf '%s\n' "$((ins + del))" >> "$tmp"
done

# Untracked files count as insertions.
if [ -n "$untracked" ]; then
  printf '%s\n' "$untracked" | while IFS= read -r path; do
    [ -z "$path" ] && continue
    if [ -n "$(excluded "$path")" ]; then
      printf '%8s  %s (untracked)\n' 'excluded' "$path"
      continue
    fi
    if grep -Iq . -- "$path" 2>/dev/null || [ ! -s "$path" ]; then
      n=$(wc -l < "$path" | tr -d ' ')
      printf '%8d  %s (untracked)\n' "$n" "$path"
      printf '%s\n' "$n" >> "$tmp"
    else
      printf '%8s  %s (untracked)\n' 'binary' "$path"
    fi
  done
fi

total=0
if [ -s "$tmp" ]; then
  total=$(awk '{ s += $1 } END { print s + 0 }' "$tmp")
fi

echo "   total  $total of $limit"
if [ "$total" -gt "$limit" ]; then
  echo "OVER by $((total - limit)): split the $unit"
  exit 1
fi
echo "OK"
