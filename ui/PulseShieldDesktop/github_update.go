package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// Задаются при сборке: -ldflags "-X main.GitHubOwner=... -X main.GitHubRepo=..."
var (
	GitHubOwner = "DemedroL432"
	GitHubRepo  = "Flux-"
)

// UpdateInfo результат проверки GitHub Releases (latest).
type UpdateInfo struct {
	OK                   bool   `json:"ok"`
	CurrentVersion       string `json:"currentVersion"`
	LatestVersion        string `json:"latestVersion"`
	LatestTag            string `json:"latestTag"`
	UpToDate             bool   `json:"upToDate"`
	ReleaseURL           string `json:"releaseUrl"`
	DownloadURL          string `json:"downloadUrl"`
	InstallerDownloadURL string `json:"installerDownloadUrl"`
	Message              string `json:"message"`
	IsPrerelease         bool   `json:"isPrerelease"`
}

func latestReleaseAPI() string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", strings.TrimSpace(GitHubOwner), strings.TrimSpace(GitHubRepo))
}

type ghRelease struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func pickDownloadURL(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) string {
	if len(assets) == 0 {
		return ""
	}
	type scored struct {
		url   string
		score int
	}
	var list []scored
	for _, a := range assets {
		n := strings.ToLower(a.Name)
		u := strings.TrimSpace(a.BrowserDownloadURL)
		if u == "" {
			continue
		}
		s := 0
		if strings.HasSuffix(n, ".zip") {
			s += 10
		}
		if strings.Contains(n, "flux") || strings.Contains(n, "pulseshield") {
			s += 5
		}
		if strings.Contains(n, "win64") || strings.Contains(n, "windows") {
			s += 3
		}
		if strings.HasSuffix(n, ".exe") && s < 10 {
			s += 4
		}
		list = append(list, scored{url: u, score: s})
	}
	best := ""
	bestScore := -1
	for _, x := range list {
		if x.score > bestScore {
			bestScore = x.score
			best = x.url
		}
	}
	if best != "" {
		return best
	}
	for _, a := range assets {
		if u := strings.TrimSpace(a.BrowserDownloadURL); u != "" {
			return u
		}
	}
	return ""
}

func pickInstallerURL(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}) string {
	best := ""
	bestScore := -1
	for _, a := range assets {
		n := strings.ToLower(a.Name)
		u := strings.TrimSpace(a.BrowserDownloadURL)
		if u == "" || !strings.HasSuffix(n, ".exe") {
			continue
		}
		if !strings.Contains(n, "setup") && !strings.Contains(n, "installer") {
			continue
		}
		s := 0
		if strings.Contains(n, "flux") || strings.Contains(n, "pulseshield") {
			s += 6
		}
		if strings.Contains(n, "setup") {
			s += 4
		}
		if s > bestScore {
			bestScore = s
			best = u
		}
	}
	return best
}

func remoteIsNewer(current, remoteTag string) bool {
	rt := strings.TrimSpace(remoteTag)
	if rt == "" {
		return false
	}
	rc := semver.Canonical("v" + strings.TrimPrefix(rt, "v"))
	ct := strings.TrimSpace(current)
	if ct == "" || ct == "dev" {
		return semver.IsValid(rc)
	}
	lc := semver.Canonical("v" + strings.TrimPrefix(ct, "v"))
	if semver.IsValid(lc) && semver.IsValid(rc) {
		return semver.Compare(lc, rc) < 0
	}
	return strings.TrimPrefix(ct, "v") != strings.TrimPrefix(rt, "v")
}

func clipBody(body []byte, n int) string {
	if len(body) <= n {
		return strings.TrimSpace(string(body))
	}
	return strings.TrimSpace(string(body[:n])) + "…"
}

// GetUpdateInfo запрашивает последний релиз на GitHub и сравнивает с версией сборки.
func (a *App) GetUpdateInfo() UpdateInfo {
	out := UpdateInfo{
		CurrentVersion: strings.TrimSpace(a.GetAppVersion()),
	}
	if out.CurrentVersion == "" {
		out.CurrentVersion = "dev"
	}

	o := strings.TrimSpace(GitHubOwner)
	r := strings.TrimSpace(GitHubRepo)
	if o == "" || r == "" {
		out.Message = "Репозиторий GitHub не настроен при сборке (GitHubOwner/GitHubRepo)."
		return out
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, latestReleaseAPI(), nil)
	if err != nil {
		out.Message = "Ошибка запроса: " + err.Error()
		return out
	}
	req.Header.Set("User-Agent", "Flux/"+out.CurrentVersion)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		out.Message = "Не удалось связаться с GitHub: " + err.Error()
		return out
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusNotFound {
		out.Message = fmt.Sprintf("Релизы не найдены (репозиторий %s/%s или ещё нет тегов). Создайте релиз на GitHub.", o, r)
		return out
	}
	if resp.StatusCode != http.StatusOK {
		out.Message = fmt.Sprintf("GitHub: %s — %s", resp.Status, clipBody(body, 220))
		return out
	}

	var rel ghRelease
	if err := json.Unmarshal(body, &rel); err != nil || rel.TagName == "" {
		out.Message = "Не удалось разобрать ответ GitHub."
		return out
	}

	out.OK = true
	out.LatestTag = strings.TrimSpace(rel.TagName)
	out.LatestVersion = strings.TrimPrefix(out.LatestTag, "v")
	out.ReleaseURL = strings.TrimSpace(rel.HTMLURL)
	out.DownloadURL = pickDownloadURL(rel.Assets)
	out.InstallerDownloadURL = pickInstallerURL(rel.Assets)
	out.IsPrerelease = rel.Prerelease
	out.UpToDate = !remoteIsNewer(out.CurrentVersion, rel.TagName)

	if out.UpToDate {
		out.Message = fmt.Sprintf("У вас актуальная версия (%s). Последний релиз на GitHub: %s.", out.CurrentVersion, out.LatestTag)
		return out
	}

	if out.CurrentVersion == "dev" {
		out.Message = fmt.Sprintf("Сборка разработки. На GitHub последний релиз: %s. Скачайте portable с страницы релиза.", out.LatestTag)
	} else {
		out.Message = fmt.Sprintf("Доступна новая версия: %s (у вас %s).", out.LatestTag, out.CurrentVersion)
	}
	return out
}

// CheckForUpdates оставлен для совместимости; возвращает только текст сообщения.
func (a *App) CheckForUpdates() string {
	return a.GetUpdateInfo().Message
}
