package forum

import (
	"github.com/jackc/pgx"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type PostService struct {
	db *pgx.ConnPool
}

func NewPostService(db *pgx.ConnPool) *PostService {
	return &PostService{db: db}
}

func (ps *PostService) CreatePosts(thread Thread, created string, posts []Post) (post []Post, err error) {
	sqlQuery := "INSERT INTO forum.post(id, parent, thread, forum, author, created, message, path) VALUES "
	vals := []interface{}{}
	for _, post := range posts {
		var author string
		err = ps.db.QueryRow(`SELECT "user".nick_name FROM forum."user" WHERE LOWER("user".nick_name) = LOWER($1)`,
			post.Author,
		).Scan(&author)
		if err != nil {
			return nil, errors.New("404")
		}
		tempSqlQuery := `
		INSERT INTO forum.forum_user (forum, "user")
		VALUES ($1,$2)`
		_, _ = ps.db.Exec(tempSqlQuery, thread.Forum, author)

		if post.Parent == 0 {
			sqlQuery += "(nextval('forum.post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"ARRAY[currval(pg_get_serial_sequence('forum.post', 'id'))::bigint]),"
			vals = append(vals, post.Parent, thread.Id, thread.Forum, post.Author, created, post.Message)
		} else {
			var parentThreadId int32
			err = ps.db.QueryRow("SELECT post.thread FROM forum.post WHERE post.id = $1",
				post.Parent,
			).Scan(&parentThreadId)
			if err != nil {
				return nil, err
			}
			if parentThreadId != int32(thread.Id) {
				return nil, errors.New("Parent post was created in another thread")
			}

			sqlQuery += " (nextval('forum.post_id_seq'::regclass), ?, ?, ?, ?, ?, ?, " +
				"(SELECT post.path FROM forum.post WHERE post.id = ? AND post.thread = ?) || " +
				"currval(pg_get_serial_sequence('forum.post', 'id'))::bigint),"

			vals = append(vals, post.Parent, thread.Id, thread.Forum, post.Author, created, post.Message, post.Parent, thread.Id)
		}

	}
	sqlQuery = strings.TrimSuffix(sqlQuery, ",")

	sqlQuery += " RETURNING id, parent, thread, forum, author, created, message, is_edited "

	sqlQuery = ReplaceSQL(sqlQuery, "?")
	if len(posts) > 0 {
		rows, err := ps.db.Query(sqlQuery, vals...)
		if err != nil {
			return nil, err
		}
		scanPost := Post{}
		for rows.Next() {
			err := rows.Scan(
				&scanPost.Id,
				&scanPost.Parent,
				&scanPost.Thread,
				&scanPost.Forum,
				&scanPost.Author,
				&scanPost.Created,
				&scanPost.Message,
				&scanPost.IsEdited,
			)
			if err != nil {
				rows.Close()
				return nil, err
			}
			post = append(post, scanPost)
		}
		rows.Close()
	}
	return post, nil
}

func ReplaceSQL(old, searchPattern string) string {
	tmpCount := strings.Count(old, searchPattern)
	for m := 1; m <= tmpCount; m++ {
		old = strings.Replace(old, searchPattern, "$"+strconv.Itoa(m), 1)
	}
	return old
}
func (ps *PostService) SelectPostById(id int) (post Post, err error) {
	sqlQuery := `
	SELECT p.author, p.created, p.forum, p.id, p.is_edited, p.message, p.parent, p.thread
	FROM forum.post as p
	where p.id = $1`
	err = ps.db.QueryRow(sqlQuery, id).Scan(&post.Author, &post.Created, &post.Forum, &post.Id, &post.IsEdited, &post.Message, &post.Parent, &post.Thread)
	return
}

func (ps *PostService) UpdatePostMessage(newMessage string, id int) (countUpdateString int64, err error) {
	sqlQuery := `UPDATE forum.post SET message = $1,
                       is_edited = true
	where post.id = $2`
	result, err := ps.db.Exec(sqlQuery, newMessage, id)
	if err != nil {
		return
	}
	countUpdateString = result.RowsAffected()
	return
}
