package dpi

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// maxListFileBytes — предел размера файла списка, чтобы случайный выбор не
// уронил UI на гигабайтном файле.
const maxListFileBytes = 1 << 20 // 1 МБ

// maxListEntries — предел числа записей при сохранении из редактора.
const maxListEntries = 50000

// DpiListContent — содержимое одного списка для редактора в UI.
type DpiListContent struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Entries     []string `json:"entries"`
	Total       int      `json:"total"`
	Customized  bool     `json:"customized"` // true для нашего управляемого списка
	TooLarge    bool     `json:"tooLarge"`
	Description string   `json:"description"`
}

// validListName отсекает имена с разделителями и обходом каталога.
// Список — это только базовое имя файла в zapret/lists без .txt.
func validListName(name string) (string, error) {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, ".txt")
	if name == "" {
		return "", fmt.Errorf("пустое имя списка")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return "", fmt.Errorf("недопустимое имя списка: %q", name)
	}
	if strings.ContainsAny(name, `:*?"<>|`) {
		return "", fmt.Errorf("недопустимые символы в имени списка: %q", name)
	}
	if len(name) > 128 {
		return "", fmt.Errorf("слишком длинное имя списка")
	}
	return name, nil
}

// listFilePath возвращает путь к файлу списка по его базовому имени.
func listFilePath(name string) (string, string, error) {
	base, err := validListName(name)
	if err != nil {
		return "", "", err
	}
	zdir, err := ResolveZapretDir()
	if err != nil {
		return "", "", fmt.Errorf("zapret: %w", err)
	}
	return filepath.Join(zdir, "lists", base+".txt"), base, nil
}

// ReadDpiListContent читает записи списка для показа в редакторе.
func ReadDpiListContent(name string) (*DpiListContent, error) {
	path, base, err := listFilePath(name)
	if err != nil {
		return nil, err
	}
	out := &DpiListContent{
		Name:        base,
		Path:        path,
		Customized:  base == CustomDpiListName,
		Description: describeDpiListFile(base),
	}
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	if st.Size() > maxListFileBytes {
		out.TooLarge = true
		return out, fmt.Errorf("файл списка слишком большой (%d Б) — редактируйте вручную: %s", st.Size(), path)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out.Entries = parseListEntries(string(raw))
	out.Total = len(out.Entries)
	return out, nil
}

// WriteDpiListContent полностью заменяет содержимое списка.
//
// Для управляемого custom-dpi.txt сохраняем шапку, чтобы назначение файла
// оставалось понятным при открытии вручную.
func WriteDpiListContent(name string, entries []string) error {
	path, base, err := listFilePath(name)
	if err != nil {
		return err
	}
	if len(entries) > maxListEntries {
		return fmt.Errorf("слишком много записей: %d (максимум %d)", len(entries), maxListEntries)
	}

	// Нормализуем и убираем дубликаты, сохраняя порядок ввода.
	seen := make(map[string]struct{}, len(entries))
	clean := make([]string, 0, len(entries))
	for _, e := range entries {
		// Каждая запись может быть многострочной (пользователь вставил блоком).
		for _, line := range strings.Split(strings.ReplaceAll(e, "\r\n", "\n"), "\n") {
			n := NormalizeHostEntry(line)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			clean = append(clean, n)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	var b strings.Builder
	if base == CustomDpiListName {
		b.WriteString(customDpiHeader)
	}
	for _, e := range clean {
		b.WriteString(e)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return err
	}
	return nil
}

// AddDpiListEntry добавляет одну запись в список. Возвращает true, если добавлена.
func AddDpiListEntry(name, entry string) (bool, error) {
	content, err := ReadDpiListContent(name)
	if err != nil {
		return false, err
	}
	host := NormalizeHostEntry(entry)
	if host == "" {
		return false, fmt.Errorf("пустая или некорректная запись: %q", entry)
	}
	for _, e := range content.Entries {
		if strings.EqualFold(e, host) {
			return false, nil
		}
	}
	content.Entries = append(content.Entries, host)
	return true, WriteDpiListContent(name, content.Entries)
}

// RemoveDpiListEntry удаляет запись из списка. Возвращает true, если была и удалена.
func RemoveDpiListEntry(name, entry string) (bool, error) {
	content, err := ReadDpiListContent(name)
	if err != nil {
		return false, err
	}
	host := NormalizeHostEntry(entry)
	kept := make([]string, 0, len(content.Entries))
	removed := false
	for _, e := range content.Entries {
		if strings.EqualFold(NormalizeHostEntry(e), host) {
			removed = true
			continue
		}
		kept = append(kept, e)
	}
	if !removed {
		return false, nil
	}
	return true, WriteDpiListContent(name, kept)
}
