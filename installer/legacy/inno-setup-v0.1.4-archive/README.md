# Архив: установщик Inno Setup (до перехода на NSIS)

Здесь сохранён сценарий **Inno Setup** для PulseShield, который использовался на этапе разработки до основного конвейера **NSIS** (`scripts/installer/pulseshield.nsi`).

- **Версия сценария в архиве:** `AppVersion=0.1.4` (см. `PulseShield-InnoSetup.iss`).
- **Текущий способ сборки установщика:** `.\build.ps1 -Version x.y.z -BuildInstaller` → `dist/PulseShield-vx.y.z-Setup.exe` (NSIS).

Откат на этапе разработки: при необходимости собрать старый Inno-инсталлятор установите [Inno Setup](https://jrsoftware.org/isinfo.php), соберите portable (`scripts/build_windows.ps1`), затем откройте `PulseShield-InnoSetup.iss` в Inno и скомпилируйте — пути в скрипте рассчитаны на корень репозитория (`dist\pulseshield-win64\...`).
