package graphql

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"
	"posts_comments_service/internal/delivery/graphql/generated"
	"posts_comments_service/internal/delivery/graphql/model"
	"posts_comments_service/internal/domain/constants"
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
func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, author string, allowComments bool) (*model.Post, error) {
	domainPost, err := r.postService.CreatePost(title, content, author, allowComments)
	if err != nil {
		return nil, err
	}
	return convertDomainPostToModel(domainPost), nil
}

// CreateComment is the resolver for the createComment field.
func (r *mutationResolver) CreateComment(ctx context.Context, postID string, parentID *string, text string, author string) (*model.Comment, error) {
	domainComment, err := r.commentService.AddComment(postID, author, text, parentID)
	if err != nil {
		return nil, err
	}
	return convertDomainCommentToModel(domainComment), nil
}

// Query resolvers
func (r *queryResolver) Posts(ctx context.Context, after *string, first *int, sortOrder *model.SortOrder) ([]*model.Post, error) {
	limit := constants.DefaultLimit
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

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*model.Post, error) {
	domainPost, err := r.postService.GetPost(id)
	if err != nil {
		return nil, err
	}
	return convertDomainPostToModel(domainPost), nil
}

// Comments is the resolver for the comments field.
func (r *queryResolver) Comments(ctx context.Context, postID string, parentID *string, after *string, first *int, sortOrder *model.SortOrder) (*model.CommentConnection, error) {
	limit := constants.DefaultLimit
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

// CommentsCount is the resolver for the commentsCount field.
func (r *queryResolver) CommentsCount(ctx context.Context, postID string, parentID *string) (int, error) {
	return r.commentService.GetCommentsCount(postID, parentID)
}

// PostWithComments is the resolver for the postWithComments field.
func (r *queryResolver) PostWithComments(ctx context.Context, postID string, after *string, first *int) (*model.PostWithComments, error) {
	limit := constants.DefaultLimit
	if first != nil {
		limit = *first
	}

	post, err := r.postService.GetPost(postID)
	if err != nil {
		return nil, err
	}

	comments, _, err := r.commentService.GetComments(postID, nil, limit, after, "ASC")
	if err != nil {
		return nil, err
	}

	repliesCountMap, err := r.commentService.GetRepliesCounts(postID)
	if err != nil {
		return nil, err
	}

	total, err := r.commentService.GetCommentsCount(postID, nil)
	if err != nil {
		return nil, err
	}

	return &model.PostWithComments{
		Post:          convertDomainPostToModel(post),
		Comments:      convertDomainCommentsToModelWithReplies(comments, repliesCountMap),
		TotalComments: total,
	}, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

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
		RepliesCount: 0,
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

func convertDomainCommentsToModelWithReplies(comments []*models.Comment, repliesMap map[string]int) []*model.Comment {
	result := make([]*model.Comment, len(comments))
	for i, c := range comments {
		count := repliesMap[c.ID]
		result[i] = &model.Comment{
			ID:           c.ID,
			PostID:       c.PostID,
			ParentID:     c.ParentID,
			Text:         c.Text,
			Author:       c.Author,
			CreatedAt:    c.CreatedAt,
			RepliesCount: count,
		}
	}
	return result
}
