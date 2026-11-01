package dpi

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CustomDpiListName — базовое имя списка (без .txt), куда попадают домены,
// которым пользователь назначил обход DPI через правила маршрутизации.
const CustomDpiListName = "custom-dpi"

// CustomDpiListRel — путь для строки --hostlist= внутри пресета winws2.
// Путь относительный: winws2 запускается с рабочей папкой zapret.
const CustomDpiListRel = "lists/" + CustomDpiListName + ".txt"

// customDpiHeader — шапка файла. Не начинается с '#!' и не содержит спецсимволов,
// чтобы winws2 читал файл как обычный hostlist.
const customDpiHeader = "# Custom DPI — домены, которым назначен обход DPI.\n" +
	"# Файл управляется программой: правила маршрутизации с действием «обход DPI»\n" +
	"# автоматически добавляют и удаляют здесь свои домены. Можно править вручную.\n"

// CustomDpiListPath возвращает абсолютный путь к custom-dpi.txt.
func CustomDpiListPath() (string, error) {
	zdir, err := ResolveZapretDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zdir, "lists", CustomDpiListName+".txt"), nil
}

// NormalizeHostEntry приводит строку к виду, пригодному для hostlist:
// убирает схему, путь, порт, звёздочку-префикс, приводит к нижнему регистру.
// Возвращает "" для пустой или непригодной строки.
func NormalizeHostEntry(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	if s == "" {
		return ""
	}
	// Отбрасываем комментарии.
	if i := strings.IndexByte(s, '#'); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "//")
	// Только хост, без пути и параметров.
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	// Порт отрезаем, но осторожно с IPv6 в скобках.
	if strings.HasPrefix(s, "[") {
		if end := strings.IndexByte(s, ']'); end > 0 {
			s = s[:end+1]
		}
	} else if i := strings.LastIndexByte(s, ':'); i > 0 && strings.Count(s, ":") == 1 {
		s = s[:i]
	}
	s = strings.Trim(s, " \t.")
	s = strings.TrimPrefix(s, "*.")
	s = strings.TrimPrefix(s, ".")
	if s == "" {
		return ""
	}
	// Минимальная проверка: должен быть хотя бы один символ, кроме разделителей.
	if strings.Trim(s, ".-_") == "" {
		return ""
	}
	return s
}

// ReadCustomDpiEntries читает домены из custom-dpi.txt (без комментариев и пустых строк).
func ReadCustomDpiEntries() ([]string, error) {
	path, err := CustomDpiListPath()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return parseListEntries(string(raw)), nil
}

// parseListEntries разбирает содержимое hostlist-файла.
func parseListEntries(content string) []string {
	out := make([]string, 0, 16)
	seen := make(map[string]struct{}, 16)
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "!") {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// writeCustomDpiEntries перезаписывает custom-dpi.txt: шапка + отсортированные уникальные домены.
func writeCustomDpiEntries(entries []string) error {
	path, err := CustomDpiListPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Уникализация с сохранением нормализованного вида.
	seen := make(map[string]struct{}, len(entries))
	uniq := make([]string, 0, len(entries))
	for _, e := range entries {
		n := NormalizeHostEntry(e)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		uniq = append(uniq, n)
	}
	sort.Strings(uniq)

	var b strings.Builder
	b.WriteString(customDpiHeader)
	for _, e := range uniq {
		b.WriteString(e)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// EnsureCustomDpiList гарантирует существование файла custom-dpi.txt.
// Возвращает true, если файл был создан.
func EnsureCustomDpiList() (bool, error) {
	path, err := CustomDpiListPath()
	if err != nil {
		return false, err
	}
	if _, statErr := os.Stat(path); statErr == nil {
		return false, nil
	}
	if err := writeCustomDpiEntries(nil); err != nil {
		return false, err
	}
	return true, nil
}

// AddCustomDpiEntry добавляет домен в custom-dpi.txt.
// Возвращает true, если домен действительно добавлен (не было раньше).
func AddCustomDpiEntry(raw string) (bool, error) {
	host := NormalizeHostEntry(raw)
	if host == "" {
		return false, fmt.Errorf("пустой или некорректный домен: %q", raw)
	}
	if _, err := EnsureCustomDpiList(); err != nil {
		return false, err
	}
	entries, err := ReadCustomDpiEntries()
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if strings.EqualFold(e, host) {
			return false, nil
		}
	}
	entries = append(entries, host)
	if err := writeCustomDpiEntries(entries); err != nil {
		return false, err
	}
	return true, nil
}

// RemoveCustomDpiEntry удаляет домен из custom-dpi.txt.
// Возвращает true, если домен был и удалён.
func RemoveCustomDpiEntry(raw string) (bool, error) {
	host := NormalizeHostEntry(raw)
	if host == "" {
		return false, nil
	}
	entries, err := ReadCustomDpiEntries()
	if err != nil {
		return false, err
	}
	kept := make([]string, 0, len(entries))
	removed := false
	for _, e := range entries {
		if strings.EqualFold(e, host) {
			removed = true
			continue
		}
		kept = append(kept, e)
	}
	if !removed {
		return false, nil
	}
	return true, writeCustomDpiEntries(kept)
}

// SetCustomDpiEntries полностью заменяет содержимое custom-dpi.txt.
func SetCustomDpiEntries(entries []string) error {
	if _, err := EnsureCustomDpiList(); err != nil {
		return err
	}
	return writeCustomDpiEntries(entries)
}

// InjectCustomDpiHostlist добавляет строку --hostlist=lists/custom-dpi.txt к каждой
// секции пресета, где есть desync-действие и уже есть таргетинг по спискам.
//
// Секции winws2 разделены строкой "--new". Строку вставляем в конец секции,
// чтобы она не разрывала другие аргументы. Если секции с таргетингом нет,
// но есть desync, — пресет не трогаем: иначе обход начнёт бить по всему трафику.
func InjectCustomDpiHostlist(content string) string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	hostArg := "--hostlist=" + CustomDpiListRel

	lines := strings.Split(normalized, "\n")
	var sections [][]string
	var cur []string
	for _, line := range lines {
		cur = append(cur, line)
		if strings.TrimSpace(line) == "--new" {
			sections = append(sections, cur)
			cur = nil
		}
	}
	if len(cur) > 0 {
		sections = append(sections, cur)
	}

	var out []string
	for _, sec := range sections {
		hasTargeting := false
		hasDesync := false
		alreadyHas := false
		for _, line := range sec {
			t := strings.TrimSpace(line)
			low := strings.ToLower(t)
			if strings.HasPrefix(low, "--hostlist=lists/"+CustomDpiListName+".txt") {
				alreadyHas = true
			}
			if lineIsTargetingFilter(t) {
				hasTargeting = true
			}
			if strings.HasPrefix(low, "--lua-desync=") || strings.HasPrefix(low, "--dpi-desync=") {
				hasDesync = true
			}
		}
		if alreadyHas || !hasDesync || !hasTargeting {
			out = append(out, sec...)
			continue
		}
		// Вставляем перед завершающим --new (если он есть), иначе в конец секции.
		insertAt := len(sec)
		if strings.TrimSpace(sec[len(sec)-1]) == "--new" {
			insertAt = len(sec) - 1
		}
		merged := make([]string, 0, len(sec)+1)
		merged = append(merged, sec[:insertAt]...)
		merged = append(merged, hostArg)
		merged = append(merged, sec[insertAt:]...)
		out = append(out, merged...)
	}
	return strings.Join(out, "\n")
}
