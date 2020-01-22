package handlers

import (
	"github.com/NellinLin/tp_db_forum/internal/forum"
	"github.com/NellinLin/tp_db_forum/internal/serve"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Post struct {
	Log           *zerolog.Logger
	ForumService  *forum.ForumService
	UserService   *forum.UserService
	ThreadService *forum.ThreadService
	PostService   *forum.PostService
}

func (h *Post) GetFullPost(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	related := r.FormValue("related")

	post, err := h.PostService.SelectPostById(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	fullPost := forum.FullPost{Post: post}

	if strings.Contains(related, "user") {
		user, err := h.UserService.SelectUserByNickName(post.Author)
		if err != nil {
			if err == pgx.ErrNoRows {
				serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
				return
			}
			serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
			return
		}
		fullPost.Author = user
	}

	if strings.Contains(related, "forum") {
		fullForum, err := h.ForumService.SelectForumBySlug(post.Forum)
		if err != nil {
			if err == pgx.ErrNoRows {
				serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
				return
			}
			serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
			return
		}
		fullPost.Forum = fullForum
	}

	if strings.Contains(related, "thread") {
		thread, err := h.ThreadService.SelectThreadById(post.Thread)
		if err != nil {
			serve.ServeJSON(w, http.StatusBadRequest, "")
			return
		}
		fullPost.Thread = thread
	}
	serve.ServeJSON(w, http.StatusOK, fullPost)
}

func (h *Post) EditMessage(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	if id < 0 {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	var editMessage forum.Message

	if err := UnmarshalBody(r.Body, &editMessage); err != nil {
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	post, err := h.PostService.SelectPostById(id)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post"})
		return
	}

	if editMessage.Message != "" && editMessage.Message != post.Message {
		num, err := h.PostService.UpdatePostMessage(editMessage.Message, id)
		if err != nil {
			h.Log.Warn().Msg(err.Error())
			serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
			return
		}
		if num != 1 {
			serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Can't find post"})
			return
		}
		post.Message = editMessage.Message
		post.IsEdited = true
	}
	serve.ServeJSON(w, http.StatusOK, post)
}

func (h *Post) CreatePosts(w http.ResponseWriter, r *http.Request) {
	createdTime := time.Now().Format(time.RFC3339Nano)
	slugOrId := mux.Vars(r)["slug_or_id"]

	var newPosts []forum.Post

	if err := UnmarshalBody(r.Body, &newPosts); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		slug := slugOrId
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	}
	if len(newPosts) == 0 {
		serve.ServeJSON(w, http.StatusCreated, newPosts)
		return
	}
	_, err = h.ForumService.SelectForumBySlug(thread.Forum)
	if err != nil {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
		return
	}

	posts, err := h.PostService.CreatePosts(thread, createdTime, newPosts)
	if err != nil {
		if err.Error() == "404" {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find post author by nickname:"})
			return
		}
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusConflict, forum.ErrorMessage{Message: "Parent post was created in another thread"})
		return
	}

	err = h.ForumService.UpdatePostCount(thread.Forum, len(newPosts))
	if err != nil {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusCreated, posts)
}

func (h *Post) EditThread(w http.ResponseWriter, r *http.Request) {
	slugOrId := mux.Vars(r)["slug_or_id"]

	var editThread forum.Thread
	if err := UnmarshalBody(r.Body, &editThread); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		slug := slugOrId
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	}
	if editThread.Message != "" {
		thread.Message = editThread.Message
	}
	if editThread.Title != "" {
		thread.Title = editThread.Title
	}
	if editThread.Message == "" && editThread.Title == "" {
		serve.ServeJSON(w, http.StatusOK, thread)
		return
	}
	err = h.ThreadService.UpdateThread(thread)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusOK, thread)
}
func (h *Post) CreateVote(w http.ResponseWriter, r *http.Request) {
	slugOrId := mux.Vars(r)["slug_or_id"]

	var newVote forum.Vote
	if err := UnmarshalBody(r.Body, &newVote); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		slug := slugOrId
		thread, err = h.ThreadService.FindThreadBySlug(slug)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	} else {
		thread, err = h.ThreadService.FindThreadById(id)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	}

	user, err := h.UserService.FindUserByNickName(newVote.NickName)
	if err != nil {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
		return
	}

	newVote.ThreadId = thread.Id
	newVote.NickName = user.NickName
	vote, err := h.ThreadService.SelectVote(newVote)
	if err != nil {
		err = h.ThreadService.InsertVote(newVote)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
			return
		}
		err = h.ThreadService.UpdateVoteCount(newVote)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
			return
		}
	} else {
		_, err = h.ThreadService.UpdateVote(newVote)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
			return
		}
		if vote.Voice == -1 && newVote.Voice == 1 {
			newVote.Voice = 2
		} else {
			if vote.Voice == 1 && newVote.Voice == -1 {
				newVote.Voice = -2
			} else {
				newVote.Voice = 0
			}
		}
		err = h.ThreadService.UpdateVoteCount(newVote)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
			return
		}
	}

	thread, err = h.ThreadService.SelectThreadById(newVote.ThreadId)
	if err != nil {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
		return
	}

	serve.ServeJSON(w, http.StatusOK, thread)
}

func (h *Post) GetThread(w http.ResponseWriter, r *http.Request) {
	slugOrId := mux.Vars(r)["slug_or_id"]

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		slug := slugOrId
		thread, err = h.ThreadService.SelectThreadBySlug(slug)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	} else {
		thread, err = h.ThreadService.SelectThreadById(id)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	}

	serve.ServeJSON(w, http.StatusOK, thread)
}

func (h *Post) GetPosts(w http.ResponseWriter, r *http.Request) {
	slugOrId := mux.Vars(r)["slug_or_id"]

	limit := r.FormValue("limit")
	since := r.FormValue("since")
	sort := r.FormValue("sort")
	desc := r.FormValue("desc")

	if limit == "" {
		limit = "100"
	}

	if sort == "" {
		sort = "flat"
	}
	if desc == "true" {
		desc = "desc"
	} else {
		desc = ""
	}

	var thread forum.Thread
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		thread, err = h.ThreadService.SelectThreadBySlug(slugOrId)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	} else {
		thread, err = h.ThreadService.SelectThreadById(id)
		if err != nil {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find thread"})
			return
		}
	}

	posts, err := h.ThreadService.SelectPosts(thread.Id, limit, since, sort, desc)
	if err != nil {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		return
	}
	if len(posts) == 0 {
		postss := []Post{}
		serve.ServeJSON(w, http.StatusOK, postss)
		return
	}
	serve.ServeJSON(w, http.StatusOK, posts)
}
