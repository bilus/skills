---
name: otterwiki-search
description: Use when answering a question from the OtterWiki knowledge base through its MCP tools (wiki_semantic_search, wiki_search, wiki_list, wiki_read), before the first search call, and whenever a search returns source summaries, long analysis pages or nothing on topic instead of the page holding the answer.
---

# OtterWiki search

## Overview

`wiki_semantic_search` fuses a BM25 lexical index and a vector index with equal weight and no reranker. The query string you write is the only thing either index sees. Two to four terms in the answer page's own vocabulary find it in one call; a natural-language question or a padded phrase hands the top ranks to long pages that share its words.

## When to use

- Any question the wiki might answer, before the first search.
- A search whose top results are source summaries, long analysis pages or sibling pages rather than the page holding the answer.
- A cluster question (reading order, what a namespace covers), which search cannot answer because hub and navigation pages are not indexed.

Not for writing to the wiki; that follows the wiki's own AGENTS page.

## How retrieval behaves

Qmd runs the query against BM25 over the page text and against an `embeddinggemma-300M` vector index, then merges the two ranked lists by reciprocal rank fusion. The `score` in a result is a fusion rank (1, 0.5, 0.33 and so on), not a relevance estimate. Index, Home, The Inbox, the log pages and the cluster hubs (Objlog, ThreeLayerCake, Dokploy, Writing, BricksAndMortarPhilosophy, ErgoFramework) are excluded from the index; `wiki_search` and `wiki_list` still see them.

Three consequences, and every rule below follows from one of them:

1. Words shared by many pages (a language name, "log", "system", "agent", "code", question words) pull in source summaries and long pages, because BM25 rewards term frequency.
2. Short pages lose to long ones on the lexical side, so a short concept or tips page can rank below a long analysis page that only links to it.
3. A page at rank 1 for a precise two-word query is strong evidence. A page that appears across many loose queries is not; long pages that mention everything do the same.

## The procedure

At most three tool calls per question.

1. **One `wiki_semantic_search`, `limit` 5, two to four terms** in the vocabulary of the expected answer page: product or pattern names, exact identifiers or error strings, likely title words, the wiki's own abbreviations ("infra", not "infrastructure"). No question words. No context words unless one of them is the distinguishing term.
2. **If the top page's snippet addresses the question, take it and stop.** A page that only shares a word with the query does not count: the word "discipline" in a query put an unrelated page whose title contains it at rank 1.
3. **Otherwise one targeted follow-up**, chosen by the evidence in hand:
   - A snippet holds a `[Title](/path)` link to a page that fits better than anything returned. That page is the candidate; one search on its title words confirms it.
   - The results cluster in one namespace or share a tag. Call `wiki_list` with that `prefix` or `tags` and `include_summary: true`, and choose by summary. This is how a short page surfaces from under its longer siblings.
   - Neither applies. One more `wiki_semantic_search` with a different two to four term phrasing.

Never rank candidates by how many calls returned them. Never let a later call displace a rank-1 page whose snippet answered the question. Then `wiki_read` the chosen page before answering; a snippet is three lines around one match, not enough material for an answer.

## Quick reference

| Need | Query shape | Example |
|---|---|---|
| A concept page | the concept's name or its two defining nouns | `Append Flush`, not `object storage log Append Flush` |
| A rule inside a cluster | the rule's own words, cluster name omitted | `domain errors infra errors`, not `three-layer cake domain errors infrastructure errors` |
| A tips page for a symptom | the tool names plus the symptom word | `gofumpt go-ts-mode doom` |
| A page you know by title | its title words | `bracket mask async exceptions` |
| A cluster's contents | `wiki_list` with `prefix` | `wiki_list(prefix: "objlog/", include_summary: true)` |

## Common mistakes

| Mistake | What happens | Fix |
|---|---|---|
| Padding a precise query with context words | the source page and long neighbours outrank the target: `Haskell bracket asynchronous exceptions mask` put the target fourth, `bracket mask async exceptions` put it first | drop every word the neighbours share |
| Sending the question verbatim | function words dilute the lexical half: 9.5 of 12 against 12 of 12 for the procedure | rewrite to terms first |
| Fanning out five phrasings and merging by vote count | long pages appear in every list and push out a page that was rank 1 in one of them: 10 of 12 at five times the calls | one precise call, then at most one follow-up |
| Keeping a rank-1 page that only shares a word with the query | an unrelated page whose title matches one query word | judge by the snippet, not by the rank |
| Letting a second query displace a rank-1 answer | a harvested-terms query returned only the survey pages around the answer | a rank-1 page whose snippet answers stays |

## Measured

18 questions with known answer pages (12 fixed before any run, 6 held out until after the procedure existed), each run a separate sub-agent with the search and list tools only. The procedure: 12 of 12 and 6 of 6 at 1.3 calls per question. The question verbatim: 9.5 of 12. A keyword rewrite alone: 10.0 of 12 and 5.5 of 6, varying by half a point between runs on wording alone. The eval lives in `evals/otterwiki-search/` in this repo.

## Tool names

The tools are `wiki_semantic_search`, `wiki_search`, `wiki_list`, `wiki_tags` and `wiki_read` on the otterwiki MCP server; the harness prefixes them with `mcp__<server>__`, where `<server>` is the name given at registration. If they are deferred, load them by name before the first call.
