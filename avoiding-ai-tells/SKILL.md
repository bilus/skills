---
name: avoiding-ai-tells
description: Use whenever generating or editing code, comments, tests, documentation, READMEs, commit messages, or PR descriptions that humans will review. Removes the stylistic signals of machine generation (uniform comments, gold-plating, template READMEs, mega-commits) so the work gets judged on its merits. Not about deception - disclose AI assistance where that is expected.
---

# Avoiding AI tells in generated code

Instructions for a code-generating AI whose output humans will review. Reviewers pattern-match on signs of machine generation, and once they spot them they turn hostile: they stop evaluating the substance, review the diff more aggressively, distrust the parts that are fine, or dismiss the whole contribution as slop regardless of quality. Reviewing uniform machine output is also fatiguing in itself; the closer generated code looks to human-produced code, the less strain on the reviewer and the less hostility toward the work. The goal is not to deceive anyone about authorship (disclose AI assistance where that is expected) but to remove the stylistic noise that triggers the reaction, so the code gets judged on its merits. This pairs with staged delivery: small, human-shaped increments are easier to review than one large polished drop.

## Comments (the loudest tell)

* Do not comment uniformly. The strongest single tell is a tidy explanatory comment on every function, branch, and constant, in the same voice, across every file. Human codebases have a comment gradient: dense where the author struggled or got burned, silent for two hundred lines where the code was easy. Comment only where a future reader would otherwise be misled or have to re-derive something: the workaround, the protocol quirk, the invariant not visible in the code. If in doubt, delete the comment.
* Never narrate the code. `// increment the counter` above `counter++` is the fastest giveaway. Delete any comment that restates what the line, the function name, or the signature already says.
* No section banner comments (`// --- STATE ---`, `# ===== Helpers =====`) or decorative dividers unless the codebase already uses them.
* Avoid the assistant register: "Note that...", "This ensures...", "It's important to...", "First, we... / Next, we... / Finally, we...", "Handle edge case where...", "gracefully", "robust", "for clarity", docstrings that begin "This function...". Write flat: `// 404 here means the session expired`, not `// Note: We handle the 404 case gracefully to ensure robustness`.
* Leave no conversational residue or meta-comments about the generation process: "Here's the implementation", "As requested", "You can customize this", "In a real application you would...".
* Do not narrate changes in comments (`// changed from X to fix Y`, `// updated to use the new API`) or address the reviewer ("we could also..."). Comments talk to future maintainers, not to the person who requested the change.
* No apology or hedge comments: `// simplified for brevity`, `// in production you would...`, or `// TODO: add proper error handling` on code you were asked to finish.
* Do not justify routine decisions to an imagined reviewer ("Deliberately empty: ...", "This is safe because..."). One or two of these at genuinely tricky spots is human; ten per file is not. Same for habitual "NOTE:", "IMPORTANT:", "Edge case:" prefixes.
* Do not carry one language's comment idioms into another: Go's "funcName reports whether..." phrasing in JavaScript, JSDoc blocks in Python, PEP-257 phrasing in TypeScript.
* Skip comments in empty catch blocks and other no-op sites ("// fall through to defaults") unless the emptiness is truly non-obvious.
* No provenance breadcrumbs a human would not write: sha256 of a vendored file in a header comment, links to the exact chat session.

## Typography

* ASCII only in code, comments, commit messages, and docs unless the codebase already uses Unicode. No em dashes, no typographic ellipsis, no middle dots, no curly quotes. These are the single most pattern-matched tell in prose.
* In UI strings, prefer plain ASCII (`-`, `...`) unless the product's existing copy style says otherwise.
* No emoji anywhere: not in code, commit messages, READMEs, CLI or log output, unless the project already uses them.
* In prose, ration em dashes and semicolons; models overuse both.

## Scope and completeness

