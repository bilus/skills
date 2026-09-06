# Candidate instruction sets (fixed before any run)

Scoring per scenario: primary page in the agent's top 3 = 1.0; only an alt page = 0.5; neither = 0.
Also recorded: number of tool calls per scenario.

## P0 baseline
Call wiki_semantic_search once with the question verbatim, limit 5. Report the top pages.

## P1 keyword rewrite
Before searching, rewrite the question as a short phrase of 2 to 5 distinctive content words
(technical nouns, names, error strings). Drop question words and function words. Search once with
that phrase, limit 5. Report the top pages.

## P2 fan-out
Issue 3 to 5 different searches for the same need, each a separate wiki_semantic_search call:
(a) the question verbatim, (b) a 2-5 word keyword phrase, (c) one or two rephrasings in
alternative vocabulary the wiki might use, (d) optionally wiki_search full-text with the single
most distinctive term. Merge: rank pages by how many result lists they appear in, break ties by
best rank. Report the top 3.

## P3 harvest and refine
Search once with a keyword phrase (limit 5). Read the returned page names and snippets and extract
the wiki's own terminology (page names, section words, tags). Run a second wiki_semantic_search
using that terminology. Optionally call wiki_tags and wiki_list(tags) to narrow by tag. Report the
top 3 by accumulated evidence.
