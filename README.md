# Skills

A collection of [Claude skills](https://www.anthropic.com/news/skills) by [@bilus](https://github.com/bilus).

Each skill is a self-contained directory teaching Claude a specific pattern, workflow, or domain. Skills are loaded on demand and only when relevant — they don't bloat every conversation.

## Available skills

| Skill | What it teaches |
|-------|-----------------|
| [`three-layer-cake-haskell`](./three-layer-cake-haskell) | The three-layer cake architecture for Haskell apps: capability typeclasses, `MonadError` for domain errors, IO exceptions for infra, with testing, alternatives, handles, and compensating actions. |

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
