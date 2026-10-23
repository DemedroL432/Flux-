# Разработка PulseShield (кратко)

## Зависимости

- **Go** 1.23+ (как в CI).
- **Node.js** 20+ для фронтенда Wails.
- **Wails v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest` (CLI обычно в `%USERPROFILE%\go\bin`).
- **NSIS** — только если собираете установщик локально; в CI ставится через Chocolatey.

## Расположение UI

Проект Wails: `ui/PulseShieldDesktop/`. После изменений Go-методов, экспортируемых во фронт:

```powershell
cd ui\PulseShieldDesktop
wails generate module
```

## Сборка

Из корня репозитория:

```powershell
.\build.ps1 -Version 0.3.0
```

С установщиком (нужен `makensis.exe` в PATH или в `Program Files (x86)\NSIS`):

```powershell
.\build.ps1 -Version 0.3.0 -BuildInstaller
```

Готовая portable-папка: `dist\pulseshield-win64\` (или с суффиксом даты, если предыдущая папка была занята). Установщик при `-BuildInstaller`: `dist\PulseShield-v<версия>-Setup.exe` и файл checksums `PulseShield-v<версия>-SHA256SUMS.txt`.

## Полезные пути

- Скрипт сборки: `scripts/build_windows.ps1`
- NSIS: `scripts/installer/pulseshield.nsi`
- Релизы и чеклист: `RELEASING.md`

## Модуль Go и данные

- Импорты: `pulseshield/app/...`.
- Данные пользователя: `%LocalAppData%\PulseShield\` (при обновлении со старого клиента возможна миграция из `vpnboost`).
