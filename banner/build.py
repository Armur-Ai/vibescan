#!/usr/bin/env python3
"""Generates banner/vibescan-dark.svg and banner/vibescan-light.svg.

Run from the repo root:  python3 banner/build.py

The chips name scanners that are wired into the scan pipeline. Keep it that way:
if a tool is not called from internal/tasks, it does not go on the banner.
"""
from pathlib import Path

THEMES = {
    "dark": dict(
        bg="#0d1117", panel="#131a26", stroke="#2a3447", chip="#0f2a27", chip_stroke="#1f5c52",
        text="#e6edf3", muted="#8b949e", teal="#4fd1b5", amber="#f5a524", red="#f0645a",
        purple="#a78bfa", blue="#60a5fa", pill="#161d2b",
    ),
    "light": dict(
        bg="#ffffff", panel="#f6f8fa", stroke="#d0d7de", chip="#e6f6f2", chip_stroke="#8fd1c2",
        text="#1f2328", muted="#59636e", teal="#0f766e", amber="#b45309", red="#c2410c",
        purple="#6d28d9", blue="#1d4ed8", pill="#f6f8fa",
    ),
}

SOURCES = [("Cursor", "teal"), ("Claude Code", "amber"), ("Copilot", "purple"), ("Windsurf", "blue")]
SCANNERS = ["semgrep", "eslint", "bandit", "gosec", "trufflehog", "osv-scanner", "trivy", "checkov"]
MORE = "+ 29 more scanners"

MONO = "ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
SANS = "-apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif"


def build(t):
    out = []
    add = out.append
    add('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 400" width="1200" height="400" '
        'role="img" aria-label="vibescan: code from Cursor, Claude Code, Copilot and Windsurf flows '
        'through security scanners and comes out either ready to ship or flagged">')
    add(f"""<style>
  .flow {{ fill: none; stroke-width: 2; stroke-dasharray: 7 7; animation: flow 1.4s linear infinite; }}
  .chip {{ animation: pulse 3.2s ease-in-out infinite; }}
  .beam {{ animation: sweep 3.2s ease-in-out infinite; }}
  .hit {{ animation: blink 1.6s ease-in-out infinite; }}
  @keyframes flow {{ to {{ stroke-dashoffset: -28; }} }}
  @keyframes pulse {{ 0%, 100% {{ opacity: .55; }} 50% {{ opacity: 1; }} }}
  @keyframes sweep {{ 0% {{ transform: translateX(0); }} 50% {{ transform: translateX(384px); }} 100% {{ transform: translateX(0); }} }}
  @keyframes blink {{ 0%, 100% {{ opacity: .35; }} 50% {{ opacity: 1; }} }}
  @media (prefers-reduced-motion: reduce) {{ .flow, .chip, .beam, .hit {{ animation: none; opacity: 1; }} }}
</style>""")
    add(f'<rect width="1200" height="400" rx="16" fill="{t["bg"]}"/>')

    # Title
    add(f'<text x="600" y="62" text-anchor="middle" font-family="{MONO}" font-size="42" '
        f'font-weight="700" letter-spacing="3"><tspan fill="{t["teal"]}">vibe</tspan>'
        f'<tspan fill="{t["amber"]}">scan</tspan></text>')
    add(f'<text x="600" y="92" text-anchor="middle" font-family="{SANS}" font-size="17" '
        f'fill="{t["muted"]}">Security scanner for AI-generated code</text>')

    # Sources
    for i, (name, colour) in enumerate(SOURCES):
        y = 132 + i * 55
        cy = y + 19
        c = t[colour]
        add(f'<rect x="40" y="{y}" width="176" height="38" rx="19" fill="{t["pill"]}" stroke="{t["stroke"]}"/>')
        add(f'<circle cx="62" cy="{cy}" r="5" fill="{c}"/>')
        add(f'<text x="78" y="{cy + 5}" font-family="{SANS}" font-size="15" fill="{t["text"]}">{name}</text>')
        add(f'<path class="flow" stroke="{c}" style="animation-delay:-{i * 0.35}s" '
            f'd="M216,{cy} C 310,{cy} 310,235 400,235"/>')

    # Scanner panel
    add(f'<rect x="400" y="120" width="400" height="230" rx="14" fill="{t["panel"]}" stroke="{t["stroke"]}"/>')
    for i, name in enumerate(SCANNERS):
        x = 420 + (i % 2) * 186
        y = 138 + (i // 2) * 44
        add(f'<g class="chip" style="animation-delay:-{i * 0.4}s">'
            f'<rect x="{x}" y="{y}" width="174" height="34" rx="6" fill="{t["chip"]}" stroke="{t["chip_stroke"]}"/>'
            f'<text x="{x + 87}" y="{y + 22}" text-anchor="middle" font-family="{MONO}" font-size="14" '
            f'fill="{t["teal"]}">{name}</text></g>')
    add(f'<text x="600" y="336" text-anchor="middle" font-family="{MONO}" font-size="13" '
        f'fill="{t["muted"]}">{MORE}</text>')
    add(f'<rect class="beam" x="406" y="126" width="4" height="218" rx="2" fill="{t["amber"]}" opacity=".55"/>')

    # Outcomes
    add(f'<path class="flow" stroke="{t["teal"]}" d="M800,235 C 880,235 880,179 960,179"/>')
    add(f'<path class="flow" stroke="{t["red"]}" d="M800,235 C 880,235 880,291 960,291"/>')
    add(f'<rect x="960" y="160" width="200" height="38" rx="19" fill="{t["pill"]}" stroke="{t["teal"]}"/>')
    add(f'<text x="1060" y="184" text-anchor="middle" font-family="{SANS}" font-size="15" '
        f'fill="{t["teal"]}">ready to ship ✓</text>')
    add(f'<rect x="960" y="272" width="200" height="38" rx="19" fill="{t["pill"]}" stroke="{t["red"]}"/>')
    add(f'<circle class="hit" cx="984" cy="291" r="5" fill="{t["red"]}"/>')
    add(f'<text x="1068" y="296" text-anchor="middle" font-family="{SANS}" font-size="15" '
        f'fill="{t["red"]}">flagged before prod</text>')

    add("</svg>")
    return "\n".join(out) + "\n"


if __name__ == "__main__":
    here = Path(__file__).parent
    for name, theme in THEMES.items():
        (here / f"vibescan-{name}.svg").write_text(build(theme), encoding="utf-8")
        print(f"wrote banner/vibescan-{name}.svg")
