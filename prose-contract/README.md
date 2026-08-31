# Prose Contract — README

A style contract for prose an agent writes on your behalf. It bans the vocabulary and sentence shapes that mark text as machine-generated, and it wraps the writing in a process that catches them: outline, review the outline, expand, review again, draft, then run a script and a reviewer sub-agent over the draft until both come back clean.

The contract carries no facts about your subject. It supplies only rules and procedure, and its examples deliberately come from cooking, gardening, and household machinery so nothing in it can leak into a document about something else. One rule exists purely to enforce that: no four-word run from the skill's own examples may appear in the output.

## What's in it

Three tiers of rules, split by how they get checked.

- **Section A: machine-checkable.** A hard banlist (delve, tapestry, meticulous, showcase, pave the way, and about forty more), an at-most-once list, a cap of two formal transitions per section, and mechanical bans on em dashes, curly quotes, emoji, bold bullet lead-ins, and title-case headings. A script finds every hit.
- **Section B: judgment rules.** Twenty-eight patterns that no regex catches, each with a bad and a good version: the trailing `-ing` clause that adds analysis, the padded rule of three, the unsourced "experts agree", the mirrored "not X but Y" contrast, the closing sentence that summarizes what you just read. A reviewer sub-agent applies these and reports rule id, quoted sentence, suggested rewrite.
- **Section C: leak check.** The four-word rule above.

Two rules run through everything else. Every sentence needs an actor doing something, where an object counts as an actor (the valve sticks, the letter arrives); if the assignment doesn't name the main actors, the agent stops and asks before outlining. And every number or factual claim needs a named source or gets deleted.

## What the agent will do

1. Verify the facts against your named sources first, so the outline rests on something checked.
2. Outline five to seven claims, one per paragraph, each a full sentence stating what that paragraph will prove.
3. Hand the outline to a reviewer sub-agent, fix what it returns, expand each claim into two to four sub-claims, review again.
4. Draft from the expanded outline.
5. Run sections A and C as a script, check the length with `wc`, then hand the draft and section B to a reviewer sub-agent. Loop until it returns zero violations, at most three rounds.
6. Output the document and nothing else.

## What you do

Write the assignment. The contract governs style and procedure only; the topic, length, framing, output format, sources, and any rule exemptions come from your prompt, and the skill reads it wherever it says "the assignment". Name your sources explicitly, since claims without one get dropped rather than softened. Declare exemptions when the piece needs them, for instance when a memoir's own dates and counts shouldn't need external citation.

Expect the actor question if your prompt doesn't answer it, and expect a shorter document than you asked for on the first pass, because the rules delete more than they rewrite.

## Usage

Save the skill and invoke it explicitly for a prose deliverable, or mention the prose contract in the request. It's deliberately not always-on: the process costs several sub-agent rounds and suits pieces where the writing is the deliverable, not a paragraph of chat.

The banlists are meant to be edited. Add the words your own drafts keep reaching for, and drop any rule that fights your voice, but keep the sub-agent review, since section B is where the tells that survive a grep actually die.
