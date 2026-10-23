package dpi

import "strings"

// EnsureStrategiesLuaInit добавляет --lua-init=@lua/strategies.lua после строки с zapret-auto.lua, если строки ещё нет.
func EnsureStrategiesLuaInit(body string) string {
	if strings.Contains(body, "strategies.lua") {
		return body
	}
	lines := strings.Split(body, "\n")
	var out []string
	for _, L := range lines {
		out = append(out, L)
		if strings.Contains(L, "zapret-auto.lua") {
			out = append(out, "--lua-init=@lua/strategies.lua")
		}
	}
	return strings.Join(out, "\n")
}
