# -*- coding: utf-8 -*-
"""Иконка PulseShield: розовый скруглённый квадрат (squircle) #f07575, белые PS, прозрачность снаружи."""
from __future__ import annotations

import os
import sys

from PIL import Image, ImageDraw, ImageFont

SIZE = 512
# IconColor из пресета Gamingnew
FILL = (240, 117, 117, 255)
WHITE = (255, 255, 255, 255)


def find_font(size: int) -> ImageFont.FreeTypeFont:
    candidates = [
        os.path.join(os.environ.get("WINDIR", "C:\\Windows"), "Fonts", "segoeuib.ttf"),
        os.path.join(os.environ.get("WINDIR", "C:\\Windows"), "Fonts", "arialbd.ttf"),
        os.path.join(os.environ.get("WINDIR", "C:\\Windows"), "Fonts", "arial.ttf"),
    ]
    for path in candidates:
        if os.path.isfile(path):
            try:
                return ImageFont.truetype(path, size=size)
            except OSError:
                continue
    return ImageFont.load_default()


def main() -> None:
    out = os.path.join(
        os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
        "resources",
        "branding",
        "pulseshield-logo.png",
    )
    im = Image.new("RGBA", (SIZE, SIZE), (0, 0, 0, 0))
    draw = ImageDraw.Draw(im)

    pad = 8
    side = SIZE - 2 * pad
    # Сильно скруглённые углы (форма как у иконок в списке приложений)
    corner_r = max(24, int(side * 0.22))
    draw.rounded_rectangle([pad, pad, SIZE - pad, SIZE - pad], radius=corner_r, fill=FILL)

    # Подбор размера шрифта под форму
    text = "PS"
    font_size = 200
    font = find_font(font_size)
    for _ in range(30):
        bbox = draw.textbbox((0, 0), text, font=font)
        tw = bbox[2] - bbox[0]
        th = bbox[3] - bbox[1]
        if tw < SIZE * 0.62 and th < SIZE * 0.45:
            break
        font_size -= 6
        font = find_font(font_size)

    bbox = draw.textbbox((0, 0), text, font=font)
    tw = bbox[2] - bbox[0]
    th = bbox[3] - bbox[1]
    x = (SIZE - tw) // 2 - bbox[0]
    y = (SIZE - th) // 2 - bbox[1] - 6
    draw.text((x, y), text, font=font, fill=WHITE)

    im.save(out, "PNG")
    print("OK", out)


if __name__ == "__main__":
    main()
    sys.exit(0)
