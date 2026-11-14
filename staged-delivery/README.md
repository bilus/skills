# Staged delivery

A delivery workflow for AI coding agents, built on Martin Fowler's
[example of preparatory refactoring](https://martinfowler.com/articles/preparatory-refactoring-example.html).
[SKILL.md](SKILL.md) holds the agent's rules. This file explains why the
workflow has its shape and what it asks of you.

## Why this approach

Review capacity is the bottleneck with agents, so the skill spends it in
two places and nowhere else. You approve the plan once, before the first
code change. After that you read the commit log the way you would read a
colleague's history: every commit is small, does one thing and says so in
its summary.

Preparatory refactoring keeps the commits small. The agent first changes
the structure of the code without changing its behavior, with the existing
suite as the proof, until the code has a place for the change. The change
itself is then a small commit.

## What you do

- Approve the plan. Spend your time on the stage table: the order and the
  "why here" column are where a wrong plan shows.
- Read the commits as they land, by their summaries.
- Answer when the agent stops. Section 10 of SKILL.md lists the two cases.

## Requirements

Codex as an MCP server for the prose review, and a `REQUIREMENTS.yaml`
managed by `phb`. The skill uses each when present and works without
either.
