package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
)

// CreateUserMsg is what a client should marshal to send as a json body to CreateUser
type CreateUserMsg struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Org         string `json:"org"`
	Password    string `json:"password"`
}

// UpdateUserDisplayNameMsg is the body format to update the user display name
type UpdateUserDisplayNameMsg struct {
	DisplayName string `json:"display_name"`
}

// UnmarshalJSON is a custom unmarshal for UpdateUserDisplayNameMsg
func (m *UpdateUserDisplayNameMsg) UnmarshalJSON(bs []byte) error {
	var t map[string]string
	err := json.Unmarshal(bs, &t)
	if err != nil {
		return err
	}
	v, ok := t["display_name"]
	if !ok {
		return errors.New("no display_name field found")
	}
	m.DisplayName = v
	return nil
}

// UpdateUserPasswordMsg is the body format to update the user password
type UpdateUserPasswordMsg struct {
	Password string `json:"password"`
}

// UnmarshalJSON is a custom unmarshal for UpdateUserPasswordMsg
func (m *UpdateUserPasswordMsg) UnmarshalJSON(bs []byte) error {
	var t map[string]string
	err := json.Unmarshal(bs, &t)
	if err != nil {
		return err
	}
	v, ok := t["password"]
	if !ok {
		return errors.New("no password field found")
	}
	m.Password = v
	return nil
}

// CreateUser creates a new org based on seed data in the POST body
func (srv Instance) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}

	if authLevel == AuthUser {
		http.Error(w, "auth inadequate", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sugar.Debugw("read body",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var m CreateUserMsg
	err = json.Unmarshal(body, &m)
	if err != nil {
		http.Error(w, "malformed user create", http.StatusBadRequest)
		return
	}

	u, err := user.New(m.DisplayName, m.Email, m.Org, m.Password)
	if err != nil {
		http.Error(w, "malformed user args", http.StatusBadRequest)
		return
	}

	if authLevel == AuthOrg {
		// must be in same org as prospective user
		session, ok := ctx.Value(sessionCtxKey).(Session)
		if !ok {
			panic("session missing")
		}
		if session.Org.ID != m.Org {
			http.Error(w, "not a member of requested org", http.StatusForbidden)
			return
		}
	}
	// now, either root or org owner

	err = u.Insert(ctx, srv.ST.Master, srv.ST.Key)
	if err != nil {
		if err == models.ErrConflict {
			http.Error(w, "duplicate user args", http.StatusConflict)
			return
		}
		sugar.Debugw("insert user",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", UserRoute+"/"+u.ID)
	w.WriteHeader(http.StatusCreated)
}

// ReadUser reads a user
func (srv Instance) ReadUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	id := chi.URLParam(r, IDParam)
	if len(id) == 0 {
		panic("id missing")
	}

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}
	session, ok := ctx.Value(sessionCtxKey).(Session)
	if !ok {
		panic("session missing")
	}

	var u *user.Instance
	if authLevel == AuthUser {
		if session.User.ID == id {
			u = &session.User
		} else {
			http.Error(w, "cannot read another user", http.StatusForbidden)
			return
		}
	}

	var err error
	if u == nil {
		u, err = user.Read(ctx, srv.ST.Master, srv.ST.Key, id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "user not found or inactive", http.StatusNotFound)
				return
			}
			sugar.Debugw("read user",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	if authLevel == AuthOrg {
		if session.Org.ID != u.Org {
			http.Error(w, "not a member of requested org", http.StatusForbidden)
			return
		}
	}

	bs, err := json.Marshal(u)
	if err != nil {
		sugar.Debugw("marshal user",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, err = w.Write(bs)
	if err != nil {
		panic(err.Error())
	}
}

// UpdateUser updates user display name, password, or status
func (srv Instance) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	id := chi.URLParam(r, IDParam)
	if len(id) == 0 {
		panic("id missing")
	}

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}
	session, ok := ctx.Value(sessionCtxKey).(Session)
	if !ok {
		panic("session missing")
	}

	if authLevel == AuthUser {
		http.Error(w, "auth inadequate", http.StatusForbidden)
		return
	}

	u, err := user.Read(ctx, srv.ST.RandomReplica(), srv.ST.Key, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "user not found or inactive", http.StatusNotFound)
			return
		}
		sugar.Debugw("read user",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if authLevel == AuthOrg {
		if session.Org.ID != u.Org {
			http.Error(w, "not a member of requested org", http.StatusForbidden)
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sugar.Debugw("read body",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// only one column update per call is allowed
	// try matching on display name update
	var displayNameMsg UpdateUserDisplayNameMsg
	err = json.Unmarshal(body, &displayNameMsg)
	// err will be non-nil if unmarshal fails - we have a custom unmarshal here
	if err == nil {
		err := u.UpdateDisplayName(ctx, srv.ST.Master, srv.ST.Key, displayNameMsg.DisplayName)
		if err != nil {
			sugar.Debugw("update display name",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// try matching on password update
	var passwordMsg UpdateUserPasswordMsg
	err = json.Unmarshal(body, &passwordMsg)
	// err will be non-nil if unmarshal fails - we have a custom unmarshal here
	if err == nil {
		derived, err := security.DerivePassword(passwordMsg.Password, srv.ST.Argon2Cfg)
		if err != nil {
			sugar.Debugw("update password",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		err = u.UpdatePassword(ctx, srv.ST.Master, derived)
		if err != nil {
			sugar.Debugw("update password",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// try matching on status update
	var statusMsg UpdateStatusMsg
	err = json.Unmarshal(body, &statusMsg)
	// err will be non-nil if unmarshal fails - we have a custom unmarshal here
	if err == nil {
		err := u.UpdateStatus(ctx, srv.ST.Master, statusMsg.Status)
		if err != nil {
			if err == models.ErrDisallowedValue {
				http.Error(w, "status value disallowed", http.StatusBadRequest)
				return
			}
			sugar.Debugw("update status",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// no update formats matched
	http.Error(w, "malformed update msg", http.StatusBadRequest)
}
