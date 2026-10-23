//go:build windows

package uiwin

import "github.com/lxn/walk"

func rgbBg() walk.Color        { return walk.RGB(20, 24, 31) }
func rgbSidebar() walk.Color   { return walk.RGB(16, 19, 25) }
func rgbCard() walk.Color      { return walk.RGB(30, 36, 46) }
func rgbCardAlt() walk.Color   { return walk.RGB(35, 42, 54) }
func rgbText() walk.Color      { return walk.RGB(236, 241, 248) }
func rgbTextMuted() walk.Color { return walk.RGB(151, 164, 184) }
func rgbAccent() walk.Color    { return walk.RGB(120, 246, 92) }

