# -*- coding: utf-8 -*-
"""Remove flat padding (#f07575) connected to image edges; keep logo (pink squircle + white PS)."""
from __future__ import annotations

from collections import deque

from PIL import Image


def main() -> None:
    path = r"c:\Users\galae\vpnboost\resources\branding\pulseshield-logo.png"
    im = Image.open(path).convert("RGBA")
    px = im.load()
    w, h = im.size

    def is_padding_pink(r: int, g: int, b: int, tol: int = 28) -> bool:
        return abs(r - 240) < tol and abs(g - 117) < tol and abs(b - 117) < tol

    seen: set[tuple[int, int]] = set()
    q: deque[tuple[int, int]] = deque()
    for x in range(w):
        q.append((x, 0))
        q.append((x, h - 1))
    for y in range(h):
        q.append((0, y))
        q.append((w - 1, y))

    while q:
        x, y = q.popleft()
        if (x, y) in seen:
            continue
        if x < 0 or x >= w or y < 0 or y >= h:
            continue
        seen.add((x, y))
        r, g, b, a = px[x, y]
        if a == 0:
            continue
        if not is_padding_pink(r, g, b):
            continue
        px[x, y] = (0, 0, 0, 0)
        for dx, dy in ((0, 1), (0, -1), (1, 0), (-1, 0)):
            nx, ny = x + dx, y + dy
            if (nx, ny) not in seen:
                q.append((nx, ny))

    im.save(path, "PNG")
    print("OK", path, w, h)


if __name__ == "__main__":
    main()
