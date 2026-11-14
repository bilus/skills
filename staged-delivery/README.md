# Staged delivery

A delivery workflow for AI coding agents, built on Martin Fowler's
[example of preparatory refactoring](https://martinfowler.com/articles/preparatory-refactoring-example.html).
[SKILL.md](SKILL.md) holds the agent's rules. This file explains why the
workflow has its shape and what it asks of you.

## Why this approach

Review capacity is the bottleneck with agents, so the skill spends it in
two places and nowhere else. You approve the plan once, before the first
code change. After that you review each stage as one pull request of a few
hundred lines, and read its commits the way you would read a colleague's
history: every commit is small, does one thing and says so in its summary.

Milestones give the commits their context. Each milestone ends in a win, a
new capability or a settled risk, and the riskiest part comes first, while
changing a wrong design still costs little. Feature flags keep each stage
mergeable before its feature is complete.

Preparatory refactoring keeps the commits small. The agent first changes
the structure of the code without changing its behavior, with the existing
suite as the proof, until the code has a place for the change. The change
itself is then a small commit.

## What you do

- Approve the plan. Spend your time on the milestone and stage tables:
  their order and their "why here" columns are where a wrong plan shows.
- Review each stage by its commit range in the plan, and its commits by
  their summaries.
- Open a pull request per stage, or tell the agent to open them.
- Answer when the agent stops. Section 10 of SKILL.md lists the two cases.

## Requirements

Codex as an MCP server for the prose review, and a `REQUIREMENTS.yaml`
managed by `phb`. The skill uses each when present and works without
either.
