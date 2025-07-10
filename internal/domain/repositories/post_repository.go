package repositories

import "posts_comments_service/internal/domain/models"

type PostRepository interface {
	Create(post *models.Post) error
	GetByID(id string) (*models.Post, error)
	List(limit int, after *string, sortOrder string) ([]*models.Post, error)
}
