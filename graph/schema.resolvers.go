package graph

import (
	"context"
	"time"

	"github.com/google/uuid"
	"https://github.com/pascalcheek/posts_comments_service/graph/generated"
	"https://github.com/pascalcheek/posts_comments_service/graph/model"
)

func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, allowComments bool) (*model.Post, error) {
	post := &model.Post{
		ID:            uuid.New().String(),
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
	}
	r.posts = append(r.posts, post)
	return post, nil
}

func (r *queryResolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.posts, nil
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
