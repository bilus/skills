# Prose review through Codex

Every piece of prose the user will read goes to Codex before the user sees
it. Codex reviews; this skill applies the result. Codex never edits files.

## The call

Use the Codex MCP server (the `codex` tool). If the tool is not among the
loaded tools, load it with ToolSearch first. If the server is not
connected, tell the user once, mark the plan "Codex pass: not run" and
continue.

```
codex (MCP)(
  prompt: <the prompt below, with the text and context inlined>,
  sandbox: "read-only",
  approval-policy: "never",
  cwd: "<absolute path to the repository>"
)
```

Rules that do not change:

- `cwd` is the repository, never the plan folder or a temp directory, so
  Codex can read the code the prose talks about.
- `sandbox` is `read-only`.
- `approval-policy` is `never`, or the call blocks on a prompt nobody sees.
- The prompt is self-contained. Inline the text under review and the
  context. Codex does not read files outside `cwd`, and a path is not
  context.

## The prompt

```
Review the text below as an editor, not as an author. Do not rewrite it;
list what to change, with the exact sentence and the replacement.

Context: <what the text is: a plan for <change> in this repository, the
docstrings added in stages <n>-<m>, the commit messages of stages <n>-<m>.
Name the files and functions the text refers to so you can check it
against the code.>

Check two things.

1. Does it read as if a human developer on this project wrote it? Flag
   anything that does not: filler and preambles, hedging, words no
   developer would use for the thing, marketing or blog register,
   metaphors and personification in technical sentences, lists where a
   sentence would do, sentences that restate the previous one, praise of
   the work, and any non-ASCII punctuation (dashes, curly quotes,
   ellipsis). Flag commit messages over 15 words, stage descriptions over
   20 words, and docstrings over 15 words that do not tell the caller
   something they need.

2. Is it factually right against the code in this repository? Check every
   claim that names a file, function, type, behavior, command or number.
   Quote the claim and say what the code actually does.

Report format: one item per line, "<where> | <problem> | <replacement>".
Group factual errors first. End with "FACTUAL ERRORS: <count>".

Text under review:

<the text>
```

## Applying the result

- Apply every factual correction. Apply the readability items unless one
  would break a rule in SKILL.md (word limits, vocabulary from the
  diagrams, ASCII only); then keep the rule and note the item as not
  applied.
- If `FACTUAL ERRORS` is above zero, send the corrected text once more with
  the same context. The correction is new prose. Stop after the second pass
  whatever it says; list anything still open in the plan under
  "Changes to the plan".
- Batches: the plan and all step summaries before the plan is shown; the
  docstrings and plan text written since the last pass at each review
  point; whatever is left at the end. Do not send a commit message on its
  own.
