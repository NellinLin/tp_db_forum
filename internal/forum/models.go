package forum

import "time"

type User struct {
	About    string `json:"about"`
	Email    string `json:"email"`
	FullName string `json:"fullname"`
	NickName string `json:"nickname"`
}

type Forum struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	UserId  int    `json:"-"`
	User    string `json:"user"`
	Posts   int    `json:"posts"`
	Threads int    `json:"threads"`
}

type Thread struct {
	Author  string    `json:"author"`
	Created time.Time `json:"created"`
	Forum   string    `json:"forum"`
	Id      int       `json:"id"`
	Message string    `json:"message"`
	Slug    string    `json:"slug"`
	Title   string    `json:"title"`
	Votes   int       `json:"votes"`
}

type Post struct {
	Author        string  `json:"author"`
	Created       string  `json:"created"`
	Forum         string  `json:"forum"`
	Id            int     `json:"id"`
	IsEdited      bool    `json:"isEdited"`
	Message       string  `json:"message"`
	Parent        int     `json:"parent"`
	Thread        int     `json:"thread"`
	Path          []int64 `json:"-"`
	ParentPointer *Post   `json:"-"`
}

type Vote struct {
	NickName string `json:"nickname"`
	Voice    int    `json:"voice"`
	ThreadId int    `json:"-"`
}

type Message struct {
	Message string `json:"message"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type FullPost struct {
	Forum  interface{} `json:"forum,omitempty"`
	Thread interface{} `json:"thread,omitempty"`
	Author interface{} `json:"author,omitempty"`
	Post   interface{} `json:"post"`
}

type Status struct {
	Post   int `json:"post"`
	Thread int `json:"thread"`
	User   int `json:"user"`
	Forum  int `json:"forum"`
}
