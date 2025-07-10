package model

type Post struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	Author        string `json:"author"`
	AllowComments bool   `json:"allowComments"`
	CreatedAt     string `json:"createdAt"`
}
