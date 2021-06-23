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
		panic("auth missing")
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
		panic("id missing")
	}

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}

	var o *org.Instance
	var err error

	// if root is the caller, the context org is the root org,
	// so read the requested org
	if authLevel == AuthRoot {
		o, err = org.Read(ctx, srv.ST.RandomReplica(), id)
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
	} else {
		// otherwise, a regular user or org owner
		// if caller is not root, it can only read its own org
		// (which is in context)
		session, ok := ctx.Value(sessionCtxKey).(Session)
		if !ok {
			panic("session missing")
		}
		if session.Org.ID != id {
			http.Error(w, "not a member of requested org", http.StatusForbidden)
			return
		}
		o = &session.Org
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

// UpdateOrg updates org owner or status
func (srv Instance) UpdateOrg(w http.ResponseWriter, r *http.Request) {
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
	// only root can change the org owner
	if authLevel != AuthRoot {
		http.Error(w, "auth inadequate", http.StatusForbidden)
		return
	}

	// the context org is the root org, so read in the requested org
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

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sugar.Debugw("read body",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// try matching on owner update
	var ownerMsg UpdateOrgOwnerMsg
	err = json.Unmarshal(body, &ownerMsg)
	if err == nil {
		err := o.UpdateOwner(ctx, srv.ST.Master, ownerMsg.Owner)
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
		return
	}

	// try matching on status update
	var statusMsg UpdateStatusMsg
	err = json.Unmarshal(body, &statusMsg)
	if err == nil {
		err := o.UpdateStatus(ctx, srv.ST.Master, statusMsg.Status)
		if err != nil {
			if err == models.ErrDisallowedValue {
				http.Error(w, "status value disallowed", http.StatusBadRequest)
				return
			}
			sugar.Debugw("update owner",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// no update formats matched
	http.Error(w, "malformed update", http.StatusBadRequest)
}
