# svg-diagrams

Draw system diagrams as SVG built by a small script, in a flat style that stays
readable when it lands in a document.

Most generated diagrams fail the same way: a different colour per box, arrows
crossing in the middle, and a legend explaining what the labels already say.
This skill is the set of decisions that avoids that, written down so an agent
makes them the same way twice.

## What it teaches

A build approach. The diagram is a Python script that prints one self-contained
SVG, so moving a box in six months is a one-line edit rather than archaeology in
a wall of coordinates.

An order of work. List the components and every connection as data first, choose
the grid from the connection count rather than the component count, render, then
look at the result and count the crossings.

Layout heuristics. Columns by role in the direction work flows; boxes ordered
within a column so line sets stay parallel; barriers such as proxies and
gateways drawn as one tall panel spanning what they front rather than as another
box in a row; out-of-band lines dashed and routed around the barrier.

A style contract. Background, panel tint, rules, a seven-step type scale, a
five-colour accent palette, and the marks: filled-triangle arrowheads, numbered
step circles, chips, and the background-coloured rectangle painted behind any
label that crosses a line.

A rule for colour. Colour means one thing chosen in advance, or it is noise.
Grey is a real choice for anything the reader should see but not study.

## What is in the box

`SKILL.md` is the skill. `scripts/svgkit.py` holds the primitives with the style
baked in: canvas, box, barrier panel, arrow with arrowhead, numbered badge,
chip, and the label backing. Import it or copy from it.

## Installing

```bash
mkdir -p ~/.claude/skills
curl -L https://github.com/bilus/skills/archive/refs/heads/main.tar.gz | \
  tar xz --strip-components=1 -C ~/.claude/skills skills-main/svg-diagrams
```

## A note on scope

The skill ends with the case for not drawing at all. Two boxes and an arrow is a
sentence, and a picture of a list is a list. A diagram earns its place when the
reader needs to see a shape: a boundary several paths cross, a cycle, a fan-out,
two things that look adjacent and are not.
