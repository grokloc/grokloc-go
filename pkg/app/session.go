package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/models/user"
)

// Session is the org and user instances for a user account
type Session struct {
	Org  org.Instance
	User user.Instance
}

// GetUserAndOrg reads the user and org using the X-GrokLOC-ID header,
// performs basic validation, and then adds a user and org instance to
// the context.
func (srv *Instance) GetUserAndOrg(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer srv.ST.L.Sync() // nolint
		sugar := srv.ST.L.Sugar()

		id := r.Header.Get(IDHeader)
		if len(id) == 0 {
			http.Error(w, fmt.Sprintf("missing: %s", IDHeader), http.StatusBadRequest)
			return
		}

		user, err := user.Read(ctx, srv.ST.RandomReplica(), srv.ST.Key, id)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "user not found", http.StatusBadRequest)
				return
			}
			sugar.Debugw("read user",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if user.Meta.Status != models.StatusActive {
			http.Error(w, "user not active", http.StatusBadRequest)
			return
		}

		org, err := org.Read(ctx, srv.ST.RandomReplica(), user.Org)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "org not found", http.StatusBadRequest)
				return
			}
			sugar.Debugw("read org",
				"reqid", middleware.GetReqID(ctx),
				"err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if org.Meta.Status != models.StatusActive {
			http.Error(w, "org not active", http.StatusBadRequest)
			return
		}

		session := &Session{Org: *org, User: *user}

		authLevel := AuthUser
		if session.Org.Owner == session.User.ID {
			authLevel = AuthOrg
		} else if session.User.ID == srv.ST.RootUser &&
			session.Org.ID == srv.ST.RootOrg {
			authLevel = AuthRoot
		}
		r = r.WithContext(context.WithValue(ctx, authLevelCtxKey, authLevel))
		// Note r.Context() to get ctx with authLevel.
		r = r.WithContext(context.WithValue(r.Context(), sessionCtxKey, *session))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// VerifyToken extracts the JWT from the X-GrokLOC-Token header
// and validates the claims
func (srv Instance) VerifyToken(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		defer srv.ST.L.Sync() // nolint
		sugar := srv.ST.L.Sugar()

		session, ok := ctx.Value(sessionCtxKey).(Session)
		if !ok {
			sugar.Debugw("context session missing", "reqid", middleware.GetReqID(ctx))
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		token := r.Header.Get(TokenHeader)
		if len(token) == 0 {
			http.Error(w, fmt.Sprintf("missing: %s", TokenHeader), http.StatusBadRequest)
			return
		}
		claims, err := jwt.Decode(session.User.ID, token, srv.ST.SigningKey)
		if err != nil {
			http.Error(w, "token decode error", http.StatusUnauthorized)
			return
		}
		if claims.Id != session.User.ID || claims.Org != session.Org.ID {
			http.Error(w, "token contents incorrect", http.StatusBadRequest)
			return
		}
		if claims.ExpiresAt < time.Now().Unix() {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
