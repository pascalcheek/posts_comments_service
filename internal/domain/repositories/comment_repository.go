package repositories

import "posts_comments_service/internal/domain/models"

type CommentRepository interface {
	Create(comment *models.Comment) error
	GetByPostID(postID string, parentID *string, limit int, after *string, sortOrder string) ([]*models.Comment, bool, error)
	Count(postID string, parentID *string) (int, error)
	CountReplies(postID string) (map[string]int, error)
}
