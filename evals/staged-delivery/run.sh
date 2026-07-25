#!/usr/bin/env bash
# Scaffold a fresh sandbox for one evaluation run of a staged-delivery variant.
#
# Usage:
#   ./run.sh <variant|baseline> [sandbox-parent-dir]
#
#   variant: tracer-bullet-delivery | hole-driven-delivery | test-ratchet-development | baseline
#
# Creates an empty git repo with ONLY the chosen skill installed (project-level
# .claude/skills), copies task.md in, and prints the kickoff prompt. Then you
# run `claude` inside the sandbox and act as the reviewer.
set -euo pipefail

VARIANT="${1:?usage: ./run.sh <variant|baseline> [sandbox-parent-dir]}"
PARENT="${2:-$HOME/staged-delivery-evals}"
REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
STAMP="$(date +%Y%m%d-%H%M)"
SANDBOX="$PARENT/$STAMP-$VARIANT"

if [[ "$VARIANT" != "baseline" && ! -f "$REPO_ROOT/$VARIANT/SKILL.md" ]]; then
  echo "error: unknown variant '$VARIANT' (no $VARIANT/SKILL.md in $REPO_ROOT)" >&2
  exit 1
fi

mkdir -p "$SANDBOX"
cd "$SANDBOX"
git init -q
cp "$REPO_ROOT/evals/staged-delivery/task.md" TASK.md

if [[ "$VARIANT" != "baseline" ]]; then
  mkdir -p .claude/skills
  cp -r "$REPO_ROOT/$VARIANT" ".claude/skills/$VARIANT"
fi

git add -A && git commit -qm "eval sandbox: $VARIANT"

cat <<EOF

Sandbox ready: $SANDBOX
Skill installed: $([[ "$VARIANT" == "baseline" ]] && echo "none (baseline)" || echo "$VARIANT (project-level, this sandbox only)")

Next:
  cd "$SANDBOX"
  claude

Kickoff prompt (paste verbatim):
  Complete the task described in TASK.md.

While it runs, act as the reviewer (see evals/staged-delivery/README.md for
the reviewer script and what to record). When done, copy
evals/staged-delivery/results/TEMPLATE.md into results/ and fill it in.
EOF
