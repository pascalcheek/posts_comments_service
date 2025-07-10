package services_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"posts_comments_service/internal/domain/repositories"
	"posts_comments_service/internal/domain/services"
	"posts_comments_service/internal/repository/memory"
)

func TestAddComment_Success(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	comment, err := commentService.AddComment(post.ID, "CommentAuthor", "Comment text", nil)

	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, post.ID, comment.PostID)
	assert.Equal(t, "CommentAuthor", comment.Author)
	assert.Equal(t, "Comment text", comment.Text)
}

func TestAddComment_TextTooLong(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}

	_, err = commentService.AddComment(post.ID, "Author", string(longText), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "comment text exceeds the 2000 character limit")
}

func TestAddComment_DisabledComments(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("No Comments", "Content", "Author", false)
	require.NoError(t, err)

	_, err = commentService.AddComment(post.ID, "Author", "Text", nil)
	assert.ErrorIs(t, err, repositories.ErrCommentsDisabled)
}

func TestAddComment_ParentNotFound(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	nonExistentParentID := "non-existent-id"
	_, err = commentService.AddComment(post.ID, "Author", "Text", &nonExistentParentID)
	assert.ErrorIs(t, err, repositories.ErrParentNotFound)
}

func TestGetComments(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	_, err = commentService.AddComment(post.ID, "User1", "First comment", nil)
	require.NoError(t, err)
	_, err = commentService.AddComment(post.ID, "User2", "Second comment", nil)
	require.NoError(t, err)

	comments, hasMore, err := commentService.GetComments(post.ID, nil, 10, nil, "ASC")

	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.False(t, hasMore)
	assert.Equal(t, "First comment", comments[0].Text)
	assert.Equal(t, "Second comment", comments[1].Text)
}

func TestGetComments_WithPagination(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	_, err = commentService.AddComment(post.ID, "User1", "Comment 1", nil)
	require.NoError(t, err)
	_, err = commentService.AddComment(post.ID, "User2", "Comment 2", nil)
	require.NoError(t, err)
	_, err = commentService.AddComment(post.ID, "User3", "Comment 3", nil)
	require.NoError(t, err)

	comments, hasMore, err := commentService.GetComments(post.ID, nil, 2, nil, "ASC")
	require.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.True(t, hasMore)

	nextComments, hasMore, err := commentService.GetComments(post.ID, nil, 2, &comments[1].ID, "ASC")
	require.NoError(t, err)
	assert.Len(t, nextComments, 1)
	assert.False(t, hasMore)
}

func TestAddNestedComment(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	parent, err := commentService.AddComment(post.ID, "Parent", "Parent comment", nil)
	require.NoError(t, err)

	child, err := commentService.AddComment(post.ID, "Child", "Child comment", &parent.ID)
	require.NoError(t, err)

	replies, hasMore, err := commentService.GetComments(post.ID, &parent.ID, 10, nil, "ASC")

	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, replies, 1)
	assert.Equal(t, child.ID, replies[0].ID)
	assert.Equal(t, parent.ID, *replies[0].ParentID)
}

func TestGetCommentsCount(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Test Post", "Content", "Author", true)
	require.NoError(t, err)

	_, err = commentService.AddComment(post.ID, "User1", "Comment 1", nil)
	require.NoError(t, err)
	_, err = commentService.AddComment(post.ID, "User2", "Comment 2", nil)
	require.NoError(t, err)

	count, err := commentService.GetCommentsCount(post.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestNestedComments(t *testing.T) {
	postRepo := memory.NewPostRepository()
	commentRepo := memory.NewCommentRepository(postRepo)

	postService := services.NewPostService(postRepo)
	commentService := services.NewCommentService(commentRepo)

	post, err := postService.CreatePost("Nested Comments Test", "Content", "author1", true)
	require.NoError(t, err)
	require.NotNil(t, post)

	rootComment1, err := commentService.AddComment(post.ID, "user1", "Root comment 1", nil)
	require.NoError(t, err)

	rootComment2, err := commentService.AddComment(post.ID, "user2", "Root comment 2", nil)
	require.NoError(t, err)

	child1OfRoot1, err := commentService.AddComment(post.ID, "user3", "Child 1 of Root 1", &rootComment1.ID)
	require.NoError(t, err)

	child2OfRoot1, err := commentService.AddComment(post.ID, "user1", "Child 2 of Root 1", &rootComment1.ID)
	require.NoError(t, err)

	_, err = commentService.AddComment(post.ID, "user4", "Child of Root 2", &rootComment2.ID)
	require.NoError(t, err)

	grandchild1, err := commentService.AddComment(post.ID, "user2", "Grandchild 1", &child1OfRoot1.ID)
	require.NoError(t, err)

	grandchild2, err := commentService.AddComment(post.ID, "user3", "Grandchild 2", &child1OfRoot1.ID)
	require.NoError(t, err)

	rootComments, hasMore, err := commentService.GetComments(post.ID, nil, 10, nil, "ASC")
	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, rootComments, 2)
	assert.Equal(t, rootComment1.ID, rootComments[0].ID)
	assert.Equal(t, rootComment2.ID, rootComments[1].ID)

	childrenOfRoot1, hasMore, err := commentService.GetComments(post.ID, &rootComment1.ID, 10, nil, "ASC")
	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, childrenOfRoot1, 2)
	assert.Equal(t, child1OfRoot1.ID, childrenOfRoot1[0].ID)
	assert.Equal(t, child2OfRoot1.ID, childrenOfRoot1[1].ID)

	grandchildren, hasMore, err := commentService.GetComments(post.ID, &child1OfRoot1.ID, 10, nil, "ASC")
	require.NoError(t, err)
	assert.False(t, hasMore)
	assert.Len(t, grandchildren, 2)
	assert.Equal(t, grandchild1.ID, grandchildren[0].ID)
	assert.Equal(t, grandchild2.ID, grandchildren[1].ID)

	rootCount, err := commentService.GetCommentsCount(post.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, 2, rootCount)

	childrenCount, err := commentService.GetCommentsCount(post.ID, &rootComment1.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, childrenCount)

	grandchildrenCount, err := commentService.GetCommentsCount(post.ID, &child1OfRoot1.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, grandchildrenCount)

	_, err = commentService.AddComment(post.ID, "user1", "Invalid cyclic", &grandchild2.ID)
	require.NoError(t, err)

	assert.Nil(t, rootComment1.ParentID)
	assert.Nil(t, rootComment2.ParentID)
	assert.Equal(t, rootComment1.ID, *child1OfRoot1.ParentID)
	assert.Equal(t, child1OfRoot1.ID, *grandchild1.ParentID)
}
