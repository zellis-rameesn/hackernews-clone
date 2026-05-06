package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/dromara/carbon/v2"
)

var (
	ErrDuplicatePostTitle = errors.New("Duplicate post title")
	ErrDuplicateVote      = errors.New("Duplicate vote")
)

type PostRepository interface {
	CreatePost(url, title string, userID int)
	AddComment(body string, userID, postID int)
	GetComments(postID int) ([]Comment, error)
	AddVote(userID, postID int)
	GetAll(filter Filter) ([]Post, Metadata, error)
	GetByID(id int)
}

type Filter struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"`
	Query    string `json:"query"`
}

func (f Filter) Validate() error {
	if f.PageSize <= 0 || f.Page >= 100 {
		return errors.New("Invalid page range: 1 to 100 max")
	}
	return nil
}

type Metadata struct {
	CurrentPage  int `json:"current_page"`
	PageSize     int `json:"page_size"`
	FirstPage    int `json:"first_page"`
	NextPage     int `json:"next_page"`
	PrevPage     int `json:"prev_page"`
	LastPage     int `json:"last_page"`
	TotalRecords int `json:"total_records"`
}

func CalculateMetadata(totalRecords, currentPage, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	meta := Metadata{
		TotalRecords: totalRecords,
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(int(math.Ceil(float64(totalRecords) / float64(pageSize)))),
		NextPage:     currentPage + 1,
		PrevPage:     currentPage - 1,
	}
	if meta.CurrentPage >= meta.LastPage {
		meta.NextPage = 0
	}
	if meta.CurrentPage <= meta.FirstPage {
		meta.PrevPage = 0
	}

	return meta
}

type PostRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

// SQL query writing order		Database processing order
// SELECT ...					FROM ...
// FROM ...						JOIN ...
// JOIN ...						WHERE ...
// WHERE ...					GROUP BY ...
// GROUP BY ...					HAVING ...
// HAVING ...					WINDOW (OVER ...)
// WINDOW (OVER ...)			SELECT ...
// ORDER BY ...					ORDER BY ...
// LIMIT ...					LIMIT ...

func (p *PostRepo) CreatePost(url, title string, userID int) (int, error) {
	stmt, err := p.db.Exec(`insert into posts (url, title, user_id) values (?, ?, ?)`, url, title, userID)

	if err != nil {
		// this is a substring of error returned when sql query fails when there's a duplicate
		if strings.Contains(err.Error(), "UNIQUE constraint failed: posts.title") {
			return 0, ErrDuplicatePostTitle
		}
		return 0, err
	}
	postID, err := stmt.LastInsertId()

	if err != nil {
		return 0, err
	}
	return int(postID), nil
}

func (p *PostRepo) AddComment(body string, userID, postID int) (int, error) {
	stmt, err := p.db.Exec(`insert into comments (body, user_id, post_id) values (?, ?, ?)`, body, userID, postID)

	if err != nil {
		return 0, err
	}
	commentID, err := stmt.LastInsertId()

	if err != nil {
		return 0, err
	}
	return int(commentID), nil
}

func (p *PostRepo) AddVote(userID, postID int) (int, error) {
	stmt, err := p.db.Exec(`insert into votes (user_id, post_id) values (?, ?)`, userID, postID)
	fmt.Println("VOTE ERR", err)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") ||
			strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return 0, ErrDuplicateVote
		}
		return 0, err
	}
	voteID, err := stmt.LastInsertId()

	if err != nil {
		return 0, err
	}
	return int(voteID), nil
}

