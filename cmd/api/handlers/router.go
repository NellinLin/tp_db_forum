package handlers

import (
	"github.com/NellinLin/tp_db_forum/internal/forum"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"net/http"
)

func API(
	log *zerolog.Logger,
	forumService *forum.ForumService,
	postService *forum.PostService,
	threadService *forum.ThreadService,
	userService *forum.UserService,
	serviceService *forum.ServiceService) http.Handler {

	user := User{Log: log, UserService: userService}
	forum := Forum{Log: log, ForumService: forumService, UserService: userService, ThreadService: threadService, ServiceService: serviceService}
	post := Post{Log: log, PostService: postService, ForumService: forumService, UserService: userService, ThreadService: threadService}

	r := mux.NewRouter()

	r.HandleFunc("/api/user/{nickname}/create", user.CreateUser).Methods("POST")
	r.HandleFunc("/api/user/{nickname}/profile", user.GetProfile).Methods("GET")
	r.HandleFunc("/api/user/{nickname}/profile", user.EditProfile).Methods("POST")

	r.HandleFunc("/api/forum/create", forum.CreateForum).Methods("POST")
	r.HandleFunc("/api/forum/{slug}/create", forum.CreateThread).Methods("POST")
	r.HandleFunc("/api/forum/{slug}/details", forum.GetForumDetails).Methods("GET")
	r.HandleFunc("/api/forum/{slug}/threads", forum.GetForumThreads).Methods("GET")
	r.HandleFunc("/api/forum/{slug}/users", forum.GetForumUsers).Methods("GET")

	r.HandleFunc("/api/thread/{slug_or_id}/details", post.GetThread).Methods("GET")
	r.HandleFunc("/api/thread/{slug_or_id}/details", post.EditThread).Methods("POST")
	r.HandleFunc("/api/thread/{slug_or_id}/posts", post.GetPosts).Methods("GET")
	r.HandleFunc("/api/thread/{slug_or_id}/create", post.CreatePosts).Methods("POST")
	r.HandleFunc("/api/thread/{slug_or_id}/vote", post.CreateVote).Methods("POST")

	r.HandleFunc("/api/post/{id}/details", post.GetFullPost).Methods("GET")
	r.HandleFunc("/api/post/{id}/details", post.EditMessage).Methods("POST")

	r.HandleFunc("/api/service/clear", forum.Clean).Methods("POST")
	r.HandleFunc("/api/service/status", forum.Status).Methods("GET")

	return r
}
