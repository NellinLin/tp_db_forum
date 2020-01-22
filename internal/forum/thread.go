package forum

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/lib/pq"
)

type ThreadService struct {
	db *pgx.ConnPool
}

func NewThreadService(db *pgx.ConnPool) *ThreadService {
	return &ThreadService{db: db}
}

func (ts *ThreadService) SelectThreadBySlug(threadSlug string) (thread Thread, err error) {
	sqlQuery := `SELECT t.id, t.author, t.created, t.forum, t.message, t.slug, t.title, t.votes
	FROM forum.thread as t 
	where Lower(t.slug) = Lower($1) `
	var slug sql.NullString
	err = ts.db.QueryRow(sqlQuery, threadSlug).Scan(&thread.Id, &thread.Author, &thread.Created, &thread.Forum, &thread.Message, &slug, &thread.Title, &thread.Votes)
	if err != nil {
		return
	}
	if slug.Valid {
		thread.Slug = slug.String
	}
	return
}

func (ts *ThreadService) SelectThreadById(id int) (thread Thread, err error) {
	sqlQuery := `SELECT t.author, t.created, t.id, t.forum, t.message, t.slug, t.title, t.votes
	FROM forum.thread as t 
	where t.id = $1`
	err = ts.db.QueryRow(sqlQuery, id).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err != nil {
		return
	}
	return
}

func (ts *ThreadService) InsertThread(thread Thread) (id int, err error) {
	sqlQuery := `INSERT INTO forum.thread (author, created, message, title, forum, slug)
	VALUES ($1,$2,$3,$4,$5,$6)
	RETURNING id`
	err = ts.db.QueryRow(sqlQuery, thread.Author, thread.Created, thread.Message, thread.Title, thread.Forum, thread.Slug).Scan(&id)
	return
}

func (ts *ThreadService) SelectThreadByForum(forum string, limit int, since string, desc bool) (threads []Thread, err error) {
	var rows *pgx.Rows
	if since == "" && !desc {
		sqlQuery := `
		SELECT t.author, t.created, t.id, t.message, t.slug, t.title, t.votes
		FROM forum.thread as t 
		WHERE lower(t.forum) = lower($1)
		ORDER BY t.created 
		LIMIT $2`
		rows, err = ts.db.Query(sqlQuery, forum, limit)
	} else if since != "" && !desc {
		sqlQuery := `
		SELECT t.author, t.created, t.id, t.message, t.slug, t.title, t.votes
		FROM forum.thread as t 
		WHERE lower(t.forum) = lower($1) AND t.created >= $3
		ORDER BY t.created 
		LIMIT $2`
		rows, err = ts.db.Query(sqlQuery, forum, limit, since)
	} else if since == "" && desc {
		sqlQuery := `
		SELECT t.author, t.created, t.id, t.message, t.slug, t.title, t.votes
		FROM forum.thread as t 
		WHERE lower(t.forum) = lower($1)
		ORDER BY t.created DESC 
		LIMIT $2`
		rows, err = ts.db.Query(sqlQuery, forum, limit)
	} else {
		sqlQuery := `
		SELECT t.author, t.created, t.id, t.message, t.slug, t.title, t.votes
		FROM forum.thread as t 
		WHERE lower(t.forum) = lower($1) AND t.created <= $3
		ORDER BY t.created DESC 
		LIMIT $2`
		rows, err = ts.db.Query(sqlQuery, forum, limit, since)
	}

	defer rows.Close()

	for rows.Next() {
		threadScan := Thread{}
		slug := sql.NullString{}
		err := rows.Scan(&threadScan.Author, &threadScan.Created, &threadScan.Id, &threadScan.Message, &slug, &threadScan.Title, &threadScan.Votes)
		if err != nil {
			return threads, err
		}
		if slug.Valid {
			threadScan.Slug = slug.String
		}
		threadScan.Forum = forum
		threads = append(threads, threadScan)
	}
	return
}

func (ts *ThreadService) FindThreadBySlug(slug string) (thread Thread, err error) {
	sqlQuery := `SELECT t.author, t.created, t.id, t.forum, t.message, t.slug, t.title
	FROM forum.thread as t 
	where lower(t.slug) = lower($1)`
	err = ts.db.QueryRow(sqlQuery, slug).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title)
	return
}

func (ts *ThreadService) FindThreadById(id int) (thread Thread, err error) {
	sqlQuery := `SELECT t.author, t.created, t.id, t.forum, t.message, t.slug, t.title
	FROM forum.thread as t 
	where t.id = $1`
	err = ts.db.QueryRow(sqlQuery, id).Scan(&thread.Author, &thread.Created, &thread.Id, &thread.Forum, &thread.Message, &thread.Slug, &thread.Title)
	return
}

