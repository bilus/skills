import json, sys, re
from collections import defaultdict

import os
SCEN_FILE = os.environ.get("SCEN", os.path.join(os.path.dirname(os.path.abspath(__file__)), "scenarios.json"))
if not os.path.isabs(SCEN_FILE) and not os.path.exists(SCEN_FILE):
    SCEN_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), SCEN_FILE)
scen = {s["id"]: s for s in json.load(open(SCEN_FILE))}

def norm(p):
    p = p.strip().lower()
    p = re.sub(r"^/+", "", p)
    p = re.sub(r"\.md$", "", p)
    return p

def score_one(s, top3):
    top = [norm(p) for p in top3[:3]]
    prim = {norm(p) for p in s["primary"]}
    alt = {norm(p) for p in s["alt"]}
    if any(p in prim for p in top): return 1.0
    if any(p in alt for p in top): return 0.5
    return 0.0

def load(path):
    txt = open(path).read()
    # tolerate code fences or stray prose around the array
    m = re.search(r"\[.*\]", txt, re.S)
    return json.loads(m.group(0))

rows = {}
for path in sys.argv[1:]:
    name = path.split("/")[-1].replace(".json", "")
    res = load(path)
    by_id = {r["id"]: r for r in res}
    total = 0.0; calls = 0; by_type = defaultdict(list); per = {}
    for sid, s in scen.items():
        r = by_id.get(sid)
        sc = score_one(s, r["top3"]) if r else 0.0
        n = len(r["queries"]) if r else 0
        total += sc; calls += n
        by_type[s["type"]].append(sc)
        per[sid] = (sc, n)
    rows[name] = (total, calls, by_type, per)

ids = list(scen)
print(f"strategy   total/{len(scen)}  calls  " + "  ".join(f"{t[:4]:>4}" for t in ["lexical","task","conceptual"]))
for name, (total, calls, by_type, per) in rows.items():
    parts = "  ".join((f"{sum(by_type[t])/len(by_type[t]):>4.2f}" if by_type[t] else "   -") for t in ["lexical","task","conceptual"])
    print(f"{name:<10} {total:>7.1f}  {calls:>5}  {parts}")
print()
print("per scenario (score/calls):")
print("      " + "  ".join(f"{n:>9}" for n in rows))
for sid in ids:
    cells = "  ".join(f"{rows[n][3][sid][0]:>4.1f}/{rows[n][3][sid][1]:<3}" for n in rows)
    print(f"{sid:<5} {cells}   [{scen[sid]['type'][:4]}]")
