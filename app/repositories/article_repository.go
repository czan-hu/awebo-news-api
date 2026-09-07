package repositories

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"awebo/app/infrastructure/database/models"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) Create(article *models.Article) error {
	return r.db.Create(article).Error
}

// Update rewrites the editable fields of an existing article (slug, author and
// publish date stay untouched).
func (r *ArticleRepository) Update(article *models.Article) error {
	article.UpdatedAt = time.Now()
	return r.db.Model(&models.Article{ID: article.ID}).
		Select("Title", "Excerpt", "CategorySlug", "Cover", "ReadingMinutes", "Tags", "Content", "UpdatedAt").
		Updates(article).Error
}

// Delete hard-deletes the article; article_likes/saves/views rows are removed
// by ON DELETE CASCADE (see 000005_create_article_engagement_tables).
func (r *ArticleRepository) Delete(id int) error {
	return r.db.Delete(&models.Article{}, "id = ?", id).Error
}

type ArticleFilter struct {
	Category string
	Tag      string
	Author   string // username
	Featured *bool
	Sort     string // "new" | "popular"
}

func (r *ArticleRepository) List(filter ArticleFilter, page, limit int) ([]models.Article, int64, error) {
	query := r.db.Model(&models.Article{})

	if filter.Category != "" {
		query = query.Where("articles.category_slug = ?", filter.Category)
	}
	if filter.Tag != "" {
		tagJSON, _ := json.Marshal([]string{filter.Tag})
		query = query.Where("articles.tags @> ?::jsonb", string(tagJSON))
	}
	if filter.Author != "" {
		query = query.Joins("JOIN users ON users.id = articles.author_id").
			Where("users.username = ?", filter.Author)
	}
	if filter.Featured != nil {
		query = query.Where("articles.featured = ?", *filter.Featured)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	listQuery := query.Preload("Author")
	if filter.Sort == "popular" {
		listQuery = listQuery.
			Joins("LEFT JOIN (SELECT article_id, COUNT(*) AS like_count FROM article_likes GROUP BY article_id) al ON al.article_id = articles.id").
			Order("COALESCE(al.like_count, 0) DESC, articles.published_at DESC")
	} else {
		listQuery = listQuery.Order("articles.published_at DESC")
	}

	var articles []models.Article
	if err := listQuery.Offset((page - 1) * limit).Limit(limit).Find(&articles).Error; err != nil {
		return nil, 0, err
	}

	return articles, total, nil
}

func (r *ArticleRepository) FindBySlug(slug string) (*models.Article, error) {
	var article models.Article
	if err := r.db.Preload("Author").Where("slug = ?", slug).First(&article).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &article, nil
}

func (r *ArticleRepository) Related(categorySlug string, excludeID, limit int) ([]models.Article, error) {
	var articles []models.Article
	if err := r.db.Preload("Author").
		Where("category_slug = ? AND id <> ?", categorySlug, excludeID).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func (r *ArticleRepository) ListLiked(userID uuid.UUID, page, limit int) ([]models.Article, int64, error) {
	base := r.db.Model(&models.Article{}).
		Joins("JOIN article_likes ON article_likes.article_id = articles.id").
		Where("article_likes.user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []models.Article
	if err := base.Preload("Author").
		Order("article_likes.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *ArticleRepository) ListSaved(userID uuid.UUID, page, limit int) ([]models.Article, int64, error) {
	base := r.db.Model(&models.Article{}).
		Joins("JOIN article_saves ON article_saves.article_id = articles.id").
		Where("article_saves.user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []models.Article
	if err := base.Preload("Author").
		Order("article_saves.created_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *ArticleRepository) ListHistory(userID uuid.UUID, page, limit int) ([]models.Article, int64, error) {
	base := r.db.Model(&models.Article{}).
		Joins("JOIN article_views ON article_views.article_id = articles.id").
		Where("article_views.user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var articles []models.Article
	if err := base.Preload("Author").
		Order("article_views.viewed_at DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, 0, err
	}
	return articles, total, nil
}

func (r *ArticleRepository) Like(articleID int, userID uuid.UUID) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.ArticleLike{ArticleID: articleID, UserID: userID}).Error
}

func (r *ArticleRepository) Unlike(articleID int, userID uuid.UUID) error {
	return r.db.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&models.ArticleLike{}).Error
}

func (r *ArticleRepository) Save(articleID int, userID uuid.UUID) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.ArticleSave{ArticleID: articleID, UserID: userID}).Error
}

func (r *ArticleRepository) Unsave(articleID int, userID uuid.UUID) error {
	return r.db.Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&models.ArticleSave{}).Error
}

func (r *ArticleRepository) RecordView(articleID int, userID uuid.UUID) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "article_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"viewed_at"}),
	}).Create(&models.ArticleView{ArticleID: articleID, UserID: userID, ViewedAt: time.Now()}).Error
}

func (r *ArticleRepository) IsLiked(articleID int, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ArticleLike{}).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ArticleRepository) IsSaved(articleID int, userID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.ArticleSave{}).
		Where("article_id = ? AND user_id = ?", articleID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ArticleRepository) LikeCount(articleID int) (int64, error) {
	var count int64
	if err := r.db.Model(&models.ArticleLike{}).Where("article_id = ?", articleID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ArticleRepository) LikedArticleIDs(userID uuid.UUID, articleIDs []int) (map[int]bool, error) {
	if len(articleIDs) == 0 {
		return map[int]bool{}, nil
	}
	var rows []int
	if err := r.db.Model(&models.ArticleLike{}).
		Where("user_id = ? AND article_id IN ?", userID, articleIDs).
		Pluck("article_id", &rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int]bool, len(rows))
	for _, id := range rows {
		result[id] = true
	}
	return result, nil
}

func (r *ArticleRepository) SavedArticleIDs(userID uuid.UUID, articleIDs []int) (map[int]bool, error) {
	if len(articleIDs) == 0 {
		return map[int]bool{}, nil
	}
	var rows []int
	if err := r.db.Model(&models.ArticleSave{}).
		Where("user_id = ? AND article_id IN ?", userID, articleIDs).
		Pluck("article_id", &rows).Error; err != nil {
		return nil, err
	}
	result := make(map[int]bool, len(rows))
	for _, id := range rows {
		result[id] = true
	}
	return result, nil
}

func (r *ArticleRepository) Search(q string, limit int) ([]models.Article, error) {
	var articles []models.Article
	like := "%" + q + "%"
	if err := r.db.Preload("Author").
		Where("title ILIKE ? OR excerpt ILIKE ?", like, like).
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}
