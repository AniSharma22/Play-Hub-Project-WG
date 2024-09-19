package middleware

import (
	"net/http"
	"project2/pkg/errs"
)

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Extract role from the context
		role, ok := r.Context().Value(RoleKey).(string)
		if !ok {
			errs.NewUnexpectedError("Could not find the role").ToJSON(w)
			return
		}

		// Use the role
		if role != "admin" {
			unauthorized(w, "Unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	}
}
