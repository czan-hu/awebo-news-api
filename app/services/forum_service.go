package services

import (
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/pkg/idgen"
	"awebo/app/repositories"
)

type ForumService struct {
	repo *repositories.ForumRepository
}

func NewForumService(repo *repositories.ForumRepository) *ForumService {
	return &ForumService{repo: repo}
}

func (s *ForumService) Categories() []string {
	return entities.ForumCategories
}

func (s *ForumService) ListTopics(category, sort string, page, limit int) (*entities.ForumTopicList, error) {
	page, limit = normalizePage(page, limit)

	topics, total, err := s.repo.ListTopics(repositories.ForumTopicFilter{Category: category, Sort: sort}, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries := make([]entities.ForumTopicSummary, 0, len(topics))
	for _, t := range topics {
		summaries = append(summaries, toTopicSummary(t))
	}

	return &entities.ForumTopicList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *ForumService) CreateTopic(authorID uuid.UUID, input dto.CreateTopicInput) (*entities.ForumTopic, error) {
	if !entities.IsValidForumCategory(input.Category) {
		return nil, exceptions.ValidationError("Проверьте правильность заполнения полей.", map[string]string{
			"category": "Недопустимый раздел форума.",
		})
	}

	id, err := s.generateTopicID(input.Title)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	topic := &models.ForumTopic{
		ID:       id,
		Title:    input.Title,
		Category: input.Category,
		AuthorID: authorID,
		Body:     input.Body,
	}

	if err := s.repo.CreateTopic(topic); err != nil {
		return nil, exceptions.Internal("")
	}

	created, err := s.repo.FindTopicByID(id)
	if err != nil || created == nil {
		return nil, exceptions.Internal("")
	}

	return toTopic(*created), nil
}

func (s *ForumService) GetTopic(id string) (*entities.ForumTopic, error) {
	topic, err := s.repo.FindTopicByID(id)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if topic == nil {
		return nil, exceptions.NotFound("")
	}
	return toTopic(*topic), nil
}

func (s *ForumService) ListComments(topicID, sort string, viewerID *uuid.UUID) ([]entities.ForumComment, error) {
	topic, err := s.repo.FindTopicByID(topicID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if topic == nil {
		return nil, exceptions.NotFound("")
	}

	comments, err := s.repo.ListComments(topicID, sort)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	var votes map[string]int
	if viewerID != nil {
		ids := make([]string, len(comments))
		for i, c := range comments {
			ids[i] = c.ID
		}
		votes, err = s.repo.MyVotes(*viewerID, ids)
		if err != nil {
			return nil, exceptions.Internal("")
		}
	}

	result := make([]entities.ForumComment, 0, len(comments))
	for _, c := range comments {
		myVote := 0
		if votes != nil {
			myVote = votes[c.ID]
		}
		result = append(result, toComment(c, myVote))
	}

	return result, nil
}

func (s *ForumService) CreateComment(topicID string, authorID uuid.UUID, input dto.CreateCommentInput) (*entities.ForumComment, error) {
	topic, err := s.repo.FindTopicByID(topicID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if topic == nil {
		return nil, exceptions.NotFound("Тема или родительский комментарий не найдены.")
	}

	if input.ParentID != nil {
		parent, err := s.repo.FindCommentByID(*input.ParentID)
		if err != nil {
			return nil, exceptions.Internal("")
		}
		if parent == nil || parent.TopicID != topicID {
			return nil, exceptions.NotFound("Тема или родительский комментарий не найдены.")
		}
	}

	id, err := s.generateCommentID()
	if err != nil {
		return nil, exceptions.Internal("")
	}

	comment := &models.ForumComment{
		ID:       id,
		TopicID:  topicID,
		ParentID: input.ParentID,
		AuthorID: authorID,
		Body:     input.Body,
	}

	if err := s.repo.CreateComment(comment); err != nil {
		return nil, exceptions.Internal("")
	}
	if err := s.repo.IncrementCommentCount(topicID, 1); err != nil {
		return nil, exceptions.Internal("")
	}

	created, err := s.repo.FindCommentByID(id)
	if err != nil || created == nil {
		return nil, exceptions.Internal("")
	}

	return &entities.ForumComment{
		ID:        created.ID,
		TopicID:   created.TopicID,
		ParentID:  created.ParentID,
		Author:    entities.PublicUser{Name: created.Author.Name, Username: usernameOf(created.Author), Avatar: created.Author.AvatarPath},
		Body:      created.Body,
		CreatedAt: created.CreatedAt,
		Score:     created.Score,
		MyVote:    1,
	}, nil
}

func (s *ForumService) Vote(commentID string, userID uuid.UUID, direction int) (*entities.VoteResult, error) {
	comment, err := s.repo.FindCommentByID(commentID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if comment == nil {
		return nil, exceptions.NotFound("")
	}

	score, err := s.repo.SetVote(commentID, userID, direction)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return &entities.VoteResult{ID: commentID, Score: score, MyVote: direction}, nil
}

func (s *ForumService) generateTopicID(title string) (string, error) {
	base := idgen.Slugify(title)
	if base == "" {
		base = "topic"
	}

	candidate := base
	for i := 0; i < 5; i++ {
		exists, err := s.repo.TopicExists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
		candidate = base + "-" + idgen.RandomHex(2)
	}
	return base + "-" + idgen.RandomHex(4), nil
}

func (s *ForumService) generateCommentID() (string, error) {
	for i := 0; i < 5; i++ {
		candidate := "c-" + idgen.RandomHex(4)
		exists, err := s.repo.CommentExists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "c-" + idgen.RandomHex(8), nil
}

func toTopicSummary(t models.ForumTopic) entities.ForumTopicSummary {
	return entities.ForumTopicSummary{
		ID:           t.ID,
		Title:        t.Title,
		Category:     t.Category,
		Author:       entities.PublicUser{Name: t.Author.Name, Username: usernameOf(t.Author), Avatar: t.Author.AvatarPath},
		CreatedAt:    t.CreatedAt,
		Pinned:       t.Pinned,
		CommentCount: t.CommentCount,
		Excerpt:      excerpt(t.Body, 200),
	}
}

func toTopic(t models.ForumTopic) *entities.ForumTopic {
	summary := toTopicSummary(t)
	return &entities.ForumTopic{ForumTopicSummary: summary, Body: t.Body}
}

func toComment(c models.ForumComment, myVote int) entities.ForumComment {
	return entities.ForumComment{
		ID:        c.ID,
		TopicID:   c.TopicID,
		ParentID:  c.ParentID,
		Author:    entities.PublicUser{Name: c.Author.Name, Username: usernameOf(c.Author), Avatar: c.Author.AvatarPath},
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
		Score:     c.Score,
		MyVote:    myVote,
	}
}

func excerpt(text string, max int) string {
	r := []rune(text)
	if len(r) <= max {
		return text
	}
	return string(r[:max]) + "…"
}
