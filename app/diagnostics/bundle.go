package diagnostics

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"time"
)

type BundleResult struct {
	ZipPath string
}

func CreateBundle() (BundleResult, error) {
	rep, err := Run()
	if err != nil {
		return BundleResult{}, err
	}

	outName := "pulseshield_diagnostics_" + time.Now().Format("20060102_150405") + ".zip"
	outPath := filepath.Join(rep.DataDir, outName)
	f, err := os.Create(outPath)
	if err != nil {
		return BundleResult{}, err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	_ = addFile(zw, rep.ConfigPath, "app-config.json")

	runtimeDir := filepath.Join(rep.DataDir, "runtime")
	_ = addDir(zw, runtimeDir, "runtime")

	_ = addDir(zw, rep.LogsDir, "logs")

	return BundleResult{ZipPath: outPath}, nil
}

func addDir(zw *zip.Writer, dir string, prefix string) error {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		return addFile(zw, path, filepath.ToSlash(filepath.Join(prefix, rel)))
	})
}

func addFile(zw *zip.Writer, srcPath string, zipName string) error {
	info, err := os.Stat(srcPath)
	if err != nil || info.IsDir() {
		return nil
	}
	r, err := os.Open(srcPath)
	if err != nil {
		return nil
	}
	defer r.Close()

	w, err := zw.Create(zipName)
	if err != nil {
		return nil
	}
	_, _ = io.Copy(w, r)
	return nil
}

