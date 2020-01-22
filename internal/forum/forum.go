package forum

import (
	"github.com/jackc/pgx"
)

type ForumService struct {
	db *pgx.ConnPool
}

func NewForumService(db *pgx.ConnPool) *ForumService {
	return &ForumService{db: db}
}

func (fs *ForumService) InsertForum(forum Forum) (err error) {
	sqlQuery := `INSERT INTO forum.forum (slug, title, "user")
	VALUES ($1,$2,$3)`
	_, err = fs.db.Exec(sqlQuery, forum.Slug, forum.Title, forum.User)
	return
}

func (fs *ForumService) SelectForumBySlug(slug string) (forum Forum, err error) {
	sqlQuery := `
	SELECT f.slug, f.title, f."user", f.threads, f.posts
	FROM forum.forum as f
	where lower(f.slug) = lower($1)`
	err = fs.db.QueryRow(sqlQuery, slug).Scan(&forum.Slug, &forum.Title, &forum.User, &forum.Threads, &forum.Posts)
	return
}

func (fs *ForumService) UpdateThreadCount(forum string) (err error) {
	sqlQuery := `
	UPDATE forum.forum SET threads = threads + 1
	WHERE Lower(forum.slug) = Lower($1)`
	_, err = fs.db.Exec(sqlQuery, forum)
	return
}

func (fs *ForumService) UpdatePostCount(forum string, count int) (err error) {
	sqlQuery := `
	UPDATE forum.forum SET posts = posts + $2
	WHERE Lower(forum.slug) = Lower($1)`
	_, err = fs.db.Exec(sqlQuery, forum, count)
	return
}

func (fs *ForumService) InsertForumUser(forum, userNickName string) (err error) {
	sqlQuery := `
	INSERT INTO forum.forum_user (forum, "user")
	VALUES ($1,$2)`
	_, err = fs.db.Exec(sqlQuery, forum, userNickName)
	return
}

