#!/usr/bin/env python3
"""Generate raster brand assets from SVG sources."""

from __future__ import annotations

import subprocess
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
BRAND = ROOT / "frontend/public/brand"
PUBLIC = ROOT / "frontend/public"
ASSETS = ROOT / "assets"
APP = ROOT / "frontend/app"

TEAL = (13, 148, 136, 255)
TEXT = (22, 32, 51, 255)
FONT_CANDIDATES = [
    "/System/Library/Fonts/Supplemental/Arial Bold.ttf",
    "/System/Library/Fonts/Supplemental/Arial.ttf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
    "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
]


def run_magick(args: list[str]) -> None:
    subprocess.run(["magick", *args], check=True)


def load_font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    paths = FONT_CANDIDATES if bold else FONT_CANDIDATES[1:]
    for path in paths:
        if Path(path).exists():
            return ImageFont.truetype(path, size)
    return ImageFont.load_default()


def render_icon_png(dest: Path, size: int) -> None:
    src = BRAND / "logo-icon.svg"
    run_magick(["-background", "none", "-density", "384", str(src), "-resize", f"{size}x", str(dest)])


def render_banner_png(dest: Path, width: int) -> None:
    height = max(96, width // 5)
    icon_size = int(height * 0.72)
    canvas = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(canvas)

    icon_path = Path("/tmp/echostate-brand-icon.png")
    render_icon_png(icon_path, icon_size)
    icon = Image.open(icon_path).convert("RGBA")
    icon_y = (height - icon_size) // 2
    canvas.paste(icon, (int(height * 0.12), icon_y), icon)

    font_size = int(height * 0.42)
    wordmark_font = load_font(font_size, bold=True)
    text_x = int(height * 0.12) + icon_size + int(height * 0.14)
    baseline = int(height * 0.67)

    draw.text((text_x, baseline), "EchoState", font=wordmark_font, fill=TEXT, anchor="ls")

    canvas.save(dest, "PNG", optimize=True)


def main() -> None:
    ASSETS.mkdir(parents=True, exist_ok=True)

    render_banner_png(PUBLIC / "logo-banner.png", 640)
    render_banner_png(ASSETS / "logo-banner.png", 640)
    render_icon_png(PUBLIC / "logo.png", 256)
    render_icon_png(ASSETS / "logo.png", 256)
    render_icon_png(APP / "apple-icon.png", 180)
    render_icon_png(APP / "icon.png", 32)

    run_magick(
        [
            str(APP / "icon.png"),
            "-define",
            "icon:auto-resize=64,48,32,16",
            str(APP / "favicon.ico"),
        ]
    )

    print("Brand assets generated.")


if __name__ == "__main__":
    main()