* Do not gold-plate. Breadth-first completeness on the first commit (every edge case handled, every header set, every input validated, defensive try/catch everywhere, CSP headers, pagination, i18n hooks) reads as checklist-generated. Humans accrete edge-case handling through bug reports; models emit it all at once. Ship what the task needs; hardening arrives when there is a reason. Leave a TODO where a human plausibly would.
* Do exactly what was asked. Unrequested extras are the structural tell: a config system, a retry wrapper, a logging layer, abstraction layers, helper utilities "for future use", a CLI flag nobody asked for, backwards-compat shims for code with no users. Propose extras in the PR description instead of silently including them.
* Do not defensively handle impossible cases: try/except around code that cannot throw, null checks on values guaranteed non-null two lines up, validating arguments the type system or a trusted internal caller already guarantees, exhaustive enum arms for cases that cannot occur. Humans add defense where they have been burned; match the codebase's actual level of paranoia.
* No premature abstraction: no interface for one implementation, no constant for one use, no generic for one type, no config object with one field, no classes or factories where a function will do. Rule of three applies.
* Avoid the kitchen-sink error path: error strings that embed every variable in scope, handlers that log, wrap, re-raise, and return a fallback all at once.
* Do not restructure, reformat, or "fix" code adjacent to the change. Drive-by cleanups, import reordering, and style fixes the task did not require inflate the diff and are a known agent habit. A 40-line change arriving inside a 400-line diff makes reviewers hostile before they read anything.
* Match the existing codebase's conventions exactly: naming, error handling style, test style, comment density, quoting, type-hint usage, even when they conflict with best practice. Generic best practice inside a codebase with its own idiom is itself a tell; inconsistency with the host file is more suspicious than any individual choice.
* Do not ship a fully-formed repo skeleton (CHANGELOG at v0.1.0 in Keep-a-Changelog format, badges, CONTRIBUTING, CODE_OF_CONDUCT, issue templates, elaborate CI) with the first commit of a small project. Documentation more complete than the project is mature is a tell.

## Naming and structure

* Avoid model-favorite identifiers: `handleX`/`processX`/`performX` for everything, `data`, `result`, `item`, `temp`, `helper`, `MyClass`, `Enhanced`/`Advanced`/`Improved`/`Smart` prefixes, `utils.py`/`helpers.js` dumping grounds. Prefer the project's existing vocabulary and names a domain expert would pick: shorter, more specific, occasionally idiosyncratic.
* Avoid LLM vocabulary in identifiers, comments, and docs: comprehensive, robust, seamless, leverage, utilize, streamlined, enhanced, cutting-edge, powerful, simply, "handles X gracefully", "best practices", "production-ready".
* Avoid the reflexive triple: three examples, three bullets, three test cases, three config options, everywhere. Vary structure the way the problem actually demands.
* Do not produce perfectly parallel code: three branches with identically shaped bodies, every function exactly one screen long, symmetric guard clauses, every module with the same section order. Real codebases are lumpy; if one case is trivial, let it be trivial, and let one function run long when splitting it was not worth it.

## Error handling and logging

* Handle errors where something can be done about them; let the rest propagate. Catch-and-log at every level reads as generated boilerplate.
* Error messages: lowercase, terse, specific (`"parse config: %w"`), not apologetic prose ("An unexpected error occurred while attempting to...").
* No log lines narrating normal operation ("Successfully initialized the service").

## Tests

* Test names describe behavior, not mechanical enumeration: no `test_case_1`, no `test_valid_input` / `test_invalid_input` / `test_edge_case` in lockstep, no mirrored `test_foo_returns_bar_when_baz` for every function.
* Do not pad the suite: no tests for getters or that a constructor assigns fields, no exhaustive permutation or symmetric coverage for trivial code. Humans test the behaviors that can actually break and the bugs they feared, and include one weird case a human would think of.
* A giant table-driven suite covering every input permutation, delivered in the same commit as the feature, reads as generated. If that thoroughness is wanted, say so in the commit message.
* Never fake coverage: no asserting on mocks of the thing under test, no tests that cannot fail, no assertions that restate the implementation. Test at least one case the implementation could plausibly get wrong.

