package main

import (
	"encoding/binary"
	"encoding/xml"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestHasStartHiddenArg(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"--foo"}, false},
		{[]string{"--autostart"}, true},
		{[]string{"--minimized"}, true},
		{[]string{"--AUTOSTART"}, true},
		{[]string{"foo", " --Minimized "}, true},
	}
	for _, tc := range cases {
		if got := hasStartHiddenArg(tc.args); got != tc.want {
			t.Errorf("hasStartHiddenArg(%v) = %v, want %v", tc.args, got, tc.want)
		}
	}
}

func TestBuildAutostartTaskXML(t *testing.T) {
	raw, err := buildAutostartTaskXML(`C:\Program Files\Flux\Flux.exe`, "", `TESTPC\tester`)
	if err != nil {
		t.Fatalf("buildAutostartTaskXML() error: %v", err)
	}
	if len(raw) < 4 || raw[0] != 0xFF || raw[1] != 0xFE {
		t.Fatalf("expected UTF-16LE BOM (0xFF 0xFE), got %v", raw[:minLen(len(raw), 4)])
	}
	if (len(raw)-2)%2 != 0 {
		t.Fatalf("expected even UTF-16 byte count after BOM, got %d", len(raw)-2)
	}

	payload := raw[2:]
	u16 := make([]uint16, len(payload)/2)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(payload[i*2 : i*2+2])
	}
	decoded := string(utf16.Decode(u16))
	if !strings.Contains(decoded, `encoding="UTF-16"`) {
		t.Fatalf("expected UTF-16 XML header, got:\n%s", decoded)
	}

	utf8Ready := strings.Replace(decoded, `encoding="UTF-16"`, `encoding="UTF-8"`, 1)
	var doc taskDoc
	if err := xml.Unmarshal([]byte(utf8Ready), &doc); err != nil {
		t.Fatalf("xml.Unmarshal failed: %v\nXML:\n%s", err, decoded)
	}

	if doc.Actions.Exec.Command != `C:\Program Files\Flux\Flux.exe` {
		t.Errorf("Command = %q, want %q", doc.Actions.Exec.Command, `C:\Program Files\Flux\Flux.exe`)
	}
	if doc.Actions.Exec.Arguments != "--autostart" {
		t.Errorf("Arguments = %q, want %q", doc.Actions.Exec.Arguments, "--autostart")
	}
	if doc.Actions.Exec.WorkingDirectory != `C:\Program Files\Flux` {
		t.Errorf("WorkingDirectory = %q, want %q", doc.Actions.Exec.WorkingDirectory, `C:\Program Files\Flux`)
	}
	if doc.Principals.Principal.RunLevel != "HighestAvailable" {
		t.Errorf("RunLevel = %q, want %q", doc.Principals.Principal.RunLevel, "HighestAvailable")
	}
	if doc.Principals.Principal.LogonType != "InteractiveToken" {
		t.Errorf("LogonType = %q, want %q", doc.Principals.Principal.LogonType, "InteractiveToken")
	}
	if doc.Settings.ExecutionTimeLimit != "PT0S" {
		t.Errorf("ExecutionTimeLimit = %q, want %q", doc.Settings.ExecutionTimeLimit, "PT0S")
	}
	if doc.Settings.DisallowStartIfOnBatteries {
		t.Errorf("DisallowStartIfOnBatteries = true, want false")
	}
	if doc.Settings.StopIfGoingOnBatteries {
		t.Errorf("StopIfGoingOnBatteries = true, want false")
	}
	if !doc.Triggers.LogonTrigger.Enabled {
		t.Errorf("LogonTrigger.Enabled = false, want true")
	}
}

func minLen(a, b int) int {
	if a < b {
		return a
	}
	return b
}
