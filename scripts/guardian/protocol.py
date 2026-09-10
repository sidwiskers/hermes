"""Official documentation evidence, independent of the Go structural parser."""
import hashlib
import json
import re
from html.parser import HTMLParser


def canonical(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), ensure_ascii=False)


def digest(value):
    return hashlib.sha256(canonical(value).encode()).hexdigest()


class Document(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.sections = {}
        self.current = None
        self.heading = False
        self.table = 0
        self.row = None
        self.cell = None
        self.skip = 0

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if tag in ("script", "style"):
            self.skip += 1
        if tag in ("h1", "h2", "h3", "h4"):
            self.current = None
            self.heading = True
        if self.heading and tag == "a" and attrs.get("name"):
            key = attrs["name"]
            if key in self.sections:
                raise ValueError("duplicate documentation anchor")
            self.current = {"title": [], "body": [], "rows": []}
            self.sections[key] = self.current
        if self.current is None:
            return
        if tag == "table":
            self.table += 1
        if tag == "tr":
            self.row = []
        if tag in ("td", "th"):
            self.cell = []
        # Preserve destinations in explanatory prose (not cosmetic anchor links).
        if tag == "a" and not self.heading and attrs.get("href"):
            self.handle_data(" [" + attrs["href"] + "] ")

    def handle_endtag(self, tag):
        if tag in ("script", "style"):
            self.skip = max(0, self.skip - 1)
        if tag in ("h1", "h2", "h3", "h4"):
            self.heading = False
        if self.current is None:
            return
        if tag in ("td", "th") and self.cell is not None:
            if self.row is not None:
                self.row.append(" ".join(" ".join(self.cell).split()))
            self.cell = None
        if tag == "tr" and self.row is not None:
            self.current["rows"].append(self.row)
            self.row = None
        if tag == "table":
            self.table = max(0, self.table - 1)

    def handle_data(self, data):
        if self.current is None or self.skip:
            return
        if self.heading:
            self.current["title"].append(data)
        elif self.cell is not None:
            self.cell.append(data)
        elif not self.table:
            self.current["body"].append(data)


def snapshot(html):
    parser = Document()
    parser.feed(html)
    sections = {}
    for anchor, section in parser.sections.items():
        title = " ".join(" ".join(section["title"]).split())
        # Release announcements are retained separately as evidence, not mixed
        # with each declaration: every new release naturally changes them.
        if re.match(r"^[A-Z][a-z]+ \d{1,2}, \d{4}$", title):
            continue
        sections[anchor] = {
            "title": title,
            "body": " ".join(" ".join(section["body"]).split()),
            "rows": section["rows"],
        }
    if "sendmessage" not in sections or "update" not in sections:
        raise ValueError("official document is incomplete or its layout changed")
    return {"format": 1, "sections": sections}


def semantic_changes(before, after, structural):
    """A description, return contract, link, or existing row change needs review.

    Permit new optional fields only where the structural parser independently
    classified the update as mechanical. Never infer behavior from prose.
    """
    if not before or before.get("format") != 1:
        return ["Documentation baseline needs review"]
    reasons = []
    old, new = before["sections"], after["sections"]
    added_objects = {n.lower() for n in structural["objects"]["added"]}
    changed_objects = {n.lower() for n in structural["objects"]["changed"]}
    for name in sorted(old.keys() | new.keys()):
        if name not in old:
            if name not in added_objects:
                reasons.append(name + ": new documentation section")
            continue
        if name not in new:
            reasons.append(name + ": documentation section removed")
            continue
        left, right = old[name], new[name]
        if left == right:
            continue
        if (structural["classification"] == "mechanical" and name in changed_objects
                and left["title"] == right["title"] and left["body"] == right["body"]):
            rows = {r[0]: r for r in right["rows"] if r}
            if all(r and rows.get(r[0]) == r for r in left["rows"]):
                continue
        reasons.append(name + ": description, return contract, or field rules changed")
    return reasons


def validate_candidate(before, after):
    if after.get("source") != "https://core.telegram.org/bots/api":
        raise ValueError("unexpected schema source")
    version = lambda s: tuple(map(int, s.split(".")))
    if not re.fullmatch(r"\d+\.\d+", after.get("version", "")):
        raise ValueError("invalid Bot API version")
    if version(after["version"]) < version(before["version"]):
        raise ValueError("Telegram source moved backwards; refusing stale snapshot")
    for category in ("methods", "objects", "unions"):
        old, new = before[category], after[category]
        if len(new) < len(old) * .9:
            raise ValueError("large schema loss in " + category + "; inspect source/parser")
        names = [x["name"] for x in new]
        if len(names) != len(set(names)):
            raise ValueError("duplicate schema names")
