---
name: todo-org
description: Use when a project keeps its current work in todo.org, when the user asks to keep a TODO list in todo.org or to see it as a live page, and before adding, finishing, retitling or moving entries of todo.org.
---

# todo.org with a live page

## Overview

`todo.org` at the project root holds the current work as org headings. `plan.py`, in this skill's directory, serves it as a live HTML page that gives every entry a stable id, such as `#ce96d`, and links each `#id` it finds in the text. Copy the script; do not write another viewer.

## Setup

1. From the project root: `mkdir -p tools && cp <this skill's directory>/plan.py tools/plan.py`. The script reads `todo.org` from the parent of its own directory, and it needs Python 3 alone.
2. Add `/todo.org` to `.gitignore`, and leave `tools/plan.py` untracked.
3. Start `todo.org` with a top-level `* Next steps` heading.
4. In the Claude desktop app, add this entry to `.claude/launch.json` and start it with `preview_start` by the name `plan`. `autoPort` passes a free port in `$PORT`. If the file holds no other entry, add `/.claude/launch.json` to `.gitignore` too, since its entry needs the untracked script.

   ```json
   {"name": "plan", "runtimeExecutable": "python3", "runtimeArgs": ["tools/plan.py"], "port": 4010, "autoPort": true}
   ```

   In a terminal: `python3 tools/plan.py`, with `--port N` when 4010 is taken. The server runs until stopped, so never start it in the foreground of a tool call.

## Conventions

- An entry is an org heading. Its state is `TODO`, `DONE` or none. The page reads any other keyword, `NEXT` included, as part of the title.
- Tags go at the end of the heading, as `:issue13:`.
- The page pins `* Next steps` above the outline. It lists the entries to pick from, one per line: `- #ce96d Requirements for issue 13.`
- Never write or guess an id. Run `python3 tools/plan.py --ids` and copy it. `--find ID` prints one entry with its body.
- An id is a hash of the entry's outline path: its title and the titles above it. It survives body edits, `TODO` turning `DONE` and moves among siblings. Retitling the entry or a heading above it gives it a new id, so rerun `--ids` and update every `#id` that pointed at it. The page shows an unknown `#id` as plain text instead of a link.
- An entry's body says what the work is and, once decided, why one approach beat the obvious one.
- When an entry is done, mark it `DONE` and drop its line from `* Next steps`.

## Common mistakes

| Mistake | Fix |
|---|---|
| Writing a server, an exporter or a renderer | Copy `plan.py` |
| A `NEXT` or `WAITING` keyword | `TODO` plus a line in `* Next steps` |
| An id typed from memory | `python3 tools/plan.py --ids` |
| A stale `#id` after a retitle | Rerun `--ids` and update the references |
| Two entries with the same outline path, shown as a red id | Retitle one |
| `todo.org` committed | `/todo.org` in `.gitignore` |
| A second project's server on port 4010 | `autoPort` in `launch.json`, or `--port` |
