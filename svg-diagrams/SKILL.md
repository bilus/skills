---
name: svg-diagrams
description: Draw system diagrams as SVG built by a small script - components, services, datastores, trust boundaries, and the arrows between them - in a flat style that survives being pasted into a document. Use this whenever someone asks for a diagram, an architecture picture, a system or component overview, a deployment or data-flow sketch, a sequence or interaction diagram, or says "draw me how X works" or "show me the architecture". It applies whether or not they say SVG, and it applies especially when a new diagram has to sit beside an existing one and match it.
---

# System diagrams as SVG

A diagram earns its place by answering a question faster than prose can. Most
generated diagrams fail at that: every box a different colour, arrows crossing
in the middle, a legend explaining what the labels already say. This skill is
the discipline that avoids it.

Announce at the start: "Using svg-diagrams: parts list, then grid, then render,
then check crossings."

## Build it with a script

Write a small Python program that prints SVG to a file. Do not hand-write the
SVG, and do not reach for a diagramming library.

The reason is maintenance. A diagram lives next to a document that changes, and
in six months someone will need to move one box. A script makes that a one-line
edit and a re-run. Hand-written SVG makes it an archaeology exercise, because
every coordinate is a magic number with no name. A library makes the layout
somebody else's opinion, and you lose the one thing that makes these diagrams
readable, which is deliberate placement.

`scripts/svgkit.py` in this skill has the primitives with the house style baked
in: canvas, box, arrow with an arrowhead, numbered badge, chip, and the label
backing that keeps text readable where it crosses a line. Import it or copy the
parts you need. What matters is that the output is one self-contained SVG file
with no external fonts, images or scripts, so it renders anywhere and can be
rasterised for a Word document.

## Work in this order

**List the parts before drawing anything.** Write two lists in the script, as
data: the components, and every connection as (from, to, what it carries,
traffic or provisioning). Laying out before you know the connection count is
how diagrams end up with a line routed around three sides of the canvas.

**Choose the grid from the connection list, not from the component list.** Count
which component has the most connections and build the grid outward from it.

**Render, then look at it.** Open the SVG. Count the crossings. If two lines
cross that could have been avoided by swapping two boxes in a column, swap them.
This pass takes a minute and is the difference between a diagram someone reads
and one they skip.

**Read it at a third of its size.** If the shape of the system is not apparent
when the labels are illegible, the layout is doing no work and no amount of
styling will save it.

## Layout

Columns by role, left to right, in the direction the work flows: actors outside
the system, then any boundary they cross, then the services, then the stores the
services own. A reader who knows nothing about the system should be able to
guess the direction of causality from the shape alone.

Order the boxes within a column to minimise crossings. The heuristic that does
most of the work: put each service directly across from whatever talks to it
most, and when one actor talks to two services and another talks to a third,
group them so those line sets stay parallel instead of interleaving.

A barrier is not a box in a row. Proxies, gateways, ingress controllers and
trust boundaries front several things at once, so draw them as one tall panel
spanning everything they front, in a tint distinct from the services. Every
request line then visibly passes through it, which is the fact the diagram
exists to convey. A proxy drawn as a box the same size as a service says the
opposite of what is true.

Route out-of-band lines around the barrier rather than through it, and dash
them. Provisioning, key handover and configuration are not request traffic, and
a reader who sees a solid line through a proxy will reasonably assume the proxy
handles it.

Keep lines straight or orthogonal. A diagonal crossing a box reads as a mistake.
Where two lines genuinely must cross, cross them at a right angle: a shallow
crossing looks like a join.

## The style contract

These values come from a diagram pair that has survived review. They are a
starting point rather than scripture, but change them as a set, because a
half-changed palette looks like a bug.

```
font-family     Helvetica Neue, Helvetica, Arial, sans-serif
background      #FBFBF9
panel tint      #F2F5F1
rules, borders  #D6DDD6 at 1px, dashes as stroke-dasharray="3 4"

title           28px / 700 / #14201B
subtitle        15px / #56635C
section label   10.5px / 700 / #56635C, letter-spacing 1.2, uppercase
box title       13px / 700 / #14201B
box detail      11.5px / #6B776F
arrow label     11px / #2B352E
chip text       10.5px / 700 / #FFF
footnote        11.5px / #6B776F
```

An accent palette of five, which is enough for any diagram that a person can
read and more than most need:

```
#1B5E4A  green
#2E6E8E  blue
#4A3A8C  purple
#8A5A1B  amber
#7A2E4A  maroon
```

Marks:

- Connectors at `stroke-width="1.7"`, coloured by what they carry.
- Arrowheads as a filled triangle, 8 long and 4.2 across, in the line's colour.
  A stroked V-shape looks thin beside a 1.7px line.
- Numbered steps as a filled circle of `r="9"` with white 10.5px bold text.
- Short labels as a `rect` of `width="27" height="17" rx="3"` with white bold text.
- Behind any text that sits over a line, paint a background-coloured rectangle
  first. This one trick does more for legibility than any font choice.

## Colour carries meaning or it is noise

Pick what colour means before you use it, and write the rule in the subtitle or
a small legend. One good choice is to colour by what a line carries: one hue for
credentials, another for tokens, another for user data, grey for infrastructure.
Another is to colour by trust or ownership, so a reader sees which parts belong
to whom.

What fails is colouring by identity, a different hue per box. It looks designed
and conveys nothing, and it uses up the reader's attention on decoration before
they reach the content.

Grey is a real choice. Datastores, infrastructure and anything the reader should
see but not study belong in grey, which leaves the accent colours meaningful.

## Titles

Give the diagram a title naming the system and one subtitle line saying the
thing a reader should leave with. The subtitle is where the argument goes: not
"components and their interactions", which says nothing, but the claim the
picture supports, such as which path carries the authoritative copy, or where
the one boundary is that everything crosses.

## Keep out

No gradients, no drop shadows, no rounded-everything, no icons or clip art, no
emoji, no 3D, no hand-drawn effect. No legend for things the labels already say.
No decorative colour.

These are not aesthetic preferences. Each one costs contrast, ink or attention
that the diagram needs for its content, and each one is a signal of a picture
made to look impressive rather than to be read.

## When a diagram is the wrong answer

Two boxes and an arrow is a sentence. A picture of a list is a list. A diagram
earns its place when the reader needs to see a shape: a boundary crossed by
several paths, a cycle, a fan-out, a place where two things that look adjacent
are not. If you cannot name the shape, write the prose instead and save
everyone the render.
