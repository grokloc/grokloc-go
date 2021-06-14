package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt_go "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-go/pkg/jwt"
	"github.com/grokloc/grokloc-go/pkg/security"
)

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
		sugar.Debugw("context session missing", "reqid", middleware.GetReqID(ctx))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tokenRequest := r.Header.Get(TokenRequestHeader)
	if len(tokenRequest) == 0 {
		http.Error(w, fmt.Sprintf("missing: %s", TokenRequestHeader), http.StatusBadRequest)
		return
	}
	if tokenRequest != security.EncodedSHA256(session.User.ID+session.User.APISecret) {
		sugar.Debugw("verify token request",
			"reqid", middleware.GetReqID(ctx),
			"tokenrequest", tokenRequest,
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
		http.Error(w, "internal error.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json, err := json.Marshal(Token{Bearer: signedToken, Expires: claims.ExpiresAt})
	if err != nil {
		sugar.Debugw("marshal token",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error.", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(json)
	if err != nil {
		panic(err.Error())
	}
}
