package entities

type Category struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	// Icon — «сырая» разметка <svg>...</svg>, которую фронтенд вставляет как
	// есть (см. SvgIcon.vue). Можно вписать любую SVG-иконку прямо в БД.
	Icon string `json:"icon"`
}
