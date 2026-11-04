# Review brief

Start one sub-agent (Agent tool, general-purpose) per review point with this
brief, filled in. Give it the browser tools when the application has a UI.
Run it in the repository so it can start the application and run the tests.
Wait for its findings file before continuing.

```
You are reviewing a change by exploratory testing. You did not write it and
you have not seen it being written. Your job is to find where it does not
meet its requirement, not to restate what it does.

Requirement:
<the EARS sentences, numbered>

Acceptance tests:
<test names and the files they are in>

Plan:
<path to docs/plans/<date>-<slug>.md; read it first>

Stages under review:
<stage numbers and descriptions, with their commit ranges>

How to run the application:
<exact command, URL or entry point, test fixtures or mock data in use>

How to run the tests:
<exact command>

Do this, in this order:

1. Run the full test suite. Quote the summary line.
2. Start the application. Exercise the behavior each requirement sentence
   describes, by hand: call the endpoint, run the command, or drive the UI
   with the browser tools. Record what you did and what you saw.
3. Go beyond the happy path: empty input, the largest input you can make,
   repeated calls, the wrong tenant, a missing file, a request during
   another request, whatever the requirement makes plausible. Record each
   attempt and its result.
4. For a UI: every screen the stages touched, every control on it, with
   the keyboard as well as the mouse, and the browser console open. Note
   anything that is wrong or unfinished, including mock data that leaks
   into the screen in a way the plan does not mention.
5. Only now read the diff of the stages under review. Look for behavior
   the tests do not pin and for docstrings that describe internals or
   exceed 15 words without saying something the caller needs.

Write your findings to <path to docs/plans/<date>-<slug>/review-<n>.md>
in this form, and nothing else in the file:

# Review <n>: stages <range>

Suite: <quoted summary line>

## Blockers
<the application does not run, an acceptance test fails, or a requirement
sentence is not met for its main case; one per heading, with the exact
steps to reproduce and what you saw>

## Defects
<a requirement sentence is not met for some input, or behavior outside the
plan's stated reductions is broken; same form>

## Notes
<anything else worth knowing: unclear names, a docstring over the limit,
a test that does not pin what its name says, mock data left where the plan
does not list it>

## Not covered
<what you could not test and why>

Rules: reproduce every blocker and defect before writing it down. Give the
input, the command or the click path, and the output. Do not fix anything.
Do not propose designs. Grade by what the user would see, not by how the
code looks. If the application cannot be started, that is the first and
only blocker; say what you tried.
```

What to do with the findings is in SKILL.md, section 7.
