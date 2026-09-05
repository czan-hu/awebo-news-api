package entities

type Category struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	// Icon — имя иконки Lucide в формате фронтенда (i-lucide-xxx),
	// та же конвенция, что и у UserLink.Icon.
	Icon string `json:"icon"`
}
