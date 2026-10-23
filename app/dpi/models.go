package dpi

// DpiList описывает один файл lists/*.txt рядом с Zapret (имя без суффикса .txt).
type DpiList struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}
