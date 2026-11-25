package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf16"
)

const autostartTaskName = "FluxAutostart"

type taskDoc struct {
	XMLName    xml.Name       `xml:"http://schemas.microsoft.com/windows/2004/02/mit/task Task"`
	Version    string         `xml:"version,attr"`
	Triggers   taskTriggers   `xml:"Triggers"`
	Principals taskPrincipals `xml:"Principals"`
	Settings   taskSettings   `xml:"Settings"`
	Actions    taskActions    `xml:"Actions"`
}

type taskTriggers struct {
	LogonTrigger taskLogonTrigger `xml:"LogonTrigger"`
}

type taskLogonTrigger struct {
	Enabled bool   `xml:"Enabled"`
	UserId  string `xml:"UserId,omitempty"`
}

type taskPrincipals struct {
	Principal taskPrincipal `xml:"Principal"`
}

type taskPrincipal struct {
	ID        string `xml:"id,attr"`
	UserId    string `xml:"UserId,omitempty"`
	LogonType string `xml:"LogonType"`
	RunLevel  string `xml:"RunLevel"`
}

type taskSettings struct {
	MultipleInstancesPolicy    string `xml:"MultipleInstancesPolicy"`
	DisallowStartIfOnBatteries bool   `xml:"DisallowStartIfOnBatteries"`
	StopIfGoingOnBatteries     bool   `xml:"StopIfGoingOnBatteries"`
	ExecutionTimeLimit         string `xml:"ExecutionTimeLimit"`
	Enabled                    bool   `xml:"Enabled"`
	Hidden                     bool   `xml:"Hidden"`
}

type taskActions struct {
	Context string         `xml:"Context,attr"`
	Exec    taskExecAction `xml:"Exec"`
}

type taskExecAction struct {
	Command          string `xml:"Command"`
	Arguments        string `xml:"Arguments,omitempty"`
	WorkingDirectory string `xml:"WorkingDirectory,omitempty"`
}

func buildAutostartTaskXML(exePath, workDir, userID string) ([]byte, error) {
	exePath = strings.TrimSpace(exePath)
	if exePath == "" {
		return nil, fmt.Errorf("пустой путь к исполняемому файлу")
	}
	workDir = strings.TrimSpace(workDir)
	if workDir == "" {
		workDir = filepath.Dir(exePath)
	}
	userID = strings.TrimSpace(userID)

	doc := taskDoc{
		Version: "1.2",
		Triggers: taskTriggers{
			LogonTrigger: taskLogonTrigger{
				Enabled: true,
				UserId:  userID,
			},
		},
		Principals: taskPrincipals{
			Principal: taskPrincipal{
				ID:        "Author",
				UserId:    userID,
				LogonType: "InteractiveToken",
				RunLevel:  "HighestAvailable",
			},
		},
		Settings: taskSettings{
			MultipleInstancesPolicy:    "IgnoreNew",
			DisallowStartIfOnBatteries: false,
			StopIfGoingOnBatteries:     false,
			ExecutionTimeLimit:         "PT0S",
			Enabled:                    true,
			Hidden:                     false,
		},
		Actions: taskActions{
			Context: "Author",
			Exec: taskExecAction{
				Command:          exePath,
				Arguments:        "--autostart",
				WorkingDirectory: workDir,
			},
		},
	}

	raw, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	xmlText := `<?xml version="1.0" encoding="UTF-16"?>` + "\r\n" + string(raw) + "\r\n"
	runes := utf16.Encode([]rune(xmlText))
	out := make([]byte, 2+len(runes)*2)
	out[0] = 0xFF
	out[1] = 0xFE
	for i, r := range runes {
		out[2+i*2] = byte(r)
		out[2+i*2+1] = byte(r >> 8)
	}
	return out, nil
}

func runSchtasks(args ...string) (string, error) {
	cmd := exec.Command("schtasks.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return strings.TrimSpace(buf.String()), err
}

func isAutostartTaskRegistered() bool {
	_, err := runSchtasks("/Query", "/TN", autostartTaskName)
	return err == nil
}

func setWindowsAutostart(enabled bool, exePath string) error {
	if !enabled {
		out, err := runSchtasks("/Delete", "/TN", autostartTaskName, "/F")
		if err != nil {
			if !isAutostartTaskRegistered() {
				return nil
			}
			if out != "" {
				return fmt.Errorf("не удалось отключить автозапуск: %s (%w)", out, err)
			}
			return fmt.Errorf("не удалось отключить автозапуск: %w", err)
		}
		return nil
	}

	exePath = strings.TrimSpace(exePath)
	if exePath == "" {
		return fmt.Errorf("не удалось определить путь к исполняемому файлу для автозапуска")
	}

	userID := ""
	if u, err := user.Current(); err == nil {
		userID = strings.TrimSpace(u.Username)
	}

	xmlBytes, err := buildAutostartTaskXML(exePath, filepath.Dir(exePath), userID)
	if err != nil {
		return fmt.Errorf("ошибка формирования задачи автозапуска: %w", err)
	}

	tmp, err := os.CreateTemp("", "flux-autostart-*.xml")
	if err != nil {
		return fmt.Errorf("не удалось создать временный файл задачи автозапуска: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(xmlBytes); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("не удалось записать XML задачи автозапуска: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть XML задачи автозапуска: %w", err)
	}

	out, err := runSchtasks("/Create", "/TN", autostartTaskName, "/XML", tmpPath, "/F")
	if err != nil {
		if out != "" {
			return fmt.Errorf("не удалось зарегистрировать задачу автозапуска: %s (%w)", out, err)
		}
		return fmt.Errorf("не удалось зарегистрировать задачу автозапуска: %w", err)
	}
	return nil
}
