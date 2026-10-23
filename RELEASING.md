# Публикация PulseShield на GitHub

Краткий указатель для навигации из репозитория: [docs/GITHUB_RELEASES.md](docs/GITHUB_RELEASES.md).

## Иконка и брендинг (Windows)

- Канонический логотип: [`resources/branding/pulseshield-logo.png`](resources/branding/pulseshield-logo.png) — **строго квадрат** (например 512×512), иначе иконка exe в проводнике может выглядеть сжатой. Перед `wails build` скрипт [`scripts/build_windows.ps1`](scripts/build_windows.ps1) копирует его в `ui/PulseShieldDesktop/build/appicon.png` (встраивается в трей и даёт материал для иконки exe).
- Wails при сборке формирует `ui/PulseShieldDesktop/build/windows/icon.ico` из `appicon.png`. Скрипт `build_windows.ps1` перед `wails build` удаляет старый `icon.ico`, иначе в exe может остаться бирюзовая иконка по умолчанию. Не добавляйте в `wails.json` шаги, которые принудительно удаляют `icon.ico` **после** копирования брендинга.

## Репозиторий и проверка обновлений

1. Создайте репозиторий на GitHub (например `ваш-логин/pulseshield`).
2. При сборке в бинарник зашиваются **владелец** и **имя** репозитория (для API `releases/latest`):
   - по умолчанию в коде: `DemedroL` / `pulseshield`;
   - переопределение:  
     `.\scripts\build_windows.ps1 -GitHubOwner ВАШ_ЛОГИН -GitHubRepo ИМЯ_РЕПО`
3. В приложении кнопка **«Проверить обновление»** запрашивает  
   `https://api.github.com/repos/<owner>/<repo>/releases/latest`, сравнивает тег с версией exe и показывает:
   - ссылку на страницу релиза;
   - основное вложение для скачивания (обычно ZIP portable, если он есть в релизе);
   - отдельно URL установщика NSIS, если в релизе есть `*Setup*.exe`.

## Автообновление из приложения

- **Portable ZIP** (`PulseShield-v*-win64.zip`): скачивание → выход → копирование файлов в каталог текущего `PulseShield.exe` → перезапуск. В оверлее показывается прогресс скачивания (проценты, если сервер отдал `Content-Length`).
- **NSIS Setup** (`PulseShield-v*-Setup.exe`): скачивание → выход → запуск установщика с флагом `/S`. Нужны права на каталог установки (как при первой установке в Program Files).

Имя установщика в релизе должно содержать `setup` или `installer` и иметь расширение `.exe`.

## Версия сборки

- Версия задаётся при сборке: `-Version` в `build_windows.ps1` / корневом `build.ps1` и `-ldflags "-X main.Version=..."` для Wails.
- **Тег `v*` и `-Version` должны совпадать по числу:** тег `v0.2.3` → сборка с `-Version 0.2.3` (без префикса `v` в параметре). Иначе в приложении и на GitHub отобразятся разные версии.

## Чеклист перед тегом `v*`

Перед `git tag` и `git push origin v0.2.x`:

1. **sing-box в CI:** шаг «Fetch sing-box» в `.github/workflows/release.yml` должен успешно находить архив `windows-amd64` в последнем релизе sing-box (если upstream переименует артефакты — поправить фильтр в workflow).
2. **NSIS на раннере:** в workflow NSIS ставится через Chocolatey и путь добавляется в `PATH`; локально нужен `makensis.exe` для ручной сборки установщика.
3. **Артефакты в релизе** (как прикрепляет Actions):
   - ZIP: `PulseShield-<тег>-win64.zip` (например `PulseShield-v0.2.0-win64.zip`) — в имени желательно есть `pulseshield` и `win64`, чтобы логика выбора в `pickDownloadURL` в `ui/PulseShieldDesktop/github_update.go` отдала portable как основное вложение.
   - Установщик: `PulseShield-<тег>-Setup.exe` (например `PulseShield-v0.2.0-Setup.exe`) — в имени должны быть `setup` или `installer` и расширение `.exe`, см. `pickInstallerURL` в том же файле.
   - SHA256: `PulseShield-<тег>-SHA256SUMS.txt` — хеши ZIP и Setup для ручной проверки.
4. **Владелец/репо в бинарнике** совпадают с реальным GitHub-репозиторием (`-GitHubOwner` / `-GitHubRepo` в сборке; в CI подставляются из `github.repository`).

## Автоматический релиз (GitHub Actions)

При **push тега** вида `v*` (например `v0.2.0`) workflow `.github/workflows/release.yml`:

1. Скачивает последний **sing-box** для Windows amd64 в `third_party/core/`.
2. Запускает `scripts/build_windows.ps1` с `-GitHubOwner` / `-GitHubRepo` из текущего репозитория.
3. Упаковывает portable-папку в `PulseShield-<тег>-win64.zip`.
4. Собирает **NSIS**-установщик `PulseShield-<тег>-Setup.exe` (нужен NSIS в CI — ставится через Chocolatey).
5. Создаёт/обновляет **GitHub Release** и прикрепляет ZIP, Setup.exe и `PulseShield-<тег>-SHA256SUMS.txt`.

Локально перед первым релизом:

```bash
git tag v0.2.0
git push origin v0.2.0
```

## Ручная сборка, ZIP и установщик

1. Положите `sing-box.exe` в `third_party/core/` (и при необходимости `goodbyedpi.exe` в `third_party/dpi/`).
2. Только portable (корневая обёртка или скрипт в `scripts/`):

```powershell
.\build.ps1 -Version 0.2.0 -GitHubOwner ВАШ_ЛОГИН -GitHubRepo ИМЯ_РЕПО
# эквивалентно:
.\scripts\build_windows.ps1 -Version 0.2.0 -GitHubOwner ВАШ_ЛОГИН -GitHubRepo ИМЯ_РЕПО
```

3. Portable плюс **NSIS** (нужен [NSIS](https://nsis.sourceforge.io/), `makensis.exe` в `PATH` или в `Program Files (x86)\NSIS`):

```powershell
.\build.ps1 -Version 0.2.0 -GitHubOwner ВАШ_ЛОГИН -GitHubRepo ИМЯ_РЕПО -BuildInstaller
```

В `dist\` появится `PulseShield-v<версия>-Setup.exe` (как git-тег, например `PulseShield-v0.2.0-Setup.exe`) и `PulseShield-v<версия>-SHA256SUMS.txt` с SHA256 для ZIP и Setup, если оба артефакта есть в `dist`.

Имя ZIP с подстроками `pulseshield`, `win64` и расширение `.zip` повышает шанс, что API выберет правильный файл как основное вложение.

## SHA256 и подпись (Authenticode)

- Контрольные суммы: [`scripts/write_release_checksums.ps1`](scripts/write_release_checksums.ps1). В GitHub Actions файл `PulseShield-v*-SHA256SUMS.txt` прикладывается к релизу вместе с ZIP и Setup.
- Подпись в CI по умолчанию не выполняется (нужен сертификат). Локально: [`scripts/sign_windows_binaries.ps1`](scripts/sign_windows_binaries.ps1) и `signtool` из Windows SDK.
- Архив старого **Inno Setup** (откат при разработке): [`installer/legacy/inno-setup-v0.1.4-archive/`](installer/legacy/inno-setup-v0.1.4-archive/README.md).

## Если релизов ещё нет

API `releases/latest` вернёт 404 — в приложении отобразится соответствующее сообщение. Создайте первый релиз с тегом (через Actions или вручную).
