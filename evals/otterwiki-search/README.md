# otterwiki-search evaluation

How the query procedure in `personal/otterwiki-search` was chosen, and how to rerun the comparison.

## What's here

```
evals/otterwiki-search/
├── README.md          # this file
├── scenarios.json     # 12 questions with primary and adjacent answer pages, fixed before any run
├── heldout.json       # 6 more, fixed before the refined instruction set existed
├── strategies.md      # the five instruction sets as given to the agents
├── score.py           # scorer: 1 for a primary page in the top three, 0.5 for an adjacent page, 0 otherwise
└── results/           # one JSON per run: queries issued, top three pages, a note
```

Result file names: `P0` to `P4` are the five strategies on `scenarios.json` with the full index; `H_P1`, `H_P3`, `H_P4` the same on `heldout.json`; `R2_*` after Index, Home, The Inbox and the log pages left the Qmd index; `R3_*` after the six cluster hubs left it as well. P4 is the procedure the skill teaches; `SKILL_H` is the shipped `SKILL.md` itself, read from disk by a fresh agent, on the held-out set.

## Scoring

```bash
python3 score.py results/P1.json results/P4.json
SCEN=heldout.json python3 score.py results/H_P1.json results/H_P4.json
```

## Rerunning

One run is one sub-agent with the otterwiki MCP tools `wiki_semantic_search`, `wiki_search`, `wiki_tags` and `wiki_list` only, no `wiki_read`, so the score measures retrieval rather than reading. Give it the strategy text (for the skill, the SKILL.md body), the question list, and this output contract:

```
Reply with ONLY a JSON array, one object per scenario in order:
{"id":"S1","queries":[{"tool":"wiki_semantic_search","args":{"query":"...","limit":5}}, ...],
 "top3":["page/a","page/b","page/c"],"note":"one short sentence"}
```

Tell it to treat scenarios independently, to list every call it made with exact arguments, and to retry a transient server error once. Save the array under `results/` and score it. Single-call strategies vary by about half a point between runs on the agent's wording alone, so compare runs on the same index and treat half-point differences as noise.

## Results

Full index, 209 pages:

| Strategy | Calls per question | 12 questions | 6 held out |
|---|---|---|---|
| P0 question verbatim | 1.0 | 9.5 | not run |
| P1 keyword rewrite | 1.0 | 10.0 | 5.5 |
| P2 fan-out, vote-count merge | 5.0 | 10.0 | not run |
| P3 keyword call, then harvested terms | 2.2 | 11.0 | 6.0 |
| P4 the procedure | 1.4 | 11.0 | 6.0 |

P1 and P4 after each index change:

| Index | P1, 12 + 6 | P4, 12 + 6 |
|---|---|---|
| all 209 pages | 10.0 + 5.5 | 11.0 + 6.0 |
| without navigation pages (192) | 9.5 + 5.5 | 11.5 + 6.0 |
| without the six hubs as well (186) | 9.0 + 5.5 | 12.0 + 6.0 |
