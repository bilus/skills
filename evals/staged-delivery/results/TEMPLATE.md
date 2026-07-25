# Run: <variant> — <YYYY-MM-DD>

- Model: <model id>
- Reviewer: <name>
- Sandbox: <path or "discarded">
- Wall clock: <start–end, total minutes>

## Outcome (verified by hand in the sandbox)

| Requirement | Works? | Notes |
|---|---|---|
| 1. POST /links creates, 201, unique 6-char slug | | |
| 2. 400 on invalid body; non-http(s) scheme rejected | | |
| 3. GET /<slug> 302 + visit count; 404 unknown | | |
| 4. GET /links sorted by visits desc | | |
| 5. Rate limit 10/min/IP, 11th → 429 | | |
| 6. Survives restart (SQLite) | | |
| Full test suite green at HEAD | | |

## Stages

| Stage | Claim (from handoff) | Net LOC | Green at boundary? | Review min | Defects found in review |
|---|---|---|---|---|---|
| 1 | | | | | |

- Stages over the 400 LOC cap: <n>
- Suite green at every boundary (checked via git checkout): <y/n, exceptions>

## Process fidelity

- Plan produced and halted for approval: <y/n> — corrections needed: <what>
- Halted at every stage boundary: <y/n, exceptions>
- Ledger kept and accurate: <y/n>
- Sign-off requests (test edits / signature changes): <count, were they legitimate?>
- Discipline breaks observed (see the variant's README for its list): <list>
- Times reviewer had to intervene beyond the script: <count, what>

## Commit quality (read `git log` as an inheriting maintainer)

- Commits look human-maintained (imperative subject, brief prose body, describes the change not the process): <y/n>
- Methodology vocabulary in messages ("hole", "fill", "stage", "ledger", "census", spec refs): <none / list offenders>
- Process-cadence commits collapsed where the skill prescribes it (e.g. holes endgame squash): <y/n/n-a>
- Would a stranger reading the log suspect a generator? <y/n, why>

## Defects that escaped to final inspection

<list, or "none">

## Verdict

<3–5 sentences: was the work correct, reviewable, and honest? What was this variant's characteristic failure or strength in this run?>
