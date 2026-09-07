package entities

// Contact — способ связи с редакцией, выводится иконкой в подвале сайта.
type Contact struct {
	Label string `json:"label"`
	Href  string `json:"href"`
	// Icon — «сырая» разметка <svg>...</svg>, которую фронтенд вставляет как
	// есть (см. SvgIcon.vue). Задаётся напрямую в БД.
	Icon string `json:"icon"`
}
