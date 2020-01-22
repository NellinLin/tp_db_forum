package handlers

import (
	"encoding/json"
	"github.com/NellinLin/tp_db_forum/internal/forum"
	"github.com/NellinLin/tp_db_forum/internal/serve"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
)

type Forum struct {
	Log            *zerolog.Logger
	ForumService   *forum.ForumService
	UserService    *forum.UserService
	ThreadService  *forum.ThreadService
	ServiceService *forum.ServiceService
}

func UnmarshalBody(rBody io.ReadCloser, v interface{}) error {
	body, err := ioutil.ReadAll(rBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		return err
	}
	return nil
}

func (h *Forum) CreateForum(w http.ResponseWriter, r *http.Request) {

	var newForum forum.Forum
	if err := UnmarshalBody(r.Body, &newForum); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	fullForum, err := h.ForumService.SelectForumBySlug(newForum.Slug)
	if err == nil {
		serve.ServeJSON(w, http.StatusConflict, fullForum)
		return
	}
	if err != pgx.ErrNoRows {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	user, err := h.UserService.FindUserByNickName(newForum.User)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	newForum.User = user.NickName

	if err = h.ForumService.InsertForum(newForum); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusCreated, newForum)
	return
}

func (h *Forum) CreateThread(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]

	var newThread forum.Thread
	if err := UnmarshalBody(r.Body, &newThread); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	newThread.Forum = slug

	threadForum, err := h.ForumService.SelectForumBySlug(newThread.Forum)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	newThread.Forum = threadForum.Slug

	author, err := h.UserService.FindUserByNickName(newThread.Author)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	newThread.Author = author.NickName

	if newThread.Slug != "" {
		thread, err := h.ThreadService.SelectThreadBySlug(newThread.Slug)
		if err == nil {
			serve.ServeJSON(w, http.StatusConflict, thread)
			return
		}
		if err != pgx.ErrNoRows {
			h.Log.Warn().Msg(err.Error())
			serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
			return
		}
	}

	threadId, err := h.ThreadService.InsertThread(newThread)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	newThread.Id = threadId

	err = h.ForumService.UpdateThreadCount(newThread.Forum)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	err = h.ForumService.InsertForumUser(newThread.Forum, author.NickName)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
	}

	serve.ServeJSON(w, http.StatusCreated, newThread)
}

func (h *Forum) GetForumDetails(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	fullForum, err := h.ForumService.SelectForumBySlug(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	serve.ServeJSON(w, http.StatusOK, fullForum)
}

func (h *Forum) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
	}

	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}
	if limit < 0 {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	since := r.FormValue("since")

	descStr := r.FormValue("desc")
	desc, err := strconv.ParseBool(descStr)
	if err != nil {
		desc = false
	}
	threadsForum, err := h.ForumService.SelectForumBySlug(slug)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
			return
		}
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	threads, err := h.ThreadService.SelectThreadByForum(threadsForum.Slug, limit, since, desc)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	if len(threads) == 0 {
		threads := []forum.Thread{}
		serve.ServeJSON(w, http.StatusOK, threads)
		return
	}
	serve.ServeJSON(w, http.StatusOK, threads)
}

func (h *Forum) GetForumUsers(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	if slug == "" {
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	usersForum, err := h.ForumService.SelectForumBySlug(slug)
	if err != nil {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find forum"})
		return
	}
	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = math.MaxInt32
	}

	since := r.FormValue("since")

	desc := r.FormValue("desc")
	if desc == "" {
		desc = "false"
	}

	users, err := h.UserService.SelectUsersByForum(usersForum.Slug, limit, since, desc)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	if users == nil {
		nullUsers := []User{}
		serve.ServeJSON(w, http.StatusOK, nullUsers)
		return
	}

	serve.ServeJSON(w, http.StatusOK, users)
}

func (h *Forum) Clean(w http.ResponseWriter, r *http.Request) {
	err := h.ServiceService.Clean()
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusOK, nil)
}

func (h *Forum) Status(w http.ResponseWriter, r *http.Request) {
	status, err := h.ServiceService.SelectStatus()
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	serve.ServeJSON(w, http.StatusOK, status)
}
