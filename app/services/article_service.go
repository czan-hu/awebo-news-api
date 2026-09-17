package services

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/pkg/idgen"
	"awebo/app/repositories"
)

type ArticleService struct {
	repo         *repositories.ArticleRepository
	categoryRepo *repositories.CategoryRepository
	userRepo     *repositories.UserRepository
}

func NewArticleService(repo *repositories.ArticleRepository, categoryRepo *repositories.CategoryRepository, userRepo *repositories.UserRepository) *ArticleService {
	return &ArticleService{repo: repo, categoryRepo: categoryRepo, userRepo: userRepo}
}

// Create publishes a new article on behalf of authorID. Callers must already
// have checked entities.CanPublishArticles for the author's role.
func (s *ArticleService) Create(authorID uuid.UUID, input dto.CreateArticleInput) (*entities.Article, error) {
	categoryExists, err := s.categoryRepo.Exists(input.Category)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if !categoryExists {
		return nil, exceptions.ValidationError("Проверьте правильность заполнения полей.", map[string]string{
			"category": "Такого раздела не существует.",
		})
	}

	if err := entities.ValidateContentBlocks(input.Content); err != nil {
		return nil, exceptions.ValidationError(err.Error(), map[string]string{"content": err.Error()})
	}

	slug, err := s.generateSlug(input.Title)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	tags := make([]string, 0, len(input.Tags))
	for _, t := range input.Tags {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}

	article := &models.Article{
		Slug:           slug,
		Title:          strings.TrimSpace(input.Title),
		Excerpt:        strings.TrimSpace(input.Excerpt),
		CategorySlug:   input.Category,
		AuthorID:       authorID,
		PublishedAt:    time.Now(),
		ReadingMinutes: entities.EstimateReadingMinutes(input.Content),
		Cover:          input.Cover,
		Featured:       false,
		Tags:           tags,
		Content:        input.Content,
	}

	if err := s.repo.Create(article); err != nil {
		return nil, exceptions.Internal("")
	}

	created, err := s.repo.FindBySlug(slug)
	if err != nil || created == nil {
		return nil, exceptions.Internal("")
	}

	summary, err := s.toSummary(*created, &authorID)
	if err != nil {
		return nil, err
	}

	content := make([]entities.ContentBlock, len(created.Content))
	copy(content, created.Content)

	return &entities.Article{ArticleSummary: summary, Content: content}, nil
}

// Update edits an existing article. Only its author or an admin may do so;
// the slug is deliberately left unchanged so existing links keep working.
func (s *ArticleService) Update(userID uuid.UUID, slug string, input dto.UpdateArticleInput) (*entities.Article, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}
	if !s.canManage(userID, article.AuthorID) {
		return nil, exceptions.Forbidden("Редактировать статью может только её автор.")
	}

	categoryExists, err := s.categoryRepo.Exists(input.Category)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if !categoryExists {
		return nil, exceptions.ValidationError("Проверьте правильность заполнения полей.", map[string]string{
			"category": "Такого раздела не существует.",
		})
	}

	if err := entities.ValidateContentBlocks(input.Content); err != nil {
		return nil, exceptions.ValidationError(err.Error(), map[string]string{"content": err.Error()})
	}

	tags := make([]string, 0, len(input.Tags))
	for _, t := range input.Tags {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}

	article.Title = strings.TrimSpace(input.Title)
	article.Excerpt = strings.TrimSpace(input.Excerpt)
	article.CategorySlug = input.Category
	article.Cover = input.Cover
	article.ReadingMinutes = entities.EstimateReadingMinutes(input.Content)
	article.Tags = tags
	article.Content = input.Content

	if err := s.repo.Update(article); err != nil {
		return nil, exceptions.Internal("")
	}

	updated, err := s.repo.FindBySlug(slug)
	if err != nil || updated == nil {
		return nil, exceptions.Internal("")
	}

	summary, err := s.toSummary(*updated, &userID)
	if err != nil {
		return nil, err
	}

	content := make([]entities.ContentBlock, len(updated.Content))
	copy(content, updated.Content)

	return &entities.Article{ArticleSummary: summary, Content: content}, nil
}

