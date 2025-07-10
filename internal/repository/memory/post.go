package memory

import (
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
	"sync"
)

type postRepository struct {
	mu          sync.RWMutex
	posts       []*models.Post
	postsById   map[string]*models.Post
	postIndices map[string]int
}

func NewPostRepository() repositories.PostRepository {
	return &postRepository{
		posts:       make([]*models.Post, 0),
		postsById:   make(map[string]*models.Post),
		postIndices: make(map[string]int),
	}
}

func (r *postRepository) Create(post *models.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts = append(r.posts, post)
	r.postsById[post.ID] = post
	r.postIndices[post.ID] = len(r.posts) - 1
	return nil
}

func (r *postRepository) GetByID(id string) (*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, ok := r.postsById[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return post, nil
}

func (r *postRepository) List(limit int, after *string, sortOrder string) ([]*models.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.posts)
	if total == 0 {
		return []*models.Post{}, nil
	}

	// Для DESC порядка (новые сначала)
	if sortOrder == "DESC" || sortOrder == "" {
		startIdx := total - 1
		if after != nil {
			post, exists := r.postsById[*after]
			if !exists {
				return nil, repositories.ErrInvalidCursor
			}
			startIdx = r.postIndices[post.ID] - 1
		}

		available := startIdx + 1
		if available > limit {
			available = limit
		}

		result := make([]*models.Post, 0, available)
		for i := 0; i < available; i++ {
			result = append(result, r.posts[startIdx-i])
		}
		return result, nil
	}

	// Для ASC порядка (старые сначала)
	start := 0
	if after != nil {
		post, exists := r.postsById[*after]
		if !exists {
			return nil, repositories.ErrInvalidCursor
		}
		start = r.postIndices[post.ID] + 1
	}

	end := start + limit
	if end > total {
		end = total
	}

	return r.posts[start:end], nil
}
