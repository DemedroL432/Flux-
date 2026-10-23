package config

import (
	"os"
	"path/filepath"
)

func ancestorCandidatesNearExe(baseDir string, parts ...string) []string {
	var out []string
	cur := baseDir
	for i := 0; i < 7; i++ {
		out = append(out, filepath.Join(append([]string{cur}, parts...)...))
		next := filepath.Dir(cur)
		if next == cur {
			break
		}
		cur = next
	}
	return out
}

func zapretHasListsDir(root string) bool {
	st, err := os.Stat(filepath.Join(root, "lists"))
	return err == nil && st.IsDir()
}

type zapretCand struct {
	bin string
	dir string
}

// ResolveZapretDirNearExe находит корень Zapret рядом с исполняемым файлом (для lists/).
// Возвращает абсолютный путь; при нескольких кандидатах предпочитает каталог, где есть подпапка lists/.
func ResolveZapretDirNearExe() (dir string, ok bool) {
	baseDir, err := ExecutableDir()
	if err != nil {
		return "", false
	}

	var dirs []string
	seen := map[string]bool{}
	pushDir := func(root string) {
		root = filepath.Clean(root)
		if root == "" || seen[root] {
			return
		}
		if st, err := os.Stat(root); err != nil || !st.IsDir() {
			return
		}
		seen[root] = true
		dirs = append(dirs, root)
	}

	for _, root := range []string{
		filepath.Join(baseDir, "zapret"),
		filepath.Join(baseDir, "third_party", "zapret"),
		filepath.Join(baseDir, "resources", "zapret"),
	} {
		pushDir(root)
	}

	var candidates []zapretCand
	candidates = append(candidates,
		zapretCand{bin: filepath.Join(baseDir, "zapret", "exe", "winws2.exe"), dir: filepath.Join(baseDir, "zapret")},
		zapretCand{bin: filepath.Join(baseDir, "zapret", "zapret", "exe", "winws2.exe"), dir: filepath.Join(baseDir, "zapret", "zapret")},
		zapretCand{bin: filepath.Join(baseDir, "exe", "winws2.exe"), dir: baseDir},
	)
	for _, d := range ancestorCandidatesNearExe(baseDir, "third_party", "zapret") {
		candidates = append(candidates,
			zapretCand{bin: filepath.Join(d, "exe", "winws2.exe"), dir: d},
			zapretCand{bin: filepath.Join(d, "zapret", "exe", "winws2.exe"), dir: filepath.Join(d, "zapret")},
		)
	}
	for _, c := range candidates {
		if st, err := os.Stat(c.bin); err == nil && !st.IsDir() {
			pushDir(c.dir)
		}
	}

	if len(dirs) == 0 {
		return "", false
	}
	for _, root := range dirs {
		if zapretHasListsDir(root) {
			return AbsPath(root), true
		}
	}
	return AbsPath(dirs[0]), true
}
