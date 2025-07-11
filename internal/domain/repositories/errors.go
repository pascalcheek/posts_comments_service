package repositories

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrInvalidCursor    = errors.New("invalid cursor")
	ErrCommentsDisabled = errors.New("comments are disabled for this post")
	ErrTextTooLong      = errors.New("comment text exceeds the 2000 character limit")
	ErrParentNotFound   = errors.New("parent comment not found")
)
