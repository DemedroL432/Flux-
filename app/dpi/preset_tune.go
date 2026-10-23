package dpi

// GeneratedTuningCandidate описывает один вариант пресета, собранный целиком в памяти (без чтения presets/*.txt с диска).
type GeneratedTuningCandidate struct {
	ID    string
	Label string
	Build func() []byte
}

// DefaultGeneratedTuningCandidates — набор вариантов для автоподбора: разные базы (Stable / Gaming) и правки wf/tcp/udp.
func DefaultGeneratedTuningCandidates() []GeneratedTuningCandidate {
	return []GeneratedTuningCandidate{
		{
			ID:    "stable-default",
			Label: "Сгенерированный Stable (стандартные wf)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("stable-default", "Stable, без изменений портов")
				core := TrimLeadingPresetComments(rawGamingStable)
				out := append([]byte(h), core...)
				return out
			},
		},
		{
			ID:    "stable-narrow-tcp",
			Label: "Сгенерированный Stable (узкий TCP: 80,443)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("stable-narrow-tcp", "Stable, только 80/443 TCP")
				core := TrimLeadingPresetComments(rawGamingStable)
				core = ReplaceWFPorts(core, "80,443", "443-65535")
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "stable-ipcache-8400",
			Label: "Сгенерированный Stable (ipcache 8400)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("stable-ipcache-8400", "Stable, ipcache-lifetime=8400")
				core := TrimLeadingPresetComments(rawGamingStable)
				core = ReplaceIPCacheLifetime(core, "8400")
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "gaming-full",
			Label: "Сгенерированный Gaming (полный набор lua; Discord voice как zapret ALT3)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("gaming-full", "Gamingnew: UDP голос Discord — порты 19294–19344/50000–50100 + quic_initial_dbankcloud_ru fake")
				core := TrimLeadingPresetComments(rawGamingNew)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "gaming-narrow-tcp",
			Label: "Сгенерированный Gaming (узкий TCP)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("gaming-narrow-tcp", "Gaming, только 80/443 TCP")
				core := TrimLeadingPresetComments(rawGamingNew)
				core = ReplaceWFPorts(core, "80,443", "443-65535")
				return append([]byte(h), core...)
			},
		},
	}
}
