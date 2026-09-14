"""Primitives for the diagrams this skill describes.

Import it, or copy the four or five functions you need. It exists so that every
diagram gets the same arrowheads, the same label backing and the same type
scale without anyone reconstructing them from the style contract each time.

    from svgkit import Canvas, INK, ACCENT

    c = Canvas(1200, 700, "Consent flow", "One write path, one cached read.")
    c.box(60, 140, 220, 90, "Receiver", ["device credential", "manufacturer key"])
    c.box(420, 140, 220, 90, "OAuth service", ["holds the public key"])
    c.arrow(280, 185, 420, 185, "signed assertion", ACCENT["amber"])
    open("flow.svg", "w").write(c.render())
"""

from html import escape

PAPER = "#FBFBF9"
PANEL = "#F2F5F1"
RULE = "#D6DDD6"
INK = "#14201B"
BODY = "#2B352E"
MUTED = "#56635C"
FAINT = "#6B776F"

ACCENT = {
    "green": "#1B5E4A",
    "blue": "#2E6E8E",
    "purple": "#4A3A8C",
    "amber": "#8A5A1B",
    "maroon": "#7A2E4A",
}

FONT = "Helvetica Neue, Helvetica, Arial, sans-serif"


class Canvas:
    def __init__(self, width, height, title=None, subtitle=None):
        self.w, self.h = width, height
        self.out = []
        self.out.append(
            f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" '
            f'viewBox="0 0 {width} {height}" font-family="{FONT}">'
        )
        self.out.append(f'<rect width="{width}" height="{height}" fill="{PAPER}"/>')
        if title:
            self.text(48, 58, title, 28, INK, weight=700)
        if subtitle:
            self.text(48, 86, subtitle, 15, MUTED)

    def raw(self, markup):
        self.out.append(markup)

    def text(self, x, y, content, size=11, fill=BODY, weight=None, anchor=None, spacing=None):
        bits = [f'x="{x}"', f'y="{y}"', f'font-size="{size}"', f'fill="{fill}"']
        if weight:
            bits.append(f'font-weight="{weight}"')
        if anchor:
            bits.append(f'text-anchor="{anchor}"')
        if spacing:
            bits.append(f'letter-spacing="{spacing}"')
        self.out.append(f'<text {" ".join(bits)}>{escape(content)}</text>')

    def box(self, x, y, w, h, title, lines=(), accent=None, fill="#FFFFFF"):
        """A component. `accent` tints the left edge, which is how a box says
        what kind of thing it is without spending a fill colour on it."""
        self.out.append(
            f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="3" '
            f'fill="{fill}" stroke="{RULE}"/>'
        )
        if accent:
            self.out.append(
                f'<rect x="{x}" y="{y}" width="3.5" height="{h}" fill="{accent}"/>'
            )
        self.text(x + 16, y + 26, title, 13, INK, weight=700)
        for i, line in enumerate(lines):
            self.text(x + 16, y + 46 + i * 16, line, 11.5, FAINT)

    def panel(self, x, y, w, h, label=None):
        """A barrier: a proxy, a gateway, a trust boundary. Tall and spanning,
        so that lines crossing it are visibly crossing something."""
        self.out.append(
            f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="3" '
            f'fill="{PANEL}" stroke="{RULE}"/>'
        )
        if label:
            cx, cy = x + w / 2, y + h / 2
            self.out.append(
                f'<text x="{cx}" y="{cy}" font-size="12.5" font-weight="700" fill="{MUTED}" '
                f'text-anchor="middle" transform="rotate(-90 {cx} {cy})" '
                f'letter-spacing="1.2">{escape(label)}</text>'
            )

    def arrow(self, x1, y1, x2, y2, label=None, colour=FAINT, dashed=False, badge=None):
        dash = ' stroke-dasharray="3 4"' if dashed else ""
        self.out.append(
            f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" '
            f'stroke="{colour}" stroke-width="1.7"{dash}/>'
        )
        self.arrowhead(x2, y2, x2 - x1, y2 - y1, colour)
        if badge is not None:
            self.badge((x1 + x2) / 2, (y1 + y2) / 2, badge, colour)
        elif label:
            self.label((x1 + x2) / 2, min(y1, y2) - 9, label)

    def arrowhead(self, x, y, dx, dy, colour):
        """Filled triangle. A stroked V reads thin next to a 1.7px line."""
        if abs(dx) >= abs(dy):
            d = 1 if dx >= 0 else -1
            self.out.append(
                f'<path d="M{x} {y} L{x - 8 * d} {y - 4.2} L{x - 8 * d} {y + 4.2} z" fill="{colour}"/>'
            )
        else:
            d = 1 if dy >= 0 else -1
            self.out.append(
                f'<path d="M{x} {y} L{x - 4.2} {y - 8 * d} L{x + 4.2} {y - 8 * d} z" fill="{colour}"/>'
            )

    def label(self, x, y, content, size=11, fill=BODY):
        """Text with the paper painted behind it, so it stays readable where it
        sits over a line. Width is estimated; check anything unusually long."""
        width = len(content) * size * 0.52
        self.out.append(
            f'<rect x="{x - width / 2 - 4}" y="{y - size + 1}" width="{width + 8}" '
            f'height="{size + 5}" fill="{PAPER}"/>'
        )
        self.text(x, y, content, size, fill, anchor="middle")

    def badge(self, x, y, number, colour):
        self.out.append(f'<circle cx="{x}" cy="{y}" r="9" fill="{colour}"/>')
        self.text(x, y + 3.6, str(number), 10.5, "#FFF", weight=700, anchor="middle")

    def chip(self, x, y, content, colour):
        self.out.append(
            f'<rect x="{x}" y="{y}" width="27" height="17" rx="3" fill="{colour}"/>'
        )
        self.text(x + 13.5, y + 13, content, 10.5, "#FFF", weight=700, anchor="middle")

    def render(self):
        return "\n".join(self.out + ["</svg>", ""])
