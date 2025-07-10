package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
)

type postRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) repositories.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(post *models.Post) error {
	_, err := r.db.Exec(`
        INSERT INTO posts (id, title, content, author, allow_comments, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)`,
		post.ID, post.Title, post.Content, post.Author, post.AllowComments, post.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func (r *postRepository) GetByID(id string) (*models.Post, error) {
	postUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, repositories.ErrNotFound
	}

	row := r.db.QueryRow(`
        SELECT id, title, content, author, allow_comments, created_at
        FROM posts WHERE id = $1`, postUUID)

	var post models.Post
	var dbUUID uuid.UUID
	var createdAt time.Time

	if err := row.Scan(&dbUUID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	post.ID = dbUUID.String()
	post.CreatedAt = createdAt.Format(time.RFC3339)
	return &post, nil
}

func (r *postRepository) List(limit int, after *string, sortOrder string) ([]*models.Post, error) {
	var query string
	var afterTime *time.Time

	if after != nil {
		id, err := uuid.Parse(*after)
		if err == nil {
			row := r.db.QueryRow("SELECT created_at FROM posts WHERE id = $1", id)
			var t time.Time
			if err := row.Scan(&t); err == nil {
				afterTime = &t
			} else if err == sql.ErrNoRows {
				return nil, repositories.ErrInvalidCursor
			}
		}
	}

	if sortOrder == "ASC" {
		query = `
            SELECT id, title, content, author, allow_comments, created_at
            FROM posts
            WHERE ($1::timestamptz IS NULL OR created_at > $1)
            ORDER BY created_at ASC
            LIMIT $2`
	} else {
		query = `
            SELECT id, title, content, author, allow_comments, created_at
            FROM posts
            WHERE ($1::timestamptz IS NULL OR created_at < $1)
            ORDER BY created_at DESC
            LIMIT $2`
	}

	rows, err := r.db.Query(query, afterTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		var post models.Post
		var dbUUID uuid.UUID
		var createdAt time.Time

		if err := rows.Scan(&dbUUID, &post.Title, &post.Content, &post.Author, &post.AllowComments, &createdAt); err != nil {
			return nil, err
		}

		post.ID = dbUUID.String()
		post.CreatedAt = createdAt.Format(time.RFC3339)
		posts = append(posts, &post)
	}

	return posts, nil
}
