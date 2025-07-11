package services

import (
	"errors"
	"posts_comments_service/internal/domain/constants"
	"time"

	"github.com/google/uuid"
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
)

type PostService struct {
	repo repositories.PostRepository
}

func NewPostService(repo repositories.PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(title, content, author string, allowComments bool) (*models.Post, error) {
	post := &models.Post{
		ID:            uuid.New().String(),
		Title:         title,
		Content:       content,
		Author:        author,
		AllowComments: allowComments,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetPost(id string) (*models.Post, error) {
	return s.repo.GetByID(id)
}

func (s *PostService) GetPosts(limit int, after *string, sortOrder string) ([]*models.Post, error) {
	if sortOrder != constants.SortAsc && sortOrder != constants.SortDesc {
		return nil, errors.New("invalid sort order")
	}
	return s.repo.List(limit, after, sortOrder)
}