func (p *PostRepo) GetByID(id int) (*Post, error) {
	query := `
	SELECT p.id, p.url, p.title, p.created_at, p.user_id,
	u.name as user_name,
	COUNT(DISTINCT c.id) as comment_count,
	COUNT(DISTINCT v.user_id) as vote_count
	from posts p
	LEFT JOIN users u on p.user_id = u.id
	LEFT JOIN comments c on p.id = c.post_id
	LEFT JOIN votes v on p.id = v.post_id
	WHERE p.id = ?
	GROUP BY p.id, p.url, p.title, p.created_at, p.user_id, u.name
`
	row := p.db.QueryRow(query, id)

	var post Post
	err := row.Scan(&post.Id, &post.Url, &post.Title, &post.CreatedAt, &post.UserID, &post.UserName, &post.CommentCount, &post.VoteCount)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (p *PostRepo) GetAll(filter Filter) ([]Post, Metadata, error) {
	if err := filter.Validate(); err != nil {
		return nil, Metadata{}, err
	}
	// COUNT(*) OVER() is a window function, so total_records will be added for each rows
	baseQuery := `
		SELECT
		COUNT(*) OVER() as total_records,
		p.id, p.url, p.title, p.created_at, p.user_id,
		u.name as user_name,
		COUNT(DISTINCT c.id) as comment_count,
		COUNT(DISTINCT v.user_id) as vote_count
		from posts p
		LEFT JOIN users u on p.user_id = u.id
		LEFT JOIN comments c on p.id = c.post_id
		LEFT JOIN votes v on p.id = v.post_id`

	var args []interface{}
	if filter.Query != "" {
		baseQuery += " WHERE LOWER(p.title) LIKE ?"
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
	}

	baseQuery += " GROUP BY p.id, p.title, p.url, p.user_id, p.created_at, u.name"

	if filter.OrderBy == "popular" {
		baseQuery += " ORDER BY vote_count DESC, p.created_at DESC" // created_at is used since if vote count is a tie then this can be used
	} else {
		baseQuery += " ORDER BY p.created_at DESC"
	}

	limit := filter.PageSize
	offset := (filter.Page - 1) * limit
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := p.db.Query(baseQuery, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var posts []Post
	var totalRecords int
	for rows.Next() {
		var post Post

		err = rows.Scan(&totalRecords, &post.Id, &post.Url, &post.Title, &post.CreatedAt, &post.UserID, &post.UserName, &post.CommentCount, &post.VoteCount)
		if err != nil {
			return nil, Metadata{}, err
		}
		post.TotalRecords = totalRecords
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	if len(posts) == 0 {
		return []Post{}, Metadata{}, err
	}
	metadata := CalculateMetadata(totalRecords, filter.Page, limit)

	return posts, metadata, nil
}

func (p *PostRepo) GetComments(postID int) ([]Comment, error) {
	query := `
		SELECT c.id, c.body, c.user_id, c.post_id, c.created_at, u.name as user_name
		from comments c
		LEFT JOIN users u on c.user_id = u.id
		WHERE post_id = ?
		ORDER BY c.created_at DESC
	`

	rows, err := p.db.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []Comment
	for rows.Next() {
		var comment Comment
		err = rows.Scan(&comment.Id, &comment.Body, &comment.UserID, &comment.PostID, &comment.CreatedAt, &comment.UserName)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return []Comment{}, nil
	}
	return comments, nil
}

// this function is called from the template, since the struct Post doesn't have a Host field
// go allows the template to excute the function attached to the struct
func (p *Post) Host() string {
	ur, err := url.Parse(p.Url)
	if err != nil {
		return "<invalid-host>"
	}
	return ur.Hostname()
}

func (p *Post) GetVoteCounts() string {
	if p.VoteCount > 1 {
		return fmt.Sprintf("%d votes", p.VoteCount)
	} else {
		return fmt.Sprintf("%d vote", p.VoteCount)
	}
}

func (p *Post) GetCommentCounts() string {
	if p.CommentCount > 1 {
		return fmt.Sprintf("%d comments", p.CommentCount)
	} else {
		return fmt.Sprintf("%d comment", p.CommentCount)
	}
}

func (p *Post) GetCreatedAt() string {
	return carbon.NewCarbon(p.CreatedAt).DiffForHumans()
}
