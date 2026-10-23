# GitHub: репозиторий, теги и автообновление

Единая пошаговая инструкция для владельца репозитория и соответствие имён артефактов коду (`pickDownloadURL` / `pickInstallerURL` в `ui/PulseShieldDesktop/github_update.go`) описаны в **[RELEASING.md](../RELEASING.md)**.

Кратко:

- Тег `vX.Y.Z` → сборка с `-Version X.Y.Z` (без префикса `v`).
- В релиз прикладываются ZIP portable, `*Setup*.exe` и `PulseShield-v*-SHA256SUMS.txt` (см. чеклист в `RELEASING.md`).
- При форке передайте `-GitHubOwner` / `-GitHubRepo` при локальной сборке, чтобы кнопка «Проверить обновление» смотрела на ваш репозиторий.
