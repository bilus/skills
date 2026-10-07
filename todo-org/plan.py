#!/usr/bin/env python3
"""Serve todo.org as a live HTML page, one stable id per entry.

    python3 tools/plan.py [--port N]      serve http://localhost:N
    python3 tools/plan.py --ids            print each entry's id and title
    python3 tools/plan.py --find ID        print one entry with its body

An entry's id is the first five hex digits of the SHA-1 of its outline path
(the titles from the top-level heading down, joined by "/"), so it survives
edits to the body, a TODO turning DONE and the entries around it moving. A
retitled entry gets a new id. The page polls the file every two seconds and
re-renders when it changed.

The server listens on --port, else on $PORT, else on 4010. The page's title
names the directory that holds todo.org.

The page pins the top-level "* Next steps" heading, with any entries under it,
above the outline, and shows a hint when todo.org has no such heading. A
"#id" of a known entry in any title or body becomes a link that opens the
entry, scrolls to it and marks it for a moment.
"""
import hashlib
import html
import json
import os
import re
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
TODO = os.path.join(ROOT, "todo.org")
HEADING = re.compile(r"^(\*+)\s+(?:(TODO|DONE)\s+)?(.*?)(?:\s+(:[\w@:.-]+:))?\s*$")


def parse(text):
    """Return the entries of an org file as a list of dicts, in document order."""
    entries, stack = [], []
    for line in text.split("\n"):
        m = HEADING.match(line)
        if not m:
            if entries:
                entries[-1]["body"].append(line)
            continue
        level, state, title, tags = len(m.group(1)), m.group(2), m.group(3).strip(), m.group(4)
        stack = stack[: level - 1] + [title]
        path = "/".join(stack)
        entries.append({
            "id": hashlib.sha1(path.encode()).hexdigest()[:5],
            "level": level,
            "state": state or "",
            "title": title,
            "tags": [t for t in (tags or "").strip(":").split(":") if t],
            "body": [],
        })
    for e in entries:
        e["body"] = "\n".join(e["body"]).strip("\n")
    ids = {}
    for e in entries:
        ids.setdefault(e["id"], []).append(e["title"])
    for e in entries:
        e["duplicate"] = len(ids[e["id"]]) > 1
    return entries


def load():
    with open(TODO, encoding="utf-8") as f:
        return parse(f.read())


