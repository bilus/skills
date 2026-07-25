# Staged-delivery skill evaluation

How to compare `tracer-bullet-delivery`, `hole-driven-delivery`, and `test-ratchet-development` (plus a no-skill baseline) on the same task, and how to rerun the whole thing later.

All three skills share an identical outer loop (plan approval, ledger, LOC budget, halt-at-boundary) and differ only in the inner growth strategy, so differences you observe are attributable to the strategy.

## What's here

```
evals/staged-delivery/
├── README.md            # this file
├── task.md              # the standard test task (linkhub) — given to the agent verbatim
├── run.sh               # scaffolds a fresh sandbox per run
└── results/
    ├── TEMPLATE.md      # copy per run, fill in
    └── <date>-<variant>.md
```

## Rerunning the evaluation

One run = one variant (or baseline) on a fresh sandbox. Runs are interactive by design: the skills halt for human review, and reviewer behavior is part of what's being measured. Budget 30–60 min per run.

```bash
./run.sh tracer-bullet-delivery    # or hole-driven-delivery / test-ratchet-development / baseline
cd <printed sandbox dir>
claude
# paste: Complete the task described in TASK.md.
```

The sandbox is an empty git repo with only the variant under test installed at the project level (`.claude/skills/`), so your globally installed skills' triggers don't contaminate the run. For the baseline, no skill is installed.

### Reviewer script

Behave the same way in every run, or the comparison is meaningless:

1. **Plan review:** approve unless a stage is layer-shaped or an acceptance criterion is missing. One round of feedback max; note what you had to correct.
2. **Stage reviews:** read the handoff and the diff. Reply "approved" unless you find a real defect; if you do, state it in one sentence and let the agent fix it. Note review minutes and defects per stage.
3. **Sign-off requests** (test edits, signature changes): judge on the merits, note each occurrence.
4. Never volunteer process advice ("you should write a test first") — the skill is supposed to supply the process. If the agent asks a process question, answer "follow your instructions."

### After the run

1. Verify the result yourself in the sandbox — this is the ground truth, independent of the agent's claims:
   - run the test suite;
   - exercise requirements 1–6 from `task.md` by hand (curl), including the `javascript:` scheme rejection, the 11th-request 429, and a restart-persistence check;
   - `git log --oneline` and, per commit, `git show --stat` for stage sizes; check the suite is green at each stage boundary (`git checkout <sha> && <test cmd>`);
   - read the final `git log` as if you'd inherited the repo: commits must look human-maintained — imperative subjects, brief bodies describing the functionality/change/bugfix, no methodology vocabulary ("hole", "fill", "stage", "ledger", "census", spec-section refs), no bullet-list bodies, and (where the skill prescribes an endgame squash) process-cadence commits collapsed into stage-sized ones.
2. Copy `results/TEMPLATE.md` to `results/<YYYY-MM-DD>-<variant>.md` and fill it in.
3. Commit the results file back to this repo. The sandbox itself is disposable; archive it only if something interesting happened.

### What a fair comparison needs

- Same model, same day if possible; note both.
- Same reviewer following the script above.
- At least 2 runs per variant before drawing conclusions — single runs have high variance.
- The baseline run matters: it's what tells you whether a skill helps at all, not just which skill wins.

## What to conclude

Per variant, the questions that matter, in order: Did the final system actually satisfy all 6 requirements (verified by hand)? How many defects escaped to the final inspection vs. were caught at stage reviews? Was every stage within budget and green at its boundary? How much reviewer time did it cost? How often did the agent break its own discipline (each skill's README lists its variant-specific breaks)? Cheap-but-broken and correct-but-unreviewable are both failures; the target is correct, reviewable, and honest about its own process.
