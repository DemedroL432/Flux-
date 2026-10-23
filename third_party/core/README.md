Place bundled core binaries here for release packaging.

Expected files for current build script:

- `third_party/core/sing-box.exe`
- **`third_party/core/wintun.dll`** (amd64) — нужен для режима **TUN** (виртуальный адаптер `pulse-tun`). Возьмите архив с [wintun.net](https://www.wintun.net), распакуйте `wintun\bin\amd64\wintun.dll` сюда. Без DLL ядро выдаст ошибку создания TUN.

Then run:

- `powershell -ExecutionPolicy Bypass -File .\scripts\build_windows.ps1`

The resulting `dist/pulseshield-win64` folder is portable and contains:

- `PulseShield.exe` (GUI app)
- `core/sing-box.exe` (bundled runtime)
- `configs/*`

End users only need this folder and do not need to install Go/sing-box manually.
