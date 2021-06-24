package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/models"
	"github.com/grokloc/grokloc-go/pkg/models/org"
	"github.com/grokloc/grokloc-go/pkg/models/user"
	"github.com/grokloc/grokloc-go/pkg/security"
)

// Session is the org and user instances for a user account
type Session struct {
	Org  org.Instance
	User user.Instance
}

// WithSession reads the user and org using the X-GrokLOC-ID header,
// performs basic validation, and then adds a user and org instance to
// the context.
func (srv *Instance) WithSession(next http.Handler) http.Handler {
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
		if session.Org.ID == srv.ST.RootOrg {
			// allow for multiple accounts in root org
			authLevel = AuthRoot
		} else if session.Org.Owner == session.User.ID {
			authLevel = AuthOrg
		}
		r = r.WithContext(context.WithValue(ctx, authLevelCtxKey, authLevel))
		// r.Context() to get ctx with authLevel
		r = r.WithContext(context.WithValue(r.Context(), sessionCtxKey, *session))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// WithToken extracts the JWT from the X-GrokLOC-Token header
// and validates the claims
func (srv Instance) WithToken(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		session, ok := ctx.Value(sessionCtxKey).(Session)
		if !ok {
			panic("session missing")
		}
		token := r.Header.Get(jwt.Authorization)
		if len(token) == 0 {
			http.Error(w, fmt.Sprintf("missing: %s", jwt.Authorization), http.StatusBadRequest)
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

// Token describes the token value and the expiration unixtime
type Token struct {
	Bearer  string `json:"bearer"`
	Expires int64  `json:"expires"`
}

// NewToken returns a response containing a new JWT
func (srv *Instance) NewToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer srv.ST.L.Sync() // nolint
	sugar := srv.ST.L.Sugar()

	session, ok := ctx.Value(sessionCtxKey).(Session)
	if !ok {
		panic("session missing")
	}
	tokenRequest := r.Header.Get(TokenRequestHeader)
	if len(tokenRequest) == 0 {
		http.Error(w, fmt.Sprintf("missing: %s", TokenRequestHeader), http.StatusBadRequest)
		return
	}
	validate := security.EncodedSHA256(session.User.ID + session.User.APISecret)
	if tokenRequest != validate {
		sugar.Debugw("verify token request",
			"reqid", middleware.GetReqID(ctx),
			"tokenrequest", tokenRequest,
			"validate", validate,
			"id", session.User.ID)
		http.Error(w, "token request invalid", http.StatusUnauthorized)
		return
	}
	claims, err := jwt.New(session.User)
	if err != nil {
		sugar.Debugw("create new claims",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	token := jwt_go.NewWithClaims(jwt_go.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(session.User.ID + string(srv.ST.SigningKey)))
	if err != nil {
		sugar.Debugw("encode token",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	bs, err := json.Marshal(Token{Bearer: signedToken, Expires: claims.ExpiresAt})
	if err != nil {
		sugar.Debugw("marshal token",
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
