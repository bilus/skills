# Skills

A collection of [Claude skills](https://www.anthropic.com/news/skills) by [@bilus](https://github.com/bilus).

Each skill is a self-contained directory teaching Claude a specific pattern, workflow, or domain. Skills are loaded on demand and only when relevant — they don't bloat every conversation.

## Available skills

| Skill | What it teaches |
|-------|-----------------|
| [`three-layer-cake-haskell`](./three-layer-cake-haskell) | The three-layer cake architecture for Haskell apps: capability typeclasses, `MonadError` for domain errors, IO exceptions for infra, with testing, alternatives, handles, and compensating actions. |
| [`tracer-bullet-delivery`](./tracer-bullet-delivery) | Staged delivery for coding agents: walking skeleton first, then one vertical slice per stage, hard LOC budget, halt for human review at every stage boundary. |
| [`hole-driven-delivery`](./hole-driven-delivery) | Staged delivery for coding agents: commit a compiling skeleton of named `HOLE(id)` stubs as the design review, then fill holes outside-in with the build never red; `grep` is the progress bar. |
| [`test-ratchet-development`](./test-ratchet-development) | Staged delivery for coding agents: each stage opens with one failing acceptance test; the suite only ever tightens — no editing, weakening, or skipping tests without human sign-off. |
| [`avoiding-ai-tells`](./avoiding-ai-tells) | Style rules that strip machine-generation signals (uniform comments, gold-plating, template READMEs, mega-commits) from agent-written code, docs, and commits so reviewers judge the work on its merits. Referenced by all three staged-delivery skills. |
| [`prose-contract`](./prose-contract) | A strict style contract for agent-written prose: outline-review-draft-verify process, actor discipline, a banlist and transition caps checked by script, and 28 judgment rules checked by a reviewer sub-agent. The prose counterpart to `avoiding-ai-tells`. |

The three staged-delivery skills are an experiment: an identical plan/checkpoint outer loop with three different inner growth strategies. Enable **only one at a time** (their triggers overlap) and see [`evals/staged-delivery/`](./evals/staged-delivery) for the standard test task and how to run the comparison. Each has a README explaining the approach.

More to come.

## Installing a skill

### Claude Code

Clone the whole collection and symlink the skills you want, or clone individual skills directly:

```bash
# Whole collection (one skill per subdirectory)
git clone https://github.com/bilus/skills.git ~/.claude/skills/bilus

# Or just one skill
mkdir -p ~/.claude/skills
curl -L https://github.com/bilus/skills/archive/refs/heads/main.tar.gz | \
  tar xz --strip-components=1 -C ~/.claude/skills \
  skills-main/three-layer-cake-haskell
```

Claude Code discovers any directory with a `SKILL.md` under `~/.claude/skills/`, so nested layouts work fine.

### Claude.ai

Each tagged release attaches `.skill` files (zipped skill directories) as assets. Download the one you want from the [latest release](../../releases/latest) and upload via **Settings → Capabilities → Skills**.

### Building `.skill` files locally

A `.skill` file is just a zip of a skill's directory with a different extension:

```bash
cd three-layer-cake-haskell
zip -r ../three-layer-cake-haskell.skill . -x "*.DS_Store"
```

The repo's [GitHub Actions workflow](.github/workflows/release.yml) builds and attaches all `.skill` files automatically on tagged releases.

## Repo layout

```
skills/
├── README.md                          # this file
├── LICENSE
├── CHANGELOG.md
├── .github/workflows/release.yml      # builds .skill files on tagged releases
└── <skill-name>/
    ├── SKILL.md                       # required: frontmatter + main content
    └── ...                            # supporting files referenced from SKILL.md
```

Each top-level directory containing a `SKILL.md` is treated as a skill. The directory name and the `name:` field in the frontmatter should match.

## Contributing

Issues and PRs welcome — especially:

- Concrete examples that expose gaps in an existing skill
- New skills for patterns you've found yourself explaining repeatedly to Claude
- Improvements to trigger descriptions if a skill over- or under-triggers

For new skills, follow the structure of an existing one: a single `SKILL.md` that's self-contained, with companion files linked from it for orthogonal concerns ("read this when…").

## License

MIT — see [LICENSE](LICENSE).
