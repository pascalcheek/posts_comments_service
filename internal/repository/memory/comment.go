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
	postsById    map[string]*models.Post
}

type commentLevel struct {
	comments []*models.Comment
	indexMap map[string]int
}

func NewCommentRepository() repositories.CommentRepository {
	return &commentRepository{
		comments:     make(map[string]*models.Comment),
		commentsTree: make(map[string]*commentLevel),
		postsById:    make(map[string]*models.Post),
	}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(comment.Text) > 2000 {
		return repositories.ErrTextTooLong
	}

	// Проверяем существование поста
	post, exists := r.postsById[comment.PostID]
	if !exists {
		return repositories.ErrNotFound
	}

	// Если это ответ на комментарий - проверяем существование родителя
	if comment.ParentID != nil {
		parentComment, exists := r.comments[*comment.ParentID]
		if !exists {
			return repositories.ErrParentNotFound
		}
		// Убедимся, что родитель принадлежит тому же посту
		if parentComment.PostID != comment.PostID {
			return repositories.ErrWrongParentPost
		}
	}

	if !post.AllowComments {
		return repositories.ErrCommentsDisabled
	}

	// Определяем уровень в иерархии
	levelKey := comment.PostID
	if comment.ParentID != nil {
		levelKey = *comment.ParentID
	}

	// Получаем или создаем уровень
	level, exists := r.commentsTree[levelKey]
	if !exists {
		level = &commentLevel{
			comments: make([]*models.Comment, 0),
			indexMap: make(map[string]int),
		}
		r.commentsTree[levelKey] = level
	}

	// Сохраняем комментарий
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

	// Для DESC (новые сначала)
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

	// Для ASC (старые сначала)
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

	// Определяем ключ уровня в иерархии
	levelKey := postID
	if parentID != nil {
		// Проверяем существование родительского комментария
		if _, exists := r.comments[*parentID]; !exists {
			return 0, repositories.ErrParentNotFound
		}
		levelKey = *parentID
	}

	// Проверяем существование поста для корневых комментариев
	if parentID == nil {
		if _, exists := r.postsById[postID]; !exists {
			return 0, repositories.ErrNotFound
		}
	}

	// Получаем уровень комментариев
	level, exists := r.commentsTree[levelKey]
	if !exists {
		return 0, nil // Уровень существует, но комментариев нет
	}

	// Возвращаем количество комментариев на этом уровне
	return len(level.comments), nil
}

func (r *commentRepository) SetPost(post *models.Post) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.postsById[post.ID] = post
}
