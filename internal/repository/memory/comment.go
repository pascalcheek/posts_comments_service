package memory

import (
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
	"sync"
)

type commentRepository struct {
	mu           sync.RWMutex
	comments     map[string]*models.Comment
	commentsTree map[string]*commentLevel
	postRepo     repositories.PostRepository
}

type commentLevel struct {
	comments []*models.Comment
	indexMap map[string]int
}

func NewCommentRepository(postRepo repositories.PostRepository) repositories.CommentRepository {
	return &commentRepository{
		comments:     make(map[string]*models.Comment),
		commentsTree: make(map[string]*commentLevel),
		postRepo:     postRepo,
	}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(comment.Text) > 2000 {
		return repositories.ErrTextTooLong
	}

	// Проверяем существование поста через postRepo
	post, err := r.postRepo.GetByID(comment.PostID)
	if err != nil {
		return repositories.ErrNotFound
	}

	if !post.AllowComments {
		return repositories.ErrCommentsDisabled
	}

	if comment.ParentID != nil {
		if parent, exists := r.comments[*comment.ParentID]; !exists || parent.PostID != comment.PostID {
			return repositories.ErrParentNotFound
		}
	}

	levelKey := comment.PostID
	if comment.ParentID != nil {
		levelKey = *comment.ParentID
	}

	if _, exists := r.commentsTree[levelKey]; !exists {
		r.commentsTree[levelKey] = &commentLevel{
			comments: make([]*models.Comment, 0),
			indexMap: make(map[string]int),
		}
	}

	level := r.commentsTree[levelKey]
	level.indexMap[comment.ID] = len(level.comments)
	level.comments = append(level.comments, comment)
	r.comments[comment.ID] = comment

	return nil
}

func (r *commentRepository) GetByPostID(postID string, parentID *string, limit int, after *string, sortOrder string) ([]*models.Comment, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if limit <= 0 {
		return []*models.Comment{}, false, nil
	}

	levelKey := postID
	if parentID != nil {
		levelKey = *parentID
	}

	level, exists := r.commentsTree[levelKey]
	if !exists {
		return nil, false, nil
	}

	total := len(level.comments)
	if total == 0 {
		return nil, false, nil
	}

	if sortOrder == "DESC" {
		start := total - 1
		if after != nil {
			if idx, ok := level.indexMap[*after]; ok {
				start = idx - 1
			} else {
				return nil, false, repositories.ErrInvalidCursor
			}
		}

		end := start - limit + 1
		if end < 0 {
			end = -1
		}

		result := make([]*models.Comment, 0, limit)
		for i := start; i > end; i-- {
			result = append(result, level.comments[i])
			if len(result) >= limit {
				break
			}
		}
		return result, start-limit >= 0, nil
	}

	// ASC order
	start := 0
	if after != nil {
		if idx, ok := level.indexMap[*after]; ok {
			start = idx + 1
		} else {
			return nil, false, repositories.ErrInvalidCursor
		}
	}

	end := start + limit
	if end > total {
		end = total
	}

	return level.comments[start:end], end < total, nil
}

func (r *commentRepository) Count(postID string, parentID *string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	levelKey := postID
	if parentID != nil {
		if _, exists := r.comments[*parentID]; !exists {
			return 0, repositories.ErrParentNotFound
		}
		levelKey = *parentID
	}

	level, exists := r.commentsTree[levelKey]
	if !exists {
		return 0, nil
	}

	return len(level.comments), nil
}
