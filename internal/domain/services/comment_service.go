package services

import (
	"errors"
	"github.com/google/uuid"
	"time"

	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
)

type CommentService struct {
	repo repositories.CommentRepository
}

func NewCommentService(repo repositories.CommentRepository) *CommentService {
	return &CommentService{repo: repo}
}

func (s *CommentService) AddComment(postID, author, text string, parentID *string) (*models.Comment, error) {
	if len(text) > 2000 {
		return nil, errors.New("comment text exceeds the 2000 character limit")
	}

	comment := &models.Comment{
		ID:        uuid.New().String(),
		PostID:    postID,
		ParentID:  parentID,
		Author:    author,
		Text:      text,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) GetComments(postID string, parentID *string, limit int, after *string, sortOrder string) ([]*models.Comment, bool, error) {
	return s.repo.GetByPostID(postID, parentID, limit, after, sortOrder)
}

func (s *CommentService) GetCommentsCount(postID string, parentID *string) (int, error) {
	return s.repo.Count(postID, parentID)
}
