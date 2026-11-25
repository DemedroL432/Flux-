package main

import (
	"testing"

	"pulseshield/app/config"
	"pulseshield/app/state"
)

func TestIsInstalledProgramFilesPath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\Program Files\Flux\Flux.exe`, true},
		{`C:\Program Files (x86)\Flux\Flux.exe`, true},
		{`D:\Portable\Flux-v0.7.7-win64\Flux.exe`, false},
		{`C:\Users\galae\Flux\dist\Flux-v0.7.7-win64\Flux.exe`, false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isInstalledProgramFilesPath(tc.path); got != tc.want {
			t.Errorf("isInstalledProgramFilesPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestPickAutoUpdateURL(t *testing.T) {
	zipURL := "https://github.com/DemedroL432/Flux-/releases/download/v0.7.7/Flux-v0.7.7-win64.zip"
	instURL := "https://github.com/DemedroL432/Flux-/releases/download/v0.7.7/Flux-v0.7.7-Setup.exe"

	cases := []struct {
		name    string
		info    UpdateInfo
		exePath string
		want    string
	}{
		{
			name:    "installed prefers installer",
			info:    UpdateInfo{DownloadURL: zipURL, InstallerDownloadURL: instURL},
			exePath: `C:\Program Files\Flux\Flux.exe`,
			want:    instURL,
		},
		{
			name:    "installed falls back to zip",
			info:    UpdateInfo{DownloadURL: zipURL},
			exePath: `C:\Program Files\Flux\Flux.exe`,
			want:    zipURL,
		},
		{
			name:    "portable prefers zip",
			info:    UpdateInfo{DownloadURL: zipURL, InstallerDownloadURL: instURL},
			exePath: `D:\Portable\Flux-v0.7.7-win64\Flux.exe`,
			want:    zipURL,
		},
		{
			name:    "portable falls back to installer",
			info:    UpdateInfo{InstallerDownloadURL: instURL},
			exePath: `D:\Portable\Flux-v0.7.7-win64\Flux.exe`,
			want:    instURL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := pickAutoUpdateURL(tc.info, tc.exePath); got != tc.want {
				t.Fatalf("pickAutoUpdateURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRemoteIsNewer(t *testing.T) {
	cases := []struct {
		current string
		remote  string
		want    bool
	}{
		{"0.7.6", "v0.7.7", true},
		{"0.7.7", "v0.7.7", false},
		{"0.7.8", "v0.7.7", false},
		{"v0.7.6", "0.7.7", true},
	}
	for _, tc := range cases {
		if got := remoteIsNewer(tc.current, tc.remote); got != tc.want {
			t.Errorf("remoteIsNewer(%q, %q) = %v, want %v", tc.current, tc.remote, got, tc.want)
		}
	}
}

func TestAllowedGitHubDownloadURL(t *testing.T) {
	valid := []string{
		"https://github.com/DemedroL432/Flux-/releases/download/v0.7.7/Flux-v0.7.7-win64.zip",
		"https://objects.githubusercontent.com/github-production-release-asset-2e65be/example.zip",
		"https://release-assets.githubusercontent.com/example.exe",
	}
	for _, u := range valid {
		if err := allowedGitHubDownloadURL(u); err != nil {
			t.Errorf("allowedGitHubDownloadURL(%q) unexpected error: %v", u, err)
		}
	}

	invalid := []string{
		"http://github.com/DemedroL432/Flux-/releases/download/v0.7.7/Flux.zip",
		"https://evil.example.com/Flux.zip",
		"https://github.com.evil.com/Flux.zip",
	}
	for _, u := range invalid {
		if err := allowedGitHubDownloadURL(u); err == nil {
			t.Errorf("allowedGitHubDownloadURL(%q) expected error, got nil", u)
		}
	}
}

func TestSetAutoUpdateEnabledAndUIState(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("USERPROFILE", tmpHome)
	t.Setenv("LOCALAPPDATA", tmpHome)

	cfg := config.DefaultConfig()
	a := &App{
		st:         state.New(cfg),
		appVersion: "0.7.7",
	}

	if got := a.GetUIState().AutoUpdateEnabled; !got {
		t.Fatalf("expected AutoUpdateEnabled=true by default, got false")
	}

	if err := a.SetAutoUpdateEnabled(false); err != nil {
		t.Fatalf("SetAutoUpdateEnabled(false) error: %v", err)
	}
	if a.st.Snapshot().AutoUpdateDisabled != true {
		t.Fatalf("expected AutoUpdateDisabled=true after disabling")
	}
	if got := a.GetUIState().AutoUpdateEnabled; got {
		t.Fatalf("expected AutoUpdateEnabled=false after disabling, got true")
	}

	if err := a.SetAutoUpdateEnabled(true); err != nil {
		t.Fatalf("SetAutoUpdateEnabled(true) error: %v", err)
	}
	if a.st.Snapshot().AutoUpdateDisabled != false {
		t.Fatalf("expected AutoUpdateDisabled=false after enabling")
	}
	if got := a.GetUIState().AutoUpdateEnabled; !got {
		t.Fatalf("expected AutoUpdateEnabled=true after enabling, got false")
	}
}