// Delete removes an article. Only its author or an admin may do so.
func (s *ArticleService) Delete(userID uuid.UUID, slug string) error {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return exceptions.Internal("")
	}
	if article == nil {
		return exceptions.NotFound("")
	}
	if !s.canManage(userID, article.AuthorID) {
		return exceptions.Forbidden("Удалить статью может только её автор.")
	}
	if err := s.repo.Delete(article.ID); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

// canManage allows the article's own author, plus any admin.
func (s *ArticleService) canManage(userID, authorID uuid.UUID) bool {
	if userID == authorID {
		return true
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil || user == nil {
		return false
	}
	return entities.IsAdmin(user.Role)
}

func (s *ArticleService) generateSlug(title string) (string, error) {
	base := idgen.Slugify(title)
	if base == "" {
		base = "article"
	}

	candidate := base
	for i := 0; i < 5; i++ {
		existing, err := s.repo.FindBySlug(candidate)
		if err != nil {
			return "", err
		}
		if existing == nil {
			return candidate, nil
		}
		candidate = base + "-" + idgen.RandomHex(2)
	}
	return base + "-" + idgen.RandomHex(4), nil
}

func normalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 12
	}
	return page, limit
}

func (s *ArticleService) List(query dto.ArticleListQuery, viewerID *uuid.UUID) (*entities.ArticleList, error) {
	page, limit := normalizePage(query.Page, query.Limit)

	filter := repositories.ArticleFilter{
		Category: query.Category,
		Tag:      query.Tag,
		Author:   query.Author,
		Featured: query.Featured,
		Sort:     query.Sort,
	}

	articles, total, err := s.repo.List(filter, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummaries(articles, viewerID)
	if err != nil {
		return nil, err
	}

	return &entities.ArticleList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ArticleService) GetBySlug(slug string, viewerID *uuid.UUID) (*entities.Article, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}

	summary, err := s.toSummary(*article, viewerID)
	if err != nil {
		return nil, err
	}

	content := make([]entities.ContentBlock, len(article.Content))
	copy(content, article.Content)

	return &entities.Article{ArticleSummary: summary, Content: content}, nil
}

func (s *ArticleService) Related(slug string, limit int, viewerID *uuid.UUID) ([]entities.ArticleSummary, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}

	if limit < 1 || limit > 12 {
		limit = 3
	}

	related, err := s.repo.Related(article.CategorySlug, article.ID, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return s.toSummaries(related, viewerID)
}

