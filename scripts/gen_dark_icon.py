"""PNG 1024×1024 для Wails appicon: прозрачный фон, крупный тёмный сквиркл, жирные буквы PS."""
from __future__ import annotations

import struct
import zlib
from pathlib import Path


def _png_chunk(tag: bytes, data: bytes) -> bytes:
    return struct.pack(">I", len(data)) + tag + data + struct.pack(">I", zlib.crc32(tag + data) & 0xFFFFFFFF)


def write_png_rgba(path: str, w: int, h: int, rgba: bytes) -> None:
    assert len(rgba) == w * h * 4
    raw = b"".join(b"\x00" + rgba[y * w * 4 : (y + 1) * w * 4] for y in range(h))
    comp = zlib.compress(raw, 9)
    ihdr = struct.pack(">IIBBBBB", w, h, 8, 6, 0, 0, 0)
    png = b"\x89PNG\r\n\x1a\n" + _png_chunk(b"IHDR", ihdr) + _png_chunk(b"IDAT", comp) + _png_chunk(b"IEND", b"")
    Path(path).parent.mkdir(parents=True, exist_ok=True)
    with open(path, "wb") as f:
        f.write(png)


def inside_squircle(nx: float, ny: float, n: float = 4.0) -> float:
    ax, ay = abs(nx), abs(ny)
    v = (ax**n + ay**n) ** (1 / n)
    if v <= 0.92:
        return 1.0
    if v >= 1.08:
        return 0.0
    return max(0.0, min(1.0, (1.08 - v) / 0.16))


def main() -> None:
    size = 1024
    cx = cy = size / 2
    rgba = bytearray(size * size * 4)
    for i in range(size * size):
        rgba[i * 4 + 3] = 0

    # Крупнее «плитка» иконки в кадре — лучше читается в .exe
    rx = size * 0.46
    for y in range(size):
        for x in range(size):
            nx = (x - cx) / rx
            ny = (y - cy) / rx
            if abs(nx) > 1.15 or abs(ny) > 1.15:
                continue
            a = inside_squircle(nx, ny, 4.0)
            if a <= 0:
                continue
            i = (y * size + x) * 4
            g = 0.82 + 0.18 * (1 - (ny + 1) / 2)
            r, g0, b = int(16 * g), int(18 * g), int(24 * g)
            rgba[i : i + 4] = (r, g0, b, int(255 * a))

    def put_block(x0: int, y0: int, pat: list[str], scale: int, rgb: tuple[int, int, int]) -> None:
        pts: list[tuple[int, int]] = []
        for ry, row in enumerate(pat):
            for rx, ch in enumerate(row):
                if ch != "#":
                    continue
                for sy in range(scale):
                    for sx in range(scale):
                        x, y = x0 + rx * scale + sx, y0 + ry * scale + sy
                        if 0 <= x < size and 0 <= y < size:
                            pts.append((x, y))
        # жирность: сдвиги соседних пикселей
        bold: set[tuple[int, int]] = set()
        for x, y in pts:
            for dx in (-1, 0, 1):
                for dy in (-1, 0, 1):
                    nx, ny = x + dx, y + dy
                    if 0 <= nx < size and 0 <= ny < size:
                        bold.add((nx, ny))
        for x, y in bold:
            j = (y * size + x) * 4
            if rgba[j + 3] > 0:
                rgba[j : j + 3] = rgb
                rgba[j + 3] = 255

    p_pat = [
        "####",
        "#..#",
        "####",
        "#...",
        "#...",
        "#...",
        "#...",
    ]
    s_pat = [
        ".####",
        "#....",
        "#....",
        ".###.",
        "....#",
        "....#",
        "####.",
    ]
    # Крупные «PS», по центру сквиркла; зазор ~1.15× scale для читаемости в 16–32px ICO
    scale = 42
    gap = max(44, int(scale * 1.12))
    p_cols, s_cols = 4, 5
    total_w = p_cols * scale + gap + s_cols * scale
    px = int(cx - total_w / 2)
    py = int(cy - (7 * scale) / 2)
    put_block(px, py, p_pat, scale, (252, 253, 255))
    put_block(px + p_cols * scale + gap, py, s_pat, scale, (252, 253, 255))

    out = Path(__file__).resolve().parent.parent / "resources" / "branding" / "pulseshield-logo.png"
    write_png_rgba(str(out), size, size, bytes(rgba))
    print(out)


if __name__ == "__main__":
    main()
