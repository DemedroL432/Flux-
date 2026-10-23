# Пользовательские профили маршрутизации

## Как работает

PulseShield при запуске загружает `*.yaml` файлы из папки `routing/` рядом с `PulseShieldDesktop.exe`.

Файлы из этой папки (`configs/routing/`) — это **шаблоны**. Чтобы применить:

1. Скопируйте нужный файл в `routing/` рядом с exe (создайте папку, если нет)
2. Запустите или перезапустите PulseShield
3. Профиль появится в интерфейсе на вкладке «Сеть»

## Приоритеты

- YAML-файл с именем встроенного профиля (`default`, `selective-proxy`, `gaming-friendly`, `rf-smart-split`) **заменяет** встроенный
- YAML-файл с новым именем **добавляется** как дополнительный профиль

## Формат

```yaml
name: мой-профиль       # Уникальное имя (латиница, дефис)
final: direct           # Финальный outbound: "direct" или "proxy" (опционально)
rules:
  - name: локальные
    domains:
      - "*.local"
      - "localhost"
    action: direct

  - name: мессенджеры
    domainSuffixes:
      - "telegram.org"
      - "t.me"
    action: proxy

  - name: только-мой-сайт
    domains:
      - "mysite.example.com"
    action: proxy
```

## Поля правила

| Поле            | Описание                                      |
|-----------------|-----------------------------------------------|
| `name`          | Метка правила (для диагностики)               |
| `domains`       | Точные домены или `*.wildcard`                |
| `domainSuffixes`| Суффиксы: `example.com` → *.example.com       |
| `cidrs`         | IP-диапазоны: `10.0.0.0/8`                   |
| `processes`     | Имена процессов: `Telegram.exe`              |
| `action`        | `proxy` / `direct` / `dpi-bypass` / `block` |

## Действия (action)

| Значение    | Описание                                        |
|-------------|--------------------------------------------------|
| `proxy`     | Трафик через VLESS-сервер                       |
| `direct`    | Прямое подключение (без прокси и DPI)           |
| `dpi-bypass`| Прямое подключение через DPI (zapret/winws2)    |
| `block`     | Соединение отклоняется                          |

## Примеры в этой папке

- `default.yaml` — весь трафик через VLESS (профиль по умолчанию)
- `selective-proxy.yaml` — только нужные сайты в VPN
- `gaming-friendly.yaml` — игры напрямую, медиа через VPN
- `rf-smart-split.yaml` — оптимизировано для РФ с DPI-обходом
