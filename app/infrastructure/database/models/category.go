package models

type Category struct {
	Slug  string `gorm:"primaryKey"`
	Title string `gorm:"not null"`
}

func (Category) TableName() string { return "categories" }
