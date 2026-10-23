package dpi

import _ "embed"

// Встроенные эталоны структуры пресета (копии из third_party/zapret/presets для автономной генерации без чтения чужих файлов с диска во время подбора).
//
//go:embed embed/Gamingnew-Stable.txt
var rawGamingStable []byte

//go:embed embed/Gamingnew.txt
var rawGamingNew []byte