## Documentation and READMEs

* The most recognized tell: a README with a badge row, emoji feature list, "Getting Started" rocket, marketing prose ("blazingly fast", "seamlessly integrates", "powerful and flexible"), and Installation / Usage / Contributing / License in canonical order, plus a table of contents for 60 lines of text. Write the paragraphs a maintainer would actually write; let structure appear when content demands it.
* No boilerplate sections nobody asked for: Contributing, Roadmap, Acknowledgements, FAQ on a 200-line library.
* No rule-of-three sentences ("simple, fast, and reliable"), no "whether you're X or Y" constructions, no "In today's fast-paced world" openers, no filler transitions ("Additionally,", "Furthermore,", "It's worth noting"), no hedged non-claims ("should generally work", "aims to provide").
* Write terse and specific. A short README with one accurate example beats a structured template. A rough edge (an idiom, a terse aside, a first-person remark) reads more human than uniform smoothness.
* Do not add a README, CHANGELOG, or LICENSE to someone else's repo unprompted. Do not add a README section, table, or example for every feature the moment it exists.
* Tables sparingly; a doc that expresses everything as tables is a tell.
* Leave no generation artifacts: CLAUDE.md, .claude/ directories, "Generated with" footers.

## Commits, history, and PRs

* No `Co-Authored-By: Claude` or "Generated with" trailers when the author does not want attribution advertised (respect explicit project policy if it requires disclosure).
* Commit messages: short imperative subject, plain and specific ("Fix array row edits writing to orphan keys"), not a summary of everything the diff touches. No bullet-point bodies enumerating every hunk, no "This commit..." openers, no conventional-commit decoration the repo does not already use. A body is warranted only when diagnosing something ("the row id was parsed as a flat key because..."). Occasional imperfect grammar is more credible than uniform polish.
* No mega-commits: a single "Initial implementation" landing thousands of lines of finished, tested, themed code plus docs and CI is the loudest possible history tell. Commit in the ragged, uneven increments real work produces, including the commit that fixes a bug a test caught.
* Cadence matters: three polished 200-line features committed two minutes apart is agent cadence, and velocity shows through timestamps. Do not fake dates or spacing; if the work was fast, let it be fast openly.
* Never rewrite history to simulate incremental work. Fabricated splits leak everywhere: config for features that arrive in later commits, diffs shifting between force-pushes, backdated committer timestamps, old objects still fetchable by SHA, the force-push visible in the activity feed. A discovered fake history is far worse than an honest squash. A squashed initial commit is normal; a synthetic diary is fraud.
* PR descriptions: state what changed and why in a paragraph sized to the diff. No "## Summary / ## Changes / ## Testing" template unless the repo requires it, no "This PR introduces..." followed by a bulleted essay for a small change.

## What not to fake

* Do not inject deliberate typos, sloppy code, misindented comments, or fabricated struggle. Injected noise degrades the work, reads as what it is once anything else raises suspicion, and poisons trust worse than polish does. The uneven texture must come from real judgment (commenting only what needs it, testing only what matters), not from simulated humanity.
* When AI assistance was substantial, one plain sentence of disclosure in the commit message or PR ("Tests generated with <tool>; they caught one real bug") converts a "gotcha" into a non-event. Reviewers punish concealment, not assistance.

## The meta-rule

Uniformity is the tell. Human code is locally inconsistent: comment density varies, some names are lazy, some edges are rough, effort concentrates where the problem was hard. When output is uniformly commented, uniformly defensive, uniformly structured, and uniformly polished, at the same temperature in every file, reviewers smell a generator even when every individual line is fine. Vary effort with difficulty, and when in doubt, do less.

The one-line version: write the smallest diff that does the job, commented only where a maintainer would actually pause, committed the way the work actually happened, described in plain prose, and when the machine did a lot of it, say so once, briefly.
