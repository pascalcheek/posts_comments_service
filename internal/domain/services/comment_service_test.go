package services_test

import (
	"github.com/stretchr/testify/require"
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"posts_comments_service/internal/domain/services"
	"posts_comments_service/internal/repository/memory"
)

func TestAddComment_Success(t *testing.T) {
	repo := memory.NewCommentRepository()
	service := services.NewCommentService(repo)

	// Добавляем пост вручную в память
	post := &models.Post{
		ID:            "post-123",
		Title:         "Test Post",
		Content:       "Test Content",
		Author:        "Tester",
		AllowComments: true,
		CreatedAt:     "2024-01-01T00:00:00Z",
	}
	repo.SetPost(post)

	comment, err := service.AddComment(post.ID, "Author", "This is a comment", nil)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, "Author", comment.Author)
	assert.Equal(t, "This is a comment", comment.Text)
}

func TestAddComment_TooLong(t *testing.T) {
	memRepo := memory.NewCommentRepository()
	service := services.NewCommentService(memRepo)

	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}
	comment, err := service.AddComment("post1", "Bob", string(longText), nil)
	assert.Error(t, err)
	assert.Nil(t, comment)
}

func TestGetComments(t *testing.T) {
	memRepo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-id",
		Title:         "Title",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	memRepo.SetPost(post)
	service := services.NewCommentService(memRepo)

	// Добавим 2 комментария
	_, err := service.AddComment(post.ID, "user1", "comment 1", nil)
	require.NoError(t, err)
	_, err = service.AddComment(post.ID, "user2", "comment 2", nil)
	require.NoError(t, err)

	comments, hasMore, err := service.GetComments(post.ID, nil, 10, nil, "ASC")
	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.False(t, hasMore)
}

func TestAddComment_DisabledComments(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-id",
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: false,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	_, err := service.AddComment(post.ID, "author", "text", nil)
	assert.ErrorIs(t, err, repositories.ErrCommentsDisabled)
}

func TestAddComment_InvalidParent(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-id",
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	parentID := "non-existent-id"
	_, err := service.AddComment(post.ID, "author", "text", &parentID)
	assert.ErrorIs(t, err, repositories.ErrParentNotFound)
}

func TestGetComments_WithPagination(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-id",
		Title:         "Test Post",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	_, _ = service.AddComment(post.ID, "author1", "text1", nil)
	_, _ = service.AddComment(post.ID, "author2", "text2", nil)
	_, _ = service.AddComment(post.ID, "author3", "text3", nil)

	comments, hasMore, err := service.GetComments(post.ID, nil, 2, nil, "ASC")
	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.True(t, hasMore)

	comments2, hasMore2, err := service.GetComments(post.ID, nil, 2, &comments[1].ID, "ASC")
	require.NoError(t, err)
	assert.Len(t, comments2, 1)
	assert.False(t, hasMore2)
}

func TestAddNestedComment_Success(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-nested",
		Title:         "Nested",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	parent, err := service.AddComment(post.ID, "user1", "parent comment", nil)
	require.NoError(t, err)
	assert.NotNil(t, parent)

	child, err := service.AddComment(post.ID, "user2", "child comment", &parent.ID)
	require.NoError(t, err)
	assert.NotNil(t, child)

	comments, _, err := service.GetComments(post.ID, &parent.ID, 10, nil, "ASC")
	require.NoError(t, err)
	require.Len(t, comments, 1)
	assert.Equal(t, "child comment", comments[0].Text)
	assert.Equal(t, parent.ID, *comments[0].ParentID)
}

func TestGetComments_LimitZero(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-limit-zero",
		Title:         "LimitZero",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	_, _ = service.AddComment(post.ID, "user1", "comment1", nil)
	_, _ = service.AddComment(post.ID, "user2", "comment2", nil)

	comments, hasMore, err := service.GetComments(post.ID, nil, 0, nil, "ASC")
	require.NoError(t, err)
	assert.Len(t, comments, 0)
	assert.False(t, hasMore)
}

func TestGetComments_InvalidAfterCursor(t *testing.T) {
	repo := memory.NewCommentRepository()
	post := &models.Post{
		ID:            "post-invalid-after",
		Title:         "InvalidAfter",
		Content:       "Content",
		Author:        "Author",
		AllowComments: true,
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	repo.SetPost(post)
	service := services.NewCommentService(repo)

	_, _ = service.AddComment(post.ID, "user1", "comment1", nil)

	invalidCursor := "non-existent-id"
	_, _, err := service.GetComments(post.ID, nil, 10, &invalidCursor, "ASC")
	assert.ErrorIs(t, err, repositories.ErrInvalidCursor)
}
