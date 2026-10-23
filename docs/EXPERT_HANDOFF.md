# PulseShield / vpnboost — передача эксперту

Этот документ и архив `PulseShield-for-expert.zip` (см. корень репозитория после упаковки) предназначены для внешнего ревью.

## Что это за проект

Настольное приложение **PulseShield** (Windows): локальный прокси на базе **sing-box** (VLESS), обход DPI через **zapret/winws2** (в комплекте `third_party/zapret`), системный прокси, DNS, профили маршрутизации. UI: **Wails v2** + **React** (`ui/PulseShieldDesktop`).

## Стек и точки входа

| Компонент | Путь |
|-----------|------|
| Wails-приложение (Go + биндинги UI) | `ui/PulseShieldDesktop/main.go`, `app.go` |
| Фронтенд | `ui/PulseShieldDesktop/frontend/src/App.tsx`, `App.css` |
| Ядро sing-box / конфиг | `app/core/` |
| DPI / zapret пресеты | `app/dpi/`, `third_party/zapret/`, embed: `app/dpi/embed/` |
| Подписки / VLESS | `app/multiserver/` |
| Модели конфига | `app/models/`, `app/config/` |
| Сборка Windows portable + NSIS | `scripts/build_windows.ps1`, `build.ps1` |
| Альтернативный вход (legacy CLI/tray отдельно) | `cmd/pulseshield/`, `app/main/` |

## Сборка (кратко)

**Чеклист агента:** после задачи, которая меняла десктоп (`ui/PulseShieldDesktop`), сборочные скрипты или код, участвующий в portable, всегда запускай из корня `powershell -ExecutionPolicy Bypass -File .\scripts\build_windows.ps1` и фиксируй результат. Правило продублировано в `.cursor/rules/windows-portable-after-task.mdc`.

- **Требования:** Go 1.21+, Node.js/npm, Wails CLI, для установщика — NSIS (`makensis.exe`).
- Полный frontend: `ui/PulseShieldDesktop/frontend`: `npm install`, `npm run build`.
- Сборка десктопа: из корня `.\scripts\build_windows.ps1` (или с `-Version x.y.z -BuildInstaller`).
- Бинарники ядра: положить `sing-box.exe` в `third_party/core/`, опционально `goodbyedpi.exe` в `third_party/dpi/` (см. README).

## Что исключено из архива для эксперта

Чтобы файл был пересылаемым, из ZIP обычно **не включаются**:

- все каталоги `node_modules`;
- собранный фронт `ui/PulseShieldDesktop/frontend/dist` (восстанавливается `npm run build`);
- локальные билды `ui/PulseShieldDesktop/build/bin`;
- каталог `dist/` с готовыми portable/installer (пересобираются скриптом);
- при наличии — `.git` (эксперт может клонировать репозиторий отдельно).

**Включено:** исходный Go-код, `vendor/` (если есть — для офлайн-сборки), `third_party/zapret` (листы/пресеты/Lua), `configs/`, скрипты, документация.

Крупные проприетарные `.exe` (sing-box, goodbyedpi) в архив могут не попадать — их пути описаны в `README.md` / `third_party/*/README.md`.

## Версия

Текущая линия десктопа: **0.6.0** (задаётся в `scripts/build_windows.ps1`, `build.ps1`, `ui/PulseShieldDesktop/wails.json` — `productVersion`, `ui/PulseShieldDesktop/main.go` — `Version`, ldflags при сборке).

## Контакты по коду

Дополнительно: `README.md`, `RELEASING.md`, `ROUTING_GUIDE.md`, `docs/DEVELOP.md`.
