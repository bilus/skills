# OtterWiki search

A query procedure for an agent answering questions from my OtterWiki over its MCP server. It exists because the wiki's semantic search runs without a reranker: the Qmd cross-encoder cost about 110 seconds per query on the CPU-only host, so retrieval is reciprocal rank fusion of BM25 and vector search, and the query string is the only lever. Phrased as a question, a query hands the top ranks to long pages that share its words. Phrased as two to four terms in the answer page's own vocabulary, it finds the page in one call.

## What the agent does

1. One `wiki_semantic_search` with two to four distinctive terms, no question words, no context words the target's neighbours share.
2. If the top page's snippet addresses the question, take it. A page that only shares a word with the query does not count.
3. Otherwise one follow-up: the page a snippet links to, a `wiki_list` by namespace or tag when results cluster, or one rephrasing. Never a vote-count merge across many queries, and never a later call displacing a rank-1 answer.
4. `wiki_read` the chosen page before answering.

## How it was measured

Five instruction sets ran as sub-agents over 12 questions with known answer pages, then the two best over 6 more questions fixed before the refined set existed, then again after each change to the index. The procedure scored 12 of 12 and 6 of 6 at 1.3 calls per question on the current index; the question verbatim scored 9.5 of 12; a five-query fan-out scored 10 of 12 at five times the calls. Scenarios, scorer and every run's output are in [`evals/otterwiki-search/`](../../evals/otterwiki-search).

## Why it is under personal/

It is tied to one wiki, one MCP server ([OtterWikiMcp](https://github.com/bilus/OtterWikiMcp)) and one index configuration: navigation pages and cluster hubs are excluded from the Qmd index, and reranking is off. The procedure transfers to any fusion-only hybrid search, but the page names and the exclusion list do not.

## Installing

```bash
git clone https://github.com/bilus/skills.git ~/.claude/skills/bilus
```

Claude Code discovers `personal/otterwiki-search/SKILL.md` inside the clone. The release workflow only packages top-level skills, so this one has no `.skill` asset.
