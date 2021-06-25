package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/user"
)

// CreateUserMsg is what a client should marshal to send as a json body to CreateUser
type CreateUserMsg struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Org         string `json:"org"`
	Password    string `json:"password"`
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
