package handlers

import (
	"github.com/NellinLin/tp_db_forum/internal/forum"
	"github.com/NellinLin/tp_db_forum/internal/serve"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/rs/zerolog"
	"net/http"
)

type User struct {
	Log         *zerolog.Logger
	UserService *forum.UserService
}

func (h *User) CreateUser(w http.ResponseWriter, r *http.Request) {
	nickName := mux.Vars(r)["nickname"]

	var newUser forum.User

	if err := UnmarshalBody(r.Body, &newUser); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	newUser.NickName = nickName

	userSlice, err := h.UserService.SelectUserByNickNameOrEmail(newUser.NickName, newUser.Email)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	if len(userSlice) > 0 {
		serve.ServeJSON(w, http.StatusConflict, userSlice)
		return
	}

	if err = h.UserService.InsertUser(newUser); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusCreated, newUser)
}

func (h *User) GetProfile(w http.ResponseWriter, r *http.Request) {
	nickName := mux.Vars(r)["nickname"]

	user, err := h.UserService.SelectUserByNickName(nickName)
	if err != nil {
		if err == pgx.ErrNoRows {
			serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Can't find user"})
			return
		}
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		return
	}
	serve.ServeJSON(w, http.StatusOK, user)
}

func (h *User) EditProfile(w http.ResponseWriter, r *http.Request) {
	nickName := mux.Vars(r)["nickname"]

	var editUser forum.User

	if err := UnmarshalBody(r.Body, &editUser); err != nil {
		serve.ServeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	editUser.NickName = nickName

	asmeUsers, err := h.UserService.SelectUserByNickNameOrEmail(editUser.NickName, editUser.Email)
	if err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	if len(asmeUsers) == 0 {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "User not found"})
		return
	}

	if len(asmeUsers) > 1 {
		serve.ServeJSON(w, http.StatusConflict, forum.ErrorMessage{Message: "This email is already registered by user"})
		return
	}

	if asmeUsers[0].NickName != editUser.NickName {
		serve.ServeJSON(w, http.StatusNotFound, forum.ErrorMessage{Message: "Error"})
		return
	}

	if editUser.About == "" {
		editUser.About = asmeUsers[0].About
	}
	if editUser.Email == "" {
		editUser.Email = asmeUsers[0].Email
	}
	if editUser.FullName == "" {
		editUser.FullName = asmeUsers[0].FullName
	}

	if err = h.UserService.UpdateUser(editUser); err != nil {
		h.Log.Warn().Msg(err.Error())
		serve.ServeJSON(w, http.StatusBadRequest, forum.ErrorMessage{Message: "Error"})
		return
	}

	serve.ServeJSON(w, http.StatusOK, editUser)
}
