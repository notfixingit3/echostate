#!/usr/bin/env python3
"""Generate PWA icons (any + maskable) and Apple touch icon from assets/logo.png."""

from __future__ import annotations

from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
LOGO_PATH = ROOT / "assets" / "logo.png"
ICONS_DIR = ROOT / "frontend" / "public" / "icons"
APPLE_ICON = ROOT / "frontend" / "app" / "apple-icon.png"

THEME_BG = (13, 148, 136, 255)  # #0d9488
LIGHT_BG = (248, 250, 252, 255)  # #f8fafc
BLACK_BG = (0, 0, 0, 255)


def render_icon(size: int, background: tuple[int, int, int, int], logo_scale: float, out: Path) -> None:
    logo = Image.open(LOGO_PATH).convert("RGBA")
    canvas = Image.new("RGBA", (size, size), background)
    logo_size = max(1, int(size * logo_scale))
    resized = logo.resize((logo_size, logo_size), Image.Resampling.LANCZOS)
    offset = ((size - logo_size) // 2, (size - logo_size) // 2)
    canvas.paste(resized, offset, resized)
    out.parent.mkdir(parents=True, exist_ok=True)
    canvas.save(out, format="PNG", optimize=True)


def main() -> None:
    if not LOGO_PATH.is_file():
        raise SystemExit(f"logo not found: {LOGO_PATH}")

    # Standard launcher icons (full-bleed logo on black, matching existing branding).
    render_icon(192, BLACK_BG, 0.86, ICONS_DIR / "icon-192.png")
    render_icon(512, BLACK_BG, 0.86, ICONS_DIR / "icon-512.png")

    # Maskable safe-zone icons (~52% logo in center on theme fill).
    render_icon(192, THEME_BG, 0.52, ICONS_DIR / "icon-192-maskable.png")
    render_icon(512, THEME_BG, 0.52, ICONS_DIR / "icon-512-maskable.png")

    # Apple touch icon — light background, padded logo for iOS home screen.
    render_icon(180, LIGHT_BG, 0.58, APPLE_ICON)

    print("Wrote PWA icons to", ICONS_DIR)
    print("Wrote Apple touch icon to", APPLE_ICON)


if __name__ == "__main__":
    main()