func (ts *ThreadService) InsertVote(vote Vote) (err error) {
	sqlQuery := `INSERT INTO forum.vote ("user", voice, thread_id)
	VALUES ($1,$2,$3)`
	_, err = ts.db.Exec(sqlQuery, vote.NickName, vote.Voice, vote.ThreadId)
	return
}

func (ts *ThreadService) SelectVote(vote Vote) (findVote Vote, err error) {
	sqlQuery := `
	SELECT v."user", v.voice, v.thread_id 
	FROM forum.vote as v
	where Lower(v."user") = Lower($2) AND v.thread_id = $1`
	err = ts.db.QueryRow(sqlQuery, vote.ThreadId, vote.NickName).Scan(&findVote.NickName, &findVote.Voice, &findVote.ThreadId)
	return
}

func (ts *ThreadService) UpdateVote(vote Vote) (countUpdatedRows int64, err error) {
	sqlQuery := `
	UPDATE forum.vote SET voice = $1
	where Lower(vote."user") = Lower($2) AND vote.thread_id = $3`
	result, err := ts.db.Exec(sqlQuery, vote.Voice, vote.NickName, vote.ThreadId)
	if err != nil {
		return
	}
	countUpdatedRows = result.RowsAffected()
	return
}

func (ts *ThreadService) UpdateThread(thread Thread) (err error) {
	sqlQuery := `
	UPDATE forum.thread SET message = $1,
	                         title = $2
	where thread.id = $3`
	_, err = ts.db.Exec(sqlQuery, thread.Message, thread.Title, thread.Id)
	return
}

func (ts *ThreadService) SelectPosts(threadID int, limit, since, sort, desc string) (Posts []Post, Err error) {
	var sqlQuery string

	conditionSign := ">"
	if desc == "desc" {
		conditionSign = "<"
	}

	if sort == "flat" {
		sqlQuery = "SELECT p.id, p.parent, p.thread, p.forum, p.author, p.created, p.message, p.is_edited, p.path FROM forum.post as p WHERE thread = $1 "
		if since != "" {
			sqlQuery += fmt.Sprintf(" AND id %s %s ", conditionSign, since)
		}
		sqlQuery += fmt.Sprintf(" ORDER BY p.created %s, p.id %s LIMIT %s", desc, desc, limit)
	} else if sort == "tree" {
		orderString := fmt.Sprintf(" ORDER BY p.path[1] %s, p.path %s ", desc, desc)
		sqlQuery = "SELECT p.id, p.parent, p.thread, p.forum, p.author, p.created, p.message, p.is_edited, p.path " +
			"FROM forum.post as p " +
			"WHERE p.thread = $1 "
		if since != "" {
			sqlQuery += fmt.Sprintf(" AND p.path %s (SELECT p.path FROM forum.post as p WHERE p.id = %s) ", conditionSign, since)
		}
		sqlQuery += orderString
		sqlQuery += fmt.Sprintf("LIMIT %s", limit)

	} else if sort == "parent_tree" {
		sqlQuery = "SELECT p.id, p.parent, p.thread, p.forum, p.author, p.created, p.message, p.is_edited, p.path " +
			"FROM forum.post as p " +
			"WHERE p.thread = $1 AND p.path::integer[] && (SELECT ARRAY (select p.id from forum.post as p WHERE p.thread = $1 AND p.parent = 0 "
		if since != "" {
			sqlQuery += fmt.Sprintf(" AND p.path %s (SELECT p.path[1:1] FROM forum.post as p WHERE p.id = %s) ", conditionSign, since)
		}
		sqlQuery += fmt.Sprintf("ORDER BY p.path[1] %s, p.path LIMIT %s)) ", desc, limit)
		sqlQuery += fmt.Sprintf("ORDER BY p.path[1] %s, p.path ", desc)
	}

	rows, err := ts.db.Query(sqlQuery, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		p := Post{}
		err := rows.Scan(&p.Id, &p.Parent, &p.Thread, &p.Forum, &p.Author, &p.Created, &p.Message, &p.IsEdited, pq.Array(&p.Path))
		if err != nil {
			return nil, err
		}

		Posts = append(Posts, p)
	}

	return
}

func (ts *ThreadService) UpdateVoteCount(vote Vote) (err error) {
	sqlQuery := `
	UPDATE forum.thread SET votes = votes + $1
	where thread.id = $2`
	_, err = ts.db.Exec(sqlQuery, vote.Voice, vote.ThreadId)
	return
}
