# Prompts for sub-agents

Fill in the paths in angle brackets. Each prompt asks for findings, not edits, so you choose the changes and record your decision on each finding in the ledger.

## Vocabulary review

```
Review a glossary for redundancy and accuracy. Do not rewrite it; report.

Glossary: <path to docs/vocabulary.md>
Sources: <plan path>, <paths of the dfd files>, <paths of the code, if any exists>

1. Redundant terms: entries that duplicate another entry, that another entry plus ordinary
   language already covers, or that no source uses. Say "remove" or "merge into <term>", and why.
2. Wrong or imprecise definitions: quote the entry and the source that contradicts it.
3. Missing terms: words the plan or the diagrams use as terms of art without an entry.
4. One thing, several words: where the sources use different words for one thing, list the
   words and where each appears, so one can be chosen.

Return numbered findings with evidence.
```

## Design review

```
Review a leveled data flow design against its rules. Do not edit anything; report.

Rules: <this skill's directory>/design.md
Design: <paths of the dfd files>
Plan: <plan path>
Code: <paths of the skeleton and of every file the change touches>
Base: <the commit before the change, for a change to an existing design>
dfdmetrics report: <paste the output of dfdmetrics docs/flow.dfd>

1. Boxes: does the action line say what the referenced functions do? Does the purpose line
   point at a later step? Are the references in the last parentheses, from one design package?
   Does every reference resolve to a declaration in the code?
2. Arrows: does each label name what the next process receives, item for item with the next
   function's declared parameters and their types? Does any arrow claim a value
   that a step changes in place and does not return?
3. State: for each store item, which write and which read pair up? Which item could travel on
   an arrow instead? Which data is carried through a step that does not use it?
4. Levels: is each store drawn in the smallest process that uses it? Does each parent diagram
   show only the data and state that cross the boundary?
5. Metaphor: read the plan's Metaphor section first, before any diagram, and list its implied
   rules, as "stack" implies push and pop at one end. Check each rule against the design, level
   by level, and report each broken rule with the level of its first break. Is the section one
   familiar thing and three or four of its parts, in at most 60 words?
6. Changes: for a change to an existing design, run git diff <base> on the dfd files. Could a
   change inside a function's body replace a changed interface, or a change in a deeper
   diagram replace one in a higher diagram?

Return numbered findings with evidence, most severe first.
```
