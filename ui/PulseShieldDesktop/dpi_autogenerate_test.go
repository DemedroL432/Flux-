package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestZapretProfileNameAutoGenerate(t *testing.T) {
	got := zapretProfileNameFromConfigPath("presets/AutoGenerate.txt")
	want := "zapret:AutoGenerate"
	if got != want {
		t.Fatalf("zapretProfileNameFromConfigPath = %q, want %q", got, want)
	}
}

func TestIsLegacyAutoPresetFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{name: "20261101213858.txt", want: true},
		{name: "_pulse_tune_001_alt11.txt", want: true},
		{name: "_PULSE_TUNE_000_HYBRID.TXT", want: true},
		{name: "AutoGenerate.txt", want: false},
		{name: "custom.txt", want: false},
		{name: "20261101.txt", want: false},
		{name: "2026110121385a.txt", want: false},
		{name: "20261101213858.log", want: false},
	}
	for _, tc := range cases {
		if got := isLegacyAutoPresetFile(tc.name); got != tc.want {
			t.Errorf("isLegacyAutoPresetFile(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestWriteAutoGeneratePresetOverwritesAndCleansLegacy(t *testing.T) {
	dir := t.TempDir()
	autoDir := filepath.Join(dir, "presets", "auto")
	if err := os.MkdirAll(autoDir, 0o755); err != nil {
		t.Fatalf("mkdir presets/auto: %v", err)
	}

	legacyFiles := []string{
		"20261101213858.txt",
		"20261101221009.txt",
		"_pulse_tune_000_test.txt",
	}
	for _, name := range legacyFiles {
		if err := os.WriteFile(filepath.Join(autoDir, name), []byte("old"), 0o644); err != nil {
			t.Fatalf("write legacy %s: %v", name, err)
		}
	}
	keepPath := filepath.Join(autoDir, "keep-me.txt")
	if err := os.WriteFile(keepPath, []byte("custom"), 0o644); err != nil {
		t.Fatalf("write keep-me.txt: %v", err)
	}

	rel, pn, err := writeAutoGeneratePreset(dir, []byte("first"))
	if err != nil {
		t.Fatalf("first writeAutoGeneratePreset: %v", err)
	}
	if rel != "presets/AutoGenerate.txt" {
		t.Fatalf("rel = %q, want presets/AutoGenerate.txt", rel)
	}
	if pn != "zapret:AutoGenerate" {
		t.Fatalf("pn = %q, want zapret:AutoGenerate", pn)
	}

	targetPath := filepath.Join(dir, "presets", dpiAutoGenerateFile)
	gotFirst, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read AutoGenerate.txt: %v", err)
	}
	if string(gotFirst) != "first" {
		t.Fatalf("content = %q, want first", string(gotFirst))
	}

	for _, name := range legacyFiles {
		if _, err := os.Stat(filepath.Join(autoDir, name)); !os.IsNotExist(err) {
			t.Errorf("legacy file %s still exists after cleanup", name)
		}
	}
	if _, err := os.Stat(keepPath); err != nil {
		t.Errorf("custom file keep-me.txt should not be deleted: %v", err)
	}

	if _, _, err := writeAutoGeneratePreset(dir, []byte("second")); err != nil {
		t.Fatalf("second writeAutoGeneratePreset: %v", err)
	}
	gotSecond, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read AutoGenerate.txt after overwrite: %v", err)
	}
	if string(gotSecond) != "second" {
		t.Fatalf("content after overwrite = %q, want second", string(gotSecond))
	}

	presetEntries, err := os.ReadDir(filepath.Join(dir, "presets"))
	if err != nil {
		t.Fatalf("readdir presets: %v", err)
	}
	var txtCount int
	for _, e := range presetEntries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".txt" {
			txtCount++
		}
	}
	if txtCount != 1 {
		t.Fatalf("expected 1 .txt file in presets, got %d", txtCount)
	}
}

func TestWriteAutoGeneratePresetRemovesEmptyLegacyDir(t *testing.T) {
	dir := t.TempDir()
	autoDir := filepath.Join(dir, "presets", "auto")
	if err := os.MkdirAll(autoDir, 0o755); err != nil {
		t.Fatalf("mkdir presets/auto: %v", err)
	}
	if err := os.WriteFile(filepath.Join(autoDir, "20261101213858.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	if _, _, err := writeAutoGeneratePreset(dir, []byte("preset")); err != nil {
		t.Fatalf("writeAutoGeneratePreset: %v", err)
	}
	if _, err := os.Stat(autoDir); !os.IsNotExist(err) {
		t.Fatalf("expected empty presets/auto dir to be removed, stat err = %v", err)
	}
}
