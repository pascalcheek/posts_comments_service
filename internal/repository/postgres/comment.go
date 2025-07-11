package postgres

import (
	"database/sql"
	"posts_comments_service/internal/domain/constants"
	"time"

	"github.com/google/uuid"
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
)

type commentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) repositories.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(comment *models.Comment) error {
	if len(comment.Text) > constants.MaxCommentLength {
		return repositories.ErrTextTooLong
	}

	postUUID, err := uuid.Parse(comment.PostID)
	if err != nil {
		return repositories.ErrNotFound
	}

	var allowComments bool
	err = r.db.QueryRow(`SELECT allow_comments FROM posts WHERE id = $1`, postUUID).Scan(&allowComments)
	if err != nil {
		if err == sql.ErrNoRows {
			return repositories.ErrNotFound
		}
		return err
	}
	if !allowComments {
		return repositories.ErrCommentsDisabled
	}

	var parentUUID *uuid.UUID
	if comment.ParentID != nil {
		if id, err := uuid.Parse(*comment.ParentID); err == nil {
			parentUUID = &id
		} else {
			return repositories.ErrNotFound
		}
	}

	_, err = r.db.Exec(`
        INSERT INTO comments (id, post_id, parent_id, author, text, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)`,
		comment.ID, postUUID, parentUUID, comment.Author, comment.Text, comment.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (r *commentRepository) GetByPostID(postID string, parentID *string, limit int, after *string, sortOrder string) ([]*models.Comment, bool, error) {
	var query string
	var afterTime *time.Time

	if after != nil {
		id, err := uuid.Parse(*after)
		if err == nil {
			row := r.db.QueryRow("SELECT created_at FROM comments WHERE id = $1", id)
			var t time.Time
			if err := row.Scan(&t); err == nil {
				afterTime = &t
			} else if err == sql.ErrNoRows {
				return nil, false, repositories.ErrInvalidCursor
			}
		}
	}

	if sortOrder == constants.SortAsc {
		query = `
            SELECT id, post_id, parent_id, author, text, created_at
            FROM comments
            WHERE post_id = $1 AND (parent_id IS NULL AND $2::uuid IS NULL OR parent_id = $2)
            AND ($3::timestamptz IS NULL OR created_at > $3)
            ORDER BY created_at ASC
            LIMIT $4`
	} else {
		query = `
            SELECT id, post_id, parent_id, author, text, created_at
            FROM comments
            WHERE post_id = $1 AND (parent_id IS NULL AND $2::uuid IS NULL OR parent_id = $2)
            AND ($3::timestamptz IS NULL OR created_at < $3)
            ORDER BY created_at DESC
            LIMIT $4`
	}

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return nil, false, repositories.ErrNotFound
	}

	var parentUUID *uuid.UUID
	if parentID != nil {
		if id, err := uuid.Parse(*parentID); err == nil {
			parentUUID = &id
		} else {
			return nil, false, repositories.ErrNotFound
		}
	}

	rows, err := r.db.Query(query, postUUID, parentUUID, afterTime, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var comments []*models.Comment
	for rows.Next() {
		var comment models.Comment
		var dbUUID uuid.UUID
		var postUUID uuid.UUID
		var parentUUID uuid.NullUUID
		var createdAt time.Time

		if err := rows.Scan(&dbUUID, &postUUID, &parentUUID, &comment.Author, &comment.Text, &createdAt); err != nil {
			return nil, false, err
		}

		comment.ID = dbUUID.String()
		comment.PostID = postUUID.String()
		comment.CreatedAt = createdAt.Format(time.RFC3339)

		if parentUUID.Valid {
			parentStr := parentUUID.UUID.String()
			comment.ParentID = &parentStr
		}

		comments = append(comments, &comment)
	}

	hasMore := false
	if len(comments) > limit {
		hasMore = true
		comments = comments[:limit]
	}

	return comments, hasMore, nil
}

func (r *commentRepository) Count(postID string, parentID *string) (int, error) {
	query := `
        SELECT COUNT(*)
        FROM comments
        WHERE post_id = $1 AND (parent_id IS NULL AND $2::uuid IS NULL OR parent_id = $2)`

	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return 0, repositories.ErrNotFound
	}

	var parentUUID *uuid.UUID
	if parentID != nil {
		if id, err := uuid.Parse(*parentID); err == nil {
			parentUUID = &id
		} else {
			return 0, repositories.ErrNotFound
		}
	}

	var count int
	err = r.db.QueryRow(query, postUUID, parentUUID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
