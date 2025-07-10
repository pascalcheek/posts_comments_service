package graphql

import (
	"context"
	"posts_comments_service/internal/delivery/graphql/generated"
	"posts_comments_service/internal/delivery/graphql/model"
	"posts_comments_service/internal/domain/models"
	"posts_comments_service/internal/domain/services"
)

type Resolver struct {
	postService    *services.PostService
	commentService *services.CommentService
}

func NewResolver(postService *services.PostService, commentService *services.CommentService) *Resolver {
	return &Resolver{
		postService:    postService,
		commentService: commentService,
	}
}

// Mutation resolvers
func (r *mutationResolver) CreatePost(
	ctx context.Context,
	title string,
	content string,
	author string,
	allowComments bool,
) (*model.Post, error) {
	domainPost, err := r.postService.CreatePost(title, content, author, allowComments)
	if err != nil {
		return nil, err
	}
	return convertDomainPostToModel(domainPost), nil
}

func (r *mutationResolver) CreateComment(
	ctx context.Context,
	postID string,
	parentID *string,
	text string,
	author string,
) (*model.Comment, error) {
	domainComment, err := r.commentService.AddComment(postID, author, text, parentID)
	if err != nil {
		return nil, err
	}
	return convertDomainCommentToModel(domainComment), nil
}

// Query resolvers
func (r *queryResolver) Posts(
	ctx context.Context,
	after *string,
	first *int,
	sortOrder *model.SortOrder,
) ([]*model.Post, error) {
	limit := 10
	if first != nil {
		limit = *first
	}

	order := "DESC"
	if sortOrder != nil {
		order = string(*sortOrder)
	}

	domainPosts, err := r.postService.GetPosts(limit, after, order)
	if err != nil {
		return nil, err
	}
	return convertDomainPostsToModel(domainPosts), nil
}

func (r *queryResolver) Post(
	ctx context.Context,
	id string,
) (*model.Post, error) {
	domainPost, err := r.postService.GetPost(id)
	if err != nil {
		return nil, err
	}
	return convertDomainPostToModel(domainPost), nil
}

func (r *queryResolver) Comments(
	ctx context.Context,
	postID string,
	parentID *string,
	after *string,
	first *int,
	sortOrder *model.SortOrder,
) (*model.CommentConnection, error) {
	limit := 10
	if first != nil {
		limit = *first
	}

	order := "ASC"
	if sortOrder != nil {
		order = string(*sortOrder)
	}

	domainComments, hasMore, err := r.commentService.GetComments(postID, parentID, limit, after, order)
	if err != nil {
		return nil, err
	}

	count, err := r.commentService.GetCommentsCount(postID, parentID)
	if err != nil {
		return nil, err
	}

	return &model.CommentConnection{
		Edges:      convertToCommentEdges(domainComments),
		PageInfo:   generatePageInfo(hasMore, domainComments),
		TotalCount: count,
	}, nil
}

func (r *queryResolver) CommentsCount(
	ctx context.Context,
	postID string,
	parentID *string,
) (int, error) {
	return r.commentService.GetCommentsCount(postID, parentID)
}

// Helper functions
func convertDomainPostToModel(post *models.Post) *model.Post {
	return &model.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		Author:        post.Author,
		AllowComments: post.AllowComments,
		CreatedAt:     post.CreatedAt,
	}
}

func convertDomainPostsToModel(posts []*models.Post) []*model.Post {
	result := make([]*model.Post, len(posts))
	for i, post := range posts {
		result[i] = convertDomainPostToModel(post)
	}
	return result
}

func convertDomainCommentToModel(comment *models.Comment) *model.Comment {
	return &model.Comment{
		ID:           comment.ID,
		PostID:       comment.PostID,
		ParentID:     comment.ParentID,
		Text:         comment.Text,
		Author:       comment.Author,
		CreatedAt:    comment.CreatedAt,
		RepliesCount: 0, // Можно реализовать подсчет если нужно
	}
}

func convertToCommentEdges(comments []*models.Comment) []*model.CommentEdge {
	edges := make([]*model.CommentEdge, len(comments))
	for i, comment := range comments {
		edges[i] = &model.CommentEdge{
			Node:   convertDomainCommentToModel(comment),
			Cursor: comment.ID,
		}
	}
	return edges
}

func generatePageInfo(hasMore bool, comments []*models.Comment) *model.PageInfo {
	if len(comments) == 0 {
		return &model.PageInfo{
			HasNextPage:     false,
			HasPreviousPage: false,
		}
	}

	lastID := comments[len(comments)-1].ID
	return &model.PageInfo{
		HasNextPage:     hasMore,
		EndCursor:       &lastID,
		HasPreviousPage: false,
	}
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
