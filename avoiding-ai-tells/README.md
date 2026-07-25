# Avoiding AI Tells — README

Style rules for AI-generated code that humans will review. Reviewers pattern-match on machine-generation signals (uniform comments, gold-plating, template READMEs, mega-commits, em dashes), and once they spot one they review hostilely or dismiss the work as slop regardless of quality. This skill strips those signals so the work gets judged on its merits. It is explicitly not about hiding authorship: it forbids faking struggle or history, and tells the agent to disclose substantial AI assistance in one plain sentence.

## What's in it

Rules grouped by where tells show up: comments (the loudest tell), typography, scope and completeness, naming and structure, error handling, tests, docs and READMEs, commits and PRs, plus what never to fake. The meta-rule: uniformity is the tell; vary effort with difficulty and, in doubt, do less.

## Usage

Save the skill and leave it enabled — it applies to any code, docs, or commits an agent produces for human review. The three staged-delivery skills in this collection (`tracer-bullet-delivery`, `hole-driven-delivery`, `test-ratchet-development`) reference it as required background: staged delivery makes diffs small enough to review, this makes them read as work rather than output.
