package models

type Category struct {
	Slug  string `gorm:"primaryKey"`
	Title string `gorm:"not null"`
	Icon  string `gorm:"not null;default:'i-lucide-hash'"`
}

func (Category) TableName() string { return "categories" }
