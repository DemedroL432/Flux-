package dpi

// GeneratedTuningCandidate описывает один вариант пресета, собранный целиком в памяти (без чтения presets/*.txt с диска).
type GeneratedTuningCandidate struct {
	ID    string
	Label string
	Build func() []byte
}

// DefaultGeneratedTuningCandidates — набор боевых вариантов десинхронизации (на базе zapret-discord-youtube):
// перебирает различные стратегии обхода для Discord (ALT11, General, ALT, Fake TLS Auto, ALT4, ALT2, Multidisorder)
// с гарантированным перехватом WebRTC UDP голоса Discord (50000–65535, 19294–19344) и фейковыми пакетами.
func DefaultGeneratedTuningCandidates() []GeneratedTuningCandidate {
	return []GeneratedTuningCandidate{
		{
			ID:    "hybrid-hostfakesplit-voice",
			Label: "Проверенный гибрид: HostFakeSplit Multi + WebRTC Voice · Рекомендуемый",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("hybrid-hostfakesplit-voice", "Гибрид: HostFakeSplit Multi (0.7.3 TCP запуск) + ACTIVE_DISCORD_UDP Voice (0.7.5)")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=hostfakesplit_multi:hosts=google.com,vimeo.com:tcp_ts=-1000:tcp_md5:repeats=2"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "alt11-multisplit-4pda",
			Label: "ALT11: Fake TLS + Multisplit 4PDA (seqovl=568)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("alt11-multisplit-4pda", "ALT11: Fake TLS (4PDA) + Multisplit (seqovl=568, pos=1) + ACTIVE_DISCORD_UDP")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=fake:blob=tls_4pda:tcp_ts=1:repeats=8:payload=tls_client_hello\n--lua-desync=multisplit:pos=1:seqovl=568:seqovl_pattern=tls_4pda:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "general-multisplit-4pda",
			Label: "General: Multisplit 4PDA (seqovl=568, pos=1)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("general-multisplit-4pda", "General: Multisplit без фейков (seqovl=568, pos=1, pattern=tls_4pda)")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=multisplit:pos=1:seqovl=568:seqovl_pattern=tls_4pda:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "alt-multisplit-google",
			Label: "ALT: Fake TLS + Multisplit Google (seqovl=681, pos=1)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("alt-multisplit-google", "ALT: Fake TLS (Google) + Multisplit (seqovl=681, pos=1, pattern=tls_google)")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=fake:blob=tls_google:tcp_ts=1:repeats=8:payload=tls_client_hello\n--lua-desync=multisplit:pos=1:seqovl=681:seqovl_pattern=tls_google:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "faketls-auto-badseq",
			Label: "Fake TLS Auto: badseq + Multidisorder (pos=1,midsld)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("faketls-auto-badseq", "Fake TLS Auto: 0x00 blob, badseq, dupsid + Multidisorder pos=1,midsld")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=fake:blob=0x00000000000000000000000000000000:repeats=11:tcp_ack=-66000:tls_mod=rnd,dupsid,sni=www.google.com:payload=tls_client_hello\n--lua-desync=multidisorder:pos=1,midsld:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "alt4-fakedsplit-1000",
			Label: "ALT4: Fake TLS + Multisplit (badseq increment=1000)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("alt4-fakedsplit-1000", "ALT4: Fake TLS (Google, repeats=6) + badseq 1000 + Multisplit pos=1")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=fake:blob=tls_google:repeats=6:tcp_ack=-1000:tcp_ts_up:payload=tls_client_hello\n--lua-desync=multisplit:pos=1:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "alt2-multisplit-pos2",
			Label: "ALT2: Multisplit Google (seqovl=652, pos=2)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("alt2-multisplit-pos2", "ALT2: Multisplit (seqovl=652, pos=2, pattern=tls_google)")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=multisplit:pos=2:seqovl=652:seqovl_pattern=tls_google:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "multidisorder-seqovl",
			Label: "Multidisorder: seqovl=1 (pos=1,host+2,sld+2...)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("multidisorder-seqovl", "Multidisorder: pos=1,host+2,sld+2,sld+5,sniext+1,sniext+2,endhost-2:seqovl=1")
				core := TrimLeadingPresetComments(rawGamingNew)
				desync := "--lua-desync=multidisorder:pos=1,host+2,sld+2,sld+5,sniext+1,sniext+2,endhost-2:seqovl=1"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
		{
			ID:    "stable-alt11",
			Label: "Stable: ipcache 86400 + ALT11 (полный Discord)",
			Build: func() []byte {
				h := PulseShieldGeneratedHeader("stable-alt11", "Stable: база GamingStable, ipcache=86400 + ALT11 Fake TLS + Multisplit 4PDA")
				core := TrimLeadingPresetComments(rawGamingStable)
				desync := "--lua-desync=fake:blob=tls_4pda:tcp_ts=1:repeats=8:payload=tls_client_hello\n--lua-desync=multisplit:pos=1:seqovl=568:seqovl_pattern=tls_4pda:payload=tls_client_hello"
				core = ReplaceDiscordDesync(core, desync)
				return append([]byte(h), core...)
			},
		},
	}
}
