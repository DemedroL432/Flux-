package dpi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"pulseshield/app/config"
)

func ancestorCandidates(baseDir string, parts ...string) []string {
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

// ZapretRootsNearExe — приоритетные каталоги Zapret относительно .exe (не cwd).
func ZapretRootsNearExe(baseDir string) []string {
	return []string{
		filepath.Join(baseDir, "zapret"),
		filepath.Join(baseDir, "third_party", "zapret"),
		filepath.Join(baseDir, "resources", "zapret"),
	}
}

// ZapretListsHintPath — абсолютный путь к lists для подсказок в UI (ожидаемое расположение portable).
func ZapretListsHintPath() string {
	baseDir, err := config.ExecutableDir()
	if err != nil {
		return "zapret/lists"
	}
	return config.AbsPath(filepath.Join(baseDir, "zapret", "lists"))
}

// ZapretListsPathForHint — абсолютный путь к lists для UI (найденный корень Zapret или ожидаемый portable).
func ZapretListsPathForHint() string {
	if dir, ok := config.ResolveZapretDirNearExe(); ok {
		return config.AbsPath(filepath.Join(dir, "lists"))
	}
	return ZapretListsHintPath()
}

// ResolveZapretWinws2 возвращает абсолютные пути к winws2.exe и корню Zapret.
func ResolveZapretWinws2() (exePath string, zapretDir string, err error) {
	baseDir, err := config.ExecutableDir()
	if err != nil {
		return "", "", err
	}
	type zapretCandidate struct {
		bin string
		dir string
	}
	var candidates []zapretCandidate
	for _, root := range ZapretRootsNearExe(baseDir) {
		candidates = append(candidates,
			zapretCandidate{bin: filepath.Join(root, "exe", "winws2.exe"), dir: root},
			zapretCandidate{bin: filepath.Join(root, "zapret", "exe", "winws2.exe"), dir: filepath.Join(root, "zapret")},
		)
	}
	candidates = append(candidates,
		zapretCandidate{bin: filepath.Join(baseDir, "exe", "winws2.exe"), dir: baseDir},
	)
	for _, d := range ancestorCandidates(baseDir, "third_party", "zapret") {
		candidates = append(candidates,
			zapretCandidate{bin: filepath.Join(d, "exe", "winws2.exe"), dir: d},
			zapretCandidate{bin: filepath.Join(d, "zapret", "exe", "winws2.exe"), dir: filepath.Join(d, "zapret")},
		)
	}
	for _, c := range candidates {
		info, err := os.Stat(c.bin)
		if err == nil && !info.IsDir() {
			return config.AbsPath(c.bin), config.AbsPath(c.dir), nil
		}
	}
	return "", "", errors.New("zapret winws2.exe not found (expected zapret/exe/winws2.exe next to app)")
}

// ResolveZapretDir — корень Zapret рядом с .exe (абсолютный путь); см. config.ResolveZapretDirNearExe.
func ResolveZapretDir() (string, error) {
	if dir, ok := config.ResolveZapretDirNearExe(); ok {
		fmt.Printf("[DPI] Zapret root: %s (lists: %s)\n", dir, filepath.Join(dir, "lists"))
		return dir, nil
	}
	baseDir, err := config.ExecutableDir()
	if err != nil {
		baseDir = "."
	}
	fmt.Printf("[DPI] Zapret folder not found near exe (%s); tried zapret, third_party/zapret, resources/zapret\n", baseDir)
	return "", errors.New("zapret folder not found next to PulseShield executable")
}
