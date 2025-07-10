package services_test

import (
	"github.com/stretchr/testify/require"
	"posts_comments_service/internal/domain/repositories"
	"testing"

	"github.com/stretchr/testify/assert"
	"posts_comments_service/internal/domain/services"
	"posts_comments_service/internal/repository/memory"
)

func TestCreatePost_Success(t *testing.T) {
	memRepo := memory.NewPostRepository()
	service := services.NewPostService(memRepo)

	post, err := service.CreatePost("Title", "Content", "Author", true)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, "Title", post.Title)
}

func TestGetPost_Success(t *testing.T) {
	memRepo := memory.NewPostRepository()
	service := services.NewPostService(memRepo)

	created, _ := service.CreatePost("Title", "Content", "Author", true)

	post, err := service.GetPost(created.ID)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, created.ID, post.ID)
}

func TestGetPosts(t *testing.T) {
	memRepo := memory.NewPostRepository()
	service := services.NewPostService(memRepo)

	_, _ = service.CreatePost("Title 1", "Content", "Author", true)
	_, _ = service.CreatePost("Title 2", "Content", "Author", true)

	posts, err := service.GetPosts(10, nil, "DESC")
	assert.NoError(t, err)
	assert.Len(t, posts, 2)
}

func TestCreatePost_InvalidAuthor(t *testing.T) {
	repo := memory.NewPostRepository()
	service := services.NewPostService(repo)

	post, err := service.CreatePost("Title", "Content", "", true)
	require.NoError(t, err)
	assert.Equal(t, "", post.Author)
}

func TestGetPost_NotFound(t *testing.T) {
	repo := memory.NewPostRepository()
	service := services.NewPostService(repo)

	post, err := service.GetPost("non-existent-id")
	assert.Nil(t, post)
	assert.ErrorIs(t, err, repositories.ErrNotFound)
}

func TestGetPosts_SortOrderValidation(t *testing.T) {
	repo := memory.NewPostRepository()
	service := services.NewPostService(repo)

	_, err := service.GetPosts(10, nil, "INVALID")
	assert.Error(t, err)
}