func (s *ArticleService) ListByAuthor(username string, viewerID *uuid.UUID, page, limit int) (*entities.ArticleList, error) {
	page, limit = normalizePage(page, limit)

	articles, total, err := s.repo.List(repositories.ArticleFilter{Author: username}, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummaries(articles, viewerID)
	if err != nil {
		return nil, err
	}

	return &entities.ArticleList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ArticleService) ListLiked(userID uuid.UUID, page, limit int) (*entities.ArticleList, error) {
	page, limit = normalizePage(page, limit)

	articles, total, err := s.repo.ListLiked(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummaries(articles, &userID)
	if err != nil {
		return nil, err
	}

	return &entities.ArticleList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ArticleService) ListSaved(userID uuid.UUID, page, limit int) (*entities.ArticleList, error) {
	page, limit = normalizePage(page, limit)

	articles, total, err := s.repo.ListSaved(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummaries(articles, &userID)
	if err != nil {
		return nil, err
	}

	return &entities.ArticleList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ArticleService) ListHistory(userID uuid.UUID, page, limit int) (*entities.ArticleList, error) {
	page, limit = normalizePage(page, limit)

	articles, total, err := s.repo.ListHistory(userID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummaries(articles, &userID)
	if err != nil {
		return nil, err
	}

	return &entities.ArticleList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ArticleService) Like(slug string, userID uuid.UUID) (*entities.ArticleInteraction, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Like(article.ID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.interaction(article.ID, slug, userID)
}

func (s *ArticleService) Unlike(slug string, userID uuid.UUID) (*entities.ArticleInteraction, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Unlike(article.ID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.interaction(article.ID, slug, userID)
}

func (s *ArticleService) SaveArticle(slug string, userID uuid.UUID) (*entities.ArticleInteraction, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Save(article.ID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.interaction(article.ID, slug, userID)
}

func (s *ArticleService) UnsaveArticle(slug string, userID uuid.UUID) (*entities.ArticleInteraction, error) {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if article == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Unsave(article.ID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.interaction(article.ID, slug, userID)
}

func (s *ArticleService) RecordView(slug string, userID uuid.UUID) error {
	article, err := s.repo.FindBySlug(slug)
	if err != nil {
		return exceptions.Internal("")
	}
	if article == nil {
		return exceptions.NotFound("")
	}
	if err := s.repo.RecordView(article.ID, userID); err != nil {
		return exceptions.Internal("")
	}
	return nil
}

func (s *ArticleService) interaction(articleID int, slug string, userID uuid.UUID) (*entities.ArticleInteraction, error) {
	liked, err := s.repo.IsLiked(articleID, userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	saved, err := s.repo.IsSaved(articleID, userID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	count, err := s.repo.LikeCount(articleID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	return &entities.ArticleInteraction{Slug: slug, Liked: liked, Saved: saved, LikeCount: count}, nil
}

// authorUsername safely dereferences the nullable users.username column
// (a freshly verified account has no username yet).
func authorUsername(u models.User) string {
	if u.Username == nil {
		return ""
	}
	return *u.Username
}

func articleSummaryOf(a models.Article) entities.ArticleSummary {
	return entities.ArticleSummary{
		ID:             a.ID,
		Slug:           a.Slug,
		Title:          a.Title,
		Excerpt:        a.Excerpt,
		Category:       a.CategorySlug,
		Author:         entities.Author{Name: a.Author.Name, Username: authorUsername(a.Author), Role: a.Author.Title, Avatar: a.Author.AvatarPath},
		PublishedAt:    a.PublishedAt,
		UpdatedAt:      a.UpdatedAt,
		ReadingMinutes: a.ReadingMinutes,
		Cover:          a.Cover,
		Tags:           append([]string{}, a.Tags...),
		Featured:       a.Featured,
	}
}

func (s *ArticleService) toSummary(a models.Article, viewerID *uuid.UUID) (entities.ArticleSummary, error) {
	summary := articleSummaryOf(a)

	if viewerID != nil {
		liked, err := s.repo.IsLiked(a.ID, *viewerID)
		if err != nil {
			return summary, exceptions.Internal("")
		}
		saved, err := s.repo.IsSaved(a.ID, *viewerID)
		if err != nil {
			return summary, exceptions.Internal("")
		}
		summary.Liked = &liked
		summary.Saved = &saved
	}

	return summary, nil
}

func (s *ArticleService) toSummaries(articles []models.Article, viewerID *uuid.UUID) ([]entities.ArticleSummary, error) {
	summaries := make([]entities.ArticleSummary, 0, len(articles))

	var likedMap, savedMap map[int]bool
	if viewerID != nil {
		ids := make([]int, len(articles))
		for i, a := range articles {
			ids[i] = a.ID
		}

		var err error
		likedMap, err = s.repo.LikedArticleIDs(*viewerID, ids)
		if err != nil {
			return nil, exceptions.Internal("")
		}
		savedMap, err = s.repo.SavedArticleIDs(*viewerID, ids)
		if err != nil {
			return nil, exceptions.Internal("")
		}
	}

	for _, a := range articles {
		summary := articleSummaryOf(a)
		if viewerID != nil {
			liked := likedMap[a.ID]
			saved := savedMap[a.ID]
			summary.Liked = &liked
			summary.Saved = &saved
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}
