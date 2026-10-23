package dpi

import (
	"path/filepath"
	"strings"

	"pulseshield/app/models"
)

// listBaseFromListArg возвращает имя списка без .txt для строк
// --ipset=lists/foo.txt и --hostlist=lists/foo.txt (а также их exclude-вариантов мы НЕ трогаем).
func listBaseFromListArg(line string) (base string, ok bool) {
	line = strings.TrimSpace(line)
	low := strings.ToLower(line)
	var prefix string
	switch {
	case strings.HasPrefix(low, "--ipset="):
		prefix = "--ipset="
	case strings.HasPrefix(low, "--hostlist="):
		prefix = "--hostlist="
	default:
		return "", false
	}
	arg := strings.TrimSpace(line[len(prefix):])
	arg = strings.Trim(arg, `"`)
	arg = strings.TrimPrefix(arg, "lists/")
	arg = strings.TrimPrefix(arg, `lists\`)
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return "", false
	}
	base = strings.TrimSuffix(filepath.Base(arg), ".txt")
	return base, base != ""
}

// lineIsTargetingFilter — строка сужает действие секции до конкретных хостов/IP
// (--ipset, --hostlist, --hostlist-domains, --ipset-ip). Если в секции есть desync,
// но НЕТ ни одной такой строки, winws2 применит desync ко всему трафику на портах из
// --filter-* — это и ломает Steam/osu/обычные сайты. Поэтому пустые секции выкидываем.
func lineIsTargetingFilter(line string) bool {
	low := strings.ToLower(strings.TrimSpace(line))
	for _, p := range []string{"--ipset=", "--hostlist=", "--hostlist-domains=", "--ipset-ip="} {
		if strings.HasPrefix(low, p) {
			return true
		}
	}
	return false
}

// FilterPresetLinesByEnabledLists удаляет строки --ipset=/--hostlist=lists/X.txt для
// выключенных в конфиге списков. Если после удаления секция (блок между --new) лишилась
// ВСЕХ таргетинг-фильтров, хотя изначально они были, — вся секция выкидывается, чтобы
// не получить desync, бьющий по всему трафику. Nil/пустая карта — фильтрация отключена.
func FilterPresetLinesByEnabledLists(content string, cfg *models.AppConfig) string {
	if cfg == nil || cfg.DpiListEnabled == nil || len(cfg.DpiListEnabled) == 0 {
		return content
	}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	// Разбиваем на секции по строкам "--new" (сама строка остаётся как разделитель).
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
		filtered, dropWhole := filterSection(sec, cfg)
		if dropWhole {
			// Сохраняем завершающий "--new" этой секции (если был), чтобы не склеить соседей.
			if last := strings.TrimSpace(sec[len(sec)-1]); last == "--new" {
				out = append(out, "--new")
			}
			continue
		}
		out = append(out, filtered...)
	}
	return strings.Join(out, "\n")
}

// filterSection обрабатывает одну секцию пресета. Возвращает отфильтрованные строки
// и флаг «выкинуть всю секцию» (если она осиротела — потеряла все таргетинг-фильтры).
func filterSection(sec []string, cfg *models.AppConfig) (result []string, dropWhole bool) {
	targetingBefore := 0
	for _, line := range sec {
		if lineIsTargetingFilter(line) {
			targetingBefore++
		}
	}

	var kept []string
	for _, line := range sec {
		if base, ok := listBaseFromListArg(line); ok {
			if !models.DpiListIsEnabled(cfg, base) {
				continue // выключенный список — удаляем строку
			}
		}
		kept = append(kept, line)
	}

	targetingAfter := 0
	hasDesync := false
	for _, line := range kept {
		if lineIsTargetingFilter(line) {
			targetingAfter++
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "--lua-desync=") ||
			strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "--dpi-desync=") {
			hasDesync = true
		}
	}

	// Секция изначально таргетировалась по спискам, но теперь все они выключены, а desync
	// остался → она применит обход ко всему трафику. Выкидываем секцию целиком.
	if targetingBefore > 0 && targetingAfter == 0 && hasDesync {
		return nil, true
	}
	return kept, false
}