PAGE = """<!doctype html>
<meta charset="utf-8">
<title>{{PROJECT}} plan</title>
<style>
  :root { --bg: #fff; --fg: #1a1a1a; --dim: #777; --done: #2a7; --todo: #c60; --id: #06c; --line: #e5e5e5; --code: #f3f3f3; --hl: #fff3b0; }
  @media (prefers-color-scheme: dark) { :root { --bg: #161616; --fg: #e6e6e6; --dim: #999; --done: #5c9; --todo: #f93; --id: #6bf; --line: #333; --code: #262626; --hl: #4a3f10; } }
  body { margin: 0; font: 14px/1.45 -apple-system, system-ui, sans-serif; background: var(--bg); color: var(--fg); }
  header { position: sticky; top: 0; background: var(--bg); border-bottom: 1px solid var(--line); padding: 8px 16px; display: flex; gap: 16px; align-items: center; flex-wrap: wrap; }
  header input[type=search] { flex: 1; min-width: 160px; padding: 4px 8px; font: inherit; }
  main { padding: 8px 16px 48px; }
  .entry { margin: 2px 0; }
  .l1 { margin-left: 0; } .l2 { margin-left: 20px; } .l3 { margin-left: 40px; } .l4 { margin-left: 60px; } .l5 { margin-left: 80px; }
  .head { display: flex; gap: 8px; align-items: baseline; cursor: pointer; }
  .id { font: 12px ui-monospace, SFMono-Regular, Menlo, monospace; color: var(--id); user-select: all; }
  .id.dup { color: #d33; }
  .state { font: 600 11px ui-monospace, Menlo, monospace; }
  .state.DONE { color: var(--done); } .state.TODO { color: var(--todo); }
  .l1 > .head > .title { font-weight: 700; font-size: 16px; }
  .tags { color: var(--dim); font: 12px ui-monospace, Menlo, monospace; }
  .body { display: none; margin: 4px 0 8px 52px; color: var(--dim); white-space: pre-wrap; max-width: 80ch; }
  .entry.open > .body { display: block; }
  code { background: var(--code); padding: 0 3px; border-radius: 3px; font-size: 12.5px; color: var(--fg); }
  .hidden { display: none; }
  .entry.hl > .head { background: var(--hl); border-radius: 4px; margin-left: -6px; padding: 2px 6px; }
  .entry.hl > .head > .title::after { content: " proposed"; color: var(--todo); font: 600 11px ui-monospace, Menlo, monospace; margin-left: 6px; }
  .count { color: var(--dim); }
  #next { margin: 12px 16px 0; padding: 8px 14px 10px; border: 1px solid var(--line); border-left: 4px solid var(--todo); border-radius: 6px; }
  #next h2 { margin: 0 0 4px; font-size: 15px; display: flex; gap: 8px; align-items: baseline; }
  #next .body { display: block; margin: 4px 0 0; color: var(--fg); white-space: normal; }
  #next p { margin: 0 0 4px; }
  #next ul { margin: 0; padding-left: 20px; }
  #next li { margin: 3px 0; }
  #next .missing { color: var(--dim); }
  a.ref { color: var(--id); font: 12px ui-monospace, SFMono-Regular, Menlo, monospace; text-decoration: none; }
  a.ref:hover { text-decoration: underline; }
  .entry.flash > .head { background: var(--hl); border-radius: 4px; }
</style>
<header>
  <strong>todo.org</strong>
  <label><input type="checkbox" id="hideDone"> hide DONE</label>
  <label><input type="checkbox" id="openAll"> open bodies</label>
  <input type="search" id="q" placeholder="filter by id, title, tag">
  <span class="count" id="count"></span>
</header>
<section id="next"></section>
<main id="plan"></main>
<script>
  let last = "";
  const hl = new Set((new URLSearchParams(location.search).get("hl") || "").split(",").filter(Boolean));
  let firstRender = true;
  let known = new Set();
  const esc = s => s.replace(/[&<>"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;"}[c]));
  const markup = s => esc(s).replace(/=([^=\\n]+)=/g, "<code>$1</code>").replace(/~([^~\\n]+)~/g, "<code>$1</code>")
    .replace(/\\[\\[[^\\]]*\\]\\[([^\\]]*)\\]\\]/g, "$1")
    .replace(/#([0-9a-f]{5})\\b/g, (m, id) => known.has(id) ? `<a class="ref" href="#e-${id}" data-ref="${id}">#${id}</a>` : m);
  function listMarkup(body) {
    const paras = [], items = [];
    for (const line of body.split("\\n")) {
      if (/^\\s*- /.test(line)) { items.push(line.replace(/^\\s*- /, "")); continue; }
      if (/^\\s+\\S/.test(line) && items.length) { items[items.length - 1] += " " + line.trim(); continue; }
      if (line.trim()) paras.push(line.trim());
    }
    return (paras.length ? `<p>${markup(paras.join(" "))}</p>` : "") +
      (items.length ? `<ul>${items.map(i => `<li>${markup(i)}</li>`).join("")}</ul>` : "");
  }
  function renderNext(next) {
    const box = document.getElementById("next");
    if (!next.length) {
      box.innerHTML = `<h2>Next steps</h2><div class="body missing">Add a "* Next steps" heading to todo.org.</div>`;
      return;
    }
    const [head, ...items] = next;
    box.innerHTML = `<h2>Next steps <span class="id">#${head.id}</span></h2>` +
      (head.body ? `<div class="body">${listMarkup(head.body)}</div>` : "") +
      items.map(e => `<div class="body"><a class="ref" href="#e-${e.id}" data-ref="${e.id}">#${e.id}</a> ${markup(e.title)}${e.body ? listMarkup(e.body) : ""}</div>`).join("");
  }
  function render(all) {
    known = new Set(all.map(e => e.id));
    const start = all.findIndex(e => e.level === 1 && e.title.toLowerCase() === "next steps");
    let end = start + 1;
    while (start >= 0 && end < all.length && all[end].level > 1) end++;
    renderNext(start >= 0 ? all.slice(start, end) : []);
    const entries = start >= 0 ? all.slice(0, start).concat(all.slice(end)) : all;
    const hideDone = document.getElementById("hideDone").checked;
    const openAll = document.getElementById("openAll").checked;
    const q = document.getElementById("q").value.trim().toLowerCase();
    const opened = new Set([...document.querySelectorAll(".entry.open")].map(e => e.dataset.id));
    let shown = 0;
    const out = entries.map(e => {
      const hay = (e.id + " " + e.state + " " + e.title + " " + e.tags.join(" ")).toLowerCase();
      const hidden = (hideDone && e.state === "DONE") || (q && !hay.includes(q));
      if (!hidden) shown++;
      const open = openAll || opened.has(e.id) || (hl.has(e.id) && firstRender);
      return `<div class="entry l${Math.min(e.level, 5)}${hidden ? " hidden" : ""}${open ? " open" : ""}${hl.has(e.id) ? " hl" : ""}" id="e-${e.id}" data-id="${e.id}">
        <div class="head" onclick="this.parentNode.classList.toggle('open')">
          <span class="id${e.duplicate ? " dup" : ""}" title="${e.duplicate ? "two entries share this id: retitle one" : "id"}">#${e.id}</span>
          ${e.state ? `<span class="state ${e.state}">${e.state}</span>` : ""}
          <span class="title">${markup(e.title)}</span>
          ${e.tags.length ? `<span class="tags">:${e.tags.join(":")}:</span>` : ""}
        </div>
        ${e.body ? `<div class="body">${markup(e.body)}</div>` : ""}
      </div>`;
    }).join("");
    document.getElementById("plan").innerHTML = out;
    if (firstRender) { const first = document.querySelector(".entry.hl"); if (first) first.scrollIntoView({block: "center"}); firstRender = false; }
    const todo = entries.filter(e => e.state === "TODO").length, done = entries.filter(e => e.state === "DONE").length;
    document.getElementById("count").textContent = `${shown} shown, ${todo} TODO, ${done} DONE`;
  }
  async function poll() {
    try {
      const r = await fetch("/plan.json", {cache: "no-store"});
      const text = await r.text();
      if (text !== last) { last = text; render(JSON.parse(text)); }
    } catch (e) {}
  }
  document.addEventListener("click", ev => {
    const a = ev.target.closest("a.ref");
    if (!a) return;
    ev.preventDefault();
    ev.stopPropagation();
    const el = document.getElementById("e-" + a.dataset.ref);
    if (!el) return;
    el.classList.remove("hidden");
    el.classList.add("open", "flash");
    el.scrollIntoView({block: "center"});
    setTimeout(() => el.classList.remove("flash"), 1500);
  }, true);
  for (const id of ["hideDone", "openAll", "q"]) document.getElementById(id).addEventListener("input", () => { if (last) render(JSON.parse(last)); });
  poll(); setInterval(poll, 2000);
</script>
"""


class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.startswith("/plan.json"):
            body = json.dumps(load()).encode()
            ctype = "application/json"
        else:
            body = PAGE.replace("{{PROJECT}}", html.escape(os.path.basename(ROOT))).encode()
            ctype = "text/html; charset=utf-8"
        self.send_response(200)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Cache-Control", "no-store")
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *args):
        pass


def main(argv):
    if "--ids" in argv:
        for e in load():
            print(f"#{e['id']}  {'*' * e['level']} {e['state']:4} {e['title']}")
        return 0
    if "--find" in argv:
        wanted = argv[argv.index("--find") + 1].lstrip("#")
        for e in load():
            if e["id"] == wanted:
                print(f"#{e['id']}  {'*' * e['level']} {e['state']} {e['title']}\n{e['body']}")
                return 0
        print(f"no entry #{wanted}", file=sys.stderr)
        return 1
    port = int(argv[argv.index("--port") + 1]) if "--port" in argv else int(os.environ.get("PORT", "4010"))
    server = ThreadingHTTPServer(("127.0.0.1", port), Handler)
    print(f"plan at http://localhost:{port}/", flush=True)
    server.serve_forever()


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
