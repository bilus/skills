# todo.org

A live page for a project's `todo.org`. The file keeps the current work as org headings, and `plan.py` serves it at `http://localhost:4010` with a stable five-character id on every entry. An id such as `#ce96d` in any title or body becomes a link to its entry, so a chat message, a commit or the pinned `* Next steps` list can point at an entry by its id.

The script comes from live-templ, where it tracks the book's stages. It needs Python 3 and nothing else, rereads `todo.org` every two seconds while the page is open, and listens on 127.0.0.1 only.

## What the agent does

It copies `plan.py` into the project's `tools/`, keeps `todo.org` out of git, and starts the file with a `* Next steps` heading that names the entries to pick from by id. It takes every id from `plan.py --ids` and refreshes the references after a retitle, because an id is a hash of the entry's outline path. In the Claude desktop app, a `.claude/launch.json` entry with `autoPort` gives each project's page its own port.

## Installing

Clone the collection, as the top-level README describes, or link this directory into `~/.claude/skills/`.
