package middleware

import (
	"net/http"
	"project2/pkg/errs"
	"project2/pkg/logger"
	"time"
)

func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extracting role from the context
		role, ok := r.Context().Value(RoleKey).(string)
		if !ok {
			logger.Logger.Errorw("Role not found in context", "method", r.Method, "url", r.URL.String(), "time", time.Now())
			//unauthorized(w, "2001", "Role not found in context") // Code 2001: Role not found
			errs.ForbiddenError("1007", "Role not found in context").ToJson2(w)
			return
		}

		// Checking if the role is admin
		if role != "admin" {
			logger.Logger.Warnw("Unauthorized access attempt", "role", role, "method", r.Method, "url", r.URL.String(), "time", time.Now())
			//unauthorized(w, "2002", "Unauthorized: Admin access required") // Code 2002: Unauthorized access
			errs.ForbiddenError("1008", "Unauthorized: Admin access required").ToJson2(w)
			return
		}

		logger.Logger.Infow("Admin access granted", "role", role, "method", r.Method, "url", r.URL.String(), "time", time.Now())
		next.ServeHTTP(w, r)
	}
}
