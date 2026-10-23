# Пресеты Zapret (`presets/*.txt`)

## `Gamingnew.txt`

Канонический пресет **Gaming** для winws2 (имя в приложении: `zapret:Gamingnew`, путь `presets/Gamingnew.txt` рядом с exe).

При изменении правил обновляйте этот файл в репозитории и пересобирайте portable, чтобы в сборку попали актуальные строки.

Голос Discord (вызовы): для UDP `19294–19344` и `50000–50100` с L7 `stun/discord` используется fake с blob `quic_discord_voice` → файл `bin/quic_initial_dbankcloud_ru.bin` (как в сборках zapret-discord-youtube «ALT3»).

Остальное дерево `zapret/` (`lua/`, `lists/`, `bin/`, `windivert.filter/` и т.д.) должно соответствовать той же базе Zapret, от которой сделан пресет.

## `Gamingnew-Stable.txt`

Запасной пресет на базе Gamingnew: убраны `zapret-multishake.lua` и `custom_diag.lua`, увеличен `ipcache-lifetime` (86400). Часто стабильнее на части провайдеров, если основной пресет ведёт себя непредсказуемо после обновлений.

Имя в приложении: `zapret:Gamingnew-Stable`. Те же правила Discord voice (`quic_discord_voice`), что и у Gamingnew.
