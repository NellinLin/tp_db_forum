package forum

import "github.com/jackc/pgx"

type ServiceService struct {
	db *pgx.ConnPool
}

func NewServiceService(db *pgx.ConnPool) *ServiceService {
	return &ServiceService{db: db}
}

func (ss *ServiceService) Clean() (err error) {
	sqlQuery := `TRUNCATE forum.vote, forum.post, forum.thread, forum.forum, forum.user, forum.forum_user RESTART IDENTITY CASCADE;`
	_, err = ss.db.Exec(sqlQuery)
	return
}

func (ss *ServiceService) SelectStatus() (status Status, err error) {
	sqlQuery := `
	SELECT *
	FROM 
		(SELECT COUNT(*) AS "user" FROM forum.user) AS Users,
		(SELECT COUNT(*) AS forum FROM forum.forum) AS Forum,
		(SELECT COUNT(*) AS thread FROM forum.thread) AS Thread,
		(SELECT COUNT(*) AS post FROM forum.post) AS Post;`
	err = ss.db.QueryRow(sqlQuery).Scan(&status.User, &status.Forum, &status.Thread, &status.Post)
	return
}