package app

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
)

// CreateOrgMsg is what a client should marshal to send as a json body to CreateOrg
type CreateOrgMsg struct {
	Name string `json:"name"`
}

// UpdateOrgOwnerMsg is what a client should marshal to send as a json body to
// UpdateOrgOwner
type UpdateOrgOwnerMsg struct {
	Owner string `json:"owner"`
}

// CreateOrg creates a new org based on seed data in the POST body
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

	var m CreateOrgMsg
	err = json.Unmarshal(body, &m)
	if err != nil {
		http.Error(w, "malformed org create", http.StatusBadRequest)
		return
	}

	o, err := org.New(m.Name)
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

// ReadOrg reads an organization
func (srv Instance) ReadOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	id := chi.URLParam(r, IDParam)
	if len(id) == 0 {
		sugar.Debugw("context id param missing",
			"reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		sugar.Debugw("context authlevel missing",
			"reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if authLevel != AuthRoot {
		session, ok := ctx.Value(sessionCtxKey).(Session)
		if !ok {
			sugar.Debugw("context session missing",
				"reqid", middleware.GetReqID(ctx))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// root may read any org; otherwise, calling user must be in org
		if session.User.Org != id {
			http.Error(w, "auth inadequate", http.StatusForbidden)
			return
		}
	}

	o, err := org.Read(ctx, srv.ST.RandomReplica(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "org not found or inactive", http.StatusNotFound)
			return
		}
		sugar.Debugw("read org",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	bs, err := json.Marshal(o)
	if err != nil {
		sugar.Debugw("marshal org",
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

// UpdateOrgOwner sets the org owner
func (srv Instance) UpdateOrgOwner(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	id := chi.URLParam(r, IDParam)
	if len(id) == 0 {
		sugar.Debugw("context id param missing",
			"reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		sugar.Debugw("context authlevel missing",
			"reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// only root can change the org owner
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

	var m UpdateOrgOwnerMsg
	err = json.Unmarshal(body, &m)
	if err != nil {
		http.Error(w, "malformed owner update", http.StatusBadRequest)
		return
	}

	o, err := org.Read(ctx, srv.ST.RandomReplica(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "org not found or inactive", http.StatusNotFound)
			return
		}
		sugar.Debugw("read org",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	err = o.UpdateOwner(ctx, srv.ST.Master, m.Owner)
	if err != nil {
		if err == models.ErrRelatedUser {
			http.Error(w, "prospective owner not in org", http.StatusBadRequest)
			return
		}
		sugar.Debugw("update owner",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
