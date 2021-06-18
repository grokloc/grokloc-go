package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
)

// CreateMsg is what a client should marshal to send as a json body to CreateOrg
type CreateMsg struct {
	Name string `json:"name"`
}

// CreateOrg creates a new org based on seed data in the POST body.
func (srv Instance) CreateOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		sugar.Debugw("context authlevel missing",
			"reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// only root can create an org
	if authLevel != AuthRoot {
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
	var c CreateMsg
	err = json.Unmarshal(body, &c)
	if err != nil {
		http.Error(w, "malformed org create", http.StatusBadRequest)
		return
	}

	o, err := org.New(c.Name)
	if err != nil {
		http.Error(w, "malformed org name", http.StatusBadRequest)
		return
	}

	err = o.Insert(ctx, srv.ST.Master)
	if err != nil {
		if err == models.ErrConflict {
			http.Error(w, "duplicate org name", http.StatusConflict)
			return
		}
		sugar.Debugw("insert org",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", OrgRoute+"/"+o.ID)
	w.WriteHeader(http.StatusCreated)
}
