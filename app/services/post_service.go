package services

import (
	"github.com/google/uuid"

	"awebo/app/dto"
	"awebo/app/entities"
	"awebo/app/exceptions"
	"awebo/app/infrastructure/database/models"
	"awebo/app/repositories"
)

type PostService struct {
	repo           *repositories.PostRepository
	friendshipRepo *repositories.FriendshipRepository
}

func NewPostService(repo *repositories.PostRepository, friendshipRepo *repositories.FriendshipRepository) *PostService {
	return &PostService{repo: repo, friendshipRepo: friendshipRepo}
}

func (s *PostService) Create(authorID uuid.UUID, input dto.CreatePostInput) (*entities.Post, error) {
	post := &models.Post{
		ID:        uuid.New(),
		AuthorID:  authorID,
		Body:      input.Body,
		ImagePath: input.Image,
	}
	if err := s.repo.Create(post); err != nil {
		return nil, exceptions.Internal("")
	}

	created, err := s.repo.FindByID(post.ID)
	if err != nil || created == nil {
		return nil, exceptions.Internal("")
	}

	liked := false
	return toPost(*created, &liked), nil
}

// Feed lists posts for the home timeline: the viewer's own posts plus their
// accepted friends', newest first.
func (s *PostService) Feed(viewerID uuid.UUID, page, limit int) (*entities.PostList, error) {
	page, limit = normalizePage(page, limit)

	friendIDs, err := s.friendshipRepo.FriendIDs(viewerID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	authorIDs := append(friendIDs, viewerID)

	posts, total, err := s.repo.ListByAuthors(authorIDs, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	summaries, err := s.toSummariesWithViewer(posts, &viewerID)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	return &entities.PostList{Data: summaries, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

func (s *PostService) GetByID(viewerID *uuid.UUID, id uuid.UUID) (*entities.Post, error) {
	post, err := s.repo.FindByID(id)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if post == nil {
		return nil, exceptions.NotFound("")
	}

	var liked *bool
	if viewerID != nil {
		ok, err := s.repo.IsLiked(post.ID, *viewerID)
		if err != nil {
			return nil, exceptions.Internal("")
		}
		liked = &ok
	}

	return toPost(*post, liked), nil
}

func (s *PostService) Like(userID, postID uuid.UUID) (*entities.Post, error) {
	post, err := s.repo.FindByID(postID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if post == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Like(postID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.GetByID(&userID, postID)
}

func (s *PostService) Unlike(userID, postID uuid.UUID) (*entities.Post, error) {
	post, err := s.repo.FindByID(postID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if post == nil {
		return nil, exceptions.NotFound("")
	}
	if err := s.repo.Unlike(postID, userID); err != nil {
		return nil, exceptions.Internal("")
	}
	return s.GetByID(&userID, postID)
}

func (s *PostService) CreateComment(authorID, postID uuid.UUID, input dto.CreatePostCommentInput) (*entities.PostComment, error) {
	post, err := s.repo.FindByID(postID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if post == nil {
		return nil, exceptions.NotFound("")
	}

	comment := &models.PostComment{
		ID:       uuid.New(),
		PostID:   postID,
		AuthorID: authorID,
		Body:     input.Body,
	}
	if err := s.repo.CreateComment(comment); err != nil {
		return nil, exceptions.Internal("")
	}

	created, err := s.repo.FindCommentByID(comment.ID)
	if err != nil || created == nil {
		return nil, exceptions.Internal("")
	}

	return toPostCommentPtr(*created), nil
}

func (s *PostService) ListComments(postID uuid.UUID, page, limit int) (*entities.PostCommentList, error) {
	page, limit = normalizePage(page, limit)

	post, err := s.repo.FindByID(postID)
	if err != nil {
		return nil, exceptions.Internal("")
	}
	if post == nil {
		return nil, exceptions.NotFound("")
	}

	comments, total, err := s.repo.ListComments(postID, page, limit)
	if err != nil {
		return nil, exceptions.Internal("")
	}

	data := make([]entities.PostComment, 0, len(comments))
	for _, c := range comments {
		data = append(data, toPostComment(c))
	}

	return &entities.PostCommentList{Data: data, Meta: entities.NewPageMeta(page, limit, total)}, nil
}

// toSummariesWithViewer batch-resolves the viewer's liked state across a
// page of posts in one query instead of one IsLiked call per post.
func (s *PostService) toSummariesWithViewer(posts []models.Post, viewerID *uuid.UUID) ([]entities.PostSummary, error) {
	var likedMap map[uuid.UUID]bool
	if viewerID != nil {
		ids := make([]uuid.UUID, len(posts))
		for i, p := range posts {
			ids[i] = p.ID
		}
		var err error
		likedMap, err = s.repo.LikedPostIDs(*viewerID, ids)
		if err != nil {
			return nil, err
		}
	}

	summaries := make([]entities.PostSummary, 0, len(posts))
	for _, p := range posts {
		var liked *bool
		if likedMap != nil {
			v := likedMap[p.ID]
			liked = &v
		}
		summaries = append(summaries, *toPost(p, liked))
	}
	return summaries, nil
}

func toPost(p models.Post, liked *bool) *entities.Post {
	return &entities.Post{
		ID:           p.ID.String(),
		Author:       entities.PublicUser{Name: p.Author.Name, Username: usernameOf(p.Author), Avatar: p.Author.AvatarPath},
		Body:         p.Body,
		Image:        p.ImagePath,
		CreatedAt:    p.CreatedAt,
		LikeCount:    p.LikeCount,
		CommentCount: p.CommentCount,
		Liked:        liked,
	}
}

func toPostComment(c models.PostComment) entities.PostComment {
	return entities.PostComment{
		ID:        c.ID.String(),
		PostID:    c.PostID.String(),
		Author:    entities.PublicUser{Name: c.Author.Name, Username: usernameOf(c.Author), Avatar: c.Author.AvatarPath},
		Body:      c.Body,
		CreatedAt: c.CreatedAt,
	}
}

func toPostCommentPtr(c models.PostComment) *entities.PostComment {
	comment := toPostComment(c)
	return &comment
}
