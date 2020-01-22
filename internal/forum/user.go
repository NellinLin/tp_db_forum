package forum

import (
	"github.com/jackc/pgx"
)

type UserService struct {
	db *pgx.ConnPool
}

func NewUserService(db *pgx.ConnPool) *UserService {
	return &UserService{db: db}
}

func (us *UserService) SelectUserByNickNameOrEmail(nickName, email string) (users []User, err error) {
	sqlQuery := `SELECT u.nick_name, u.email, u.full_name, u.about
	FROM forum.user as u 
	where lower(u.nick_name) = lower($1) or lower(u.email) = lower($2)`
	rows, err := us.db.Query(sqlQuery, nickName, email)
	if err != nil {
		return users, err
	}

	defer rows.Close()

	for rows.Next() {
		userScan := User{}
		err := rows.Scan(&userScan.NickName, &userScan.Email, &userScan.FullName, &userScan.About)
		if err != nil {
			return users, err
		}
		users = append(users, userScan)
	}
	return users, nil
}

func (us *UserService) SelectUserByNickName(nickName string) (user User, err error) {
	sqlQuery := `SELECT u.nick_name, u.email, u.full_name, u.about
	FROM forum.user as u 
	where lower(u.nick_name) = lower($1)`
	err = us.db.QueryRow(sqlQuery, nickName).Scan(&user.NickName, &user.Email, &user.FullName, &user.About)
	return
}

func (us *UserService) SelectUsersByForum(forum string, limit int, since string, desc string) (users []User, err error) {
	var rows *pgx.Rows
	if since == "" {
		if desc == "false" {
			sqlQuery := `
		SELECT u.nick_name, u.email, u.full_name, u.about
		FROM forum.user as u
		JOIN forum.forum_user as fu ON Lower(fu.user) = Lower(u.nick_name)
		WHERE Lower(fu.forum) = Lower($1)
		ORDER BY LOWER(nick_name) ASC
		LIMIT $2`
			rows, err = us.db.Query(sqlQuery, forum, limit)
			if err != nil {
				return
			}
		} else {
			sqlQuery := `
		SELECT u.nick_name, u.email, u.full_name, u.about
		FROM forum.user as u
		JOIN forum.forum_user as fu ON Lower(fu.user) = Lower(u.nick_name)
		WHERE Lower(fu.forum) = Lower($1)
		ORDER BY LOWER(nick_name) DESC
		LIMIT $2`
			rows, err = us.db.Query(sqlQuery, forum, limit)
			if err != nil {
				return
			}
		}
	} else {
		if desc == "false" {
			sqlQuery := `
		SELECT u.nick_name, u.email, u.full_name, u.about
		FROM forum.user as u
		JOIN forum.forum_user as fu ON Lower(fu.user) = Lower(u.nick_name)
		WHERE Lower(fu.forum) = Lower($1) AND LOWER(nick_name) > lower($3)
		ORDER BY LOWER(nick_name) ASC
		LIMIT $2`
			rows, err = us.db.Query(sqlQuery, forum, limit, since)
			if err != nil {
				return
			}
		} else {
			sqlQuery := `
		SELECT u.nick_name, u.email, u.full_name, u.about
		FROM forum.user as u
		JOIN forum.forum_user as fu ON Lower(fu.user) = Lower(u.nick_name)
		WHERE Lower(fu.forum) = Lower($1) AND LOWER(nick_name) < lower($3)
		ORDER BY LOWER(nick_name) DESC
		LIMIT $2`
			rows, err = us.db.Query(sqlQuery, forum, limit, since)
			if err != nil {
				return
			}
		}
	}

	defer rows.Close()

	for rows.Next() {
		user := User{}
		err := rows.Scan(&user.NickName, &user.Email, &user.FullName, &user.About)
		if err != nil {
			return users, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (us *UserService) InsertUser(user User) error {
	sqlQuery := `INSERT INTO forum.user (nick_name,email,full_name,about)
	VALUES ($1,$2,$3,$4)`
	_, err := us.db.Exec(sqlQuery, user.NickName, user.Email, user.FullName, user.About)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) UpdateUser(user User) error {
	sqlQuery := `UPDATE forum.user
	SET email = $1, 
		full_name = $2, 	
		about = $3
		WHERE Lower(nick_name) = Lower($4)`
	_, err := us.db.Exec(sqlQuery, user.Email, user.FullName, user.About, user.NickName)
	if err != nil {
		return err
	}
	return nil
}

func (us *UserService) FindUserByNickName(nickName string) (user User, err error) {
	sqlQuery := `SELECT u.nick_name
	FROM forum.user as u 
	where lower(u.nick_name) = lower($1)`
	err = us.db.QueryRow(sqlQuery, nickName).Scan(&user.NickName)
	return
}
