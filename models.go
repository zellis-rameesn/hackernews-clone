package main

import (
	"time"
)

type User struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	Profile        Profile   `json:"profile"`
}

type Profile struct {
	UserId    int    `json:"user_id"`
	Avatar    string `json:"avatar"`
	CreatedAt string `json:"created_at"`
}

type Post struct {
	Id           int       `json:"id"`
	Url          string    `json:"url"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UserID       int       `json:"user_id"`
	UserName     string    `json:"user_name"`
	CommentCount int       `json:"comment_count"`
	VoteCount    int       `json:"vote_count"`
	TotalRecords int       `json:"total_records"`
}

type Comment struct {
	Id        int       `json:"id"`
	Body      string    `json:"body"`
	UserID    int       `json:"user_id"`
	UserName  string    `json:"user_name"`
	PostID    int       `json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
