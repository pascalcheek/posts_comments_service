package model

type Comment struct {
	ID           string  `json:"id"`
	PostID       string  `json:"postId"`
	ParentID     *string `json:"parentId,omitempty"`
	Text         string  `json:"text"`
	Author       string  `json:"author"`
	CreatedAt    string  `json:"createdAt"`
	RepliesCount int     `json:"repliesCount"`
}
