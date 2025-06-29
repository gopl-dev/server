package server

// LoggingMiddleware logs the request details
//func LoggingMiddleware(next Handler) Handler {
//	return func(w http.ResponseWriter, r *http.Request) {
//		start := time.Now()
//		log.Printf("Started %s %s", r.Method, r.URL.Path)
//		next(w, r)
//		log.Printf("Completed in %v", time.Since(start))
//	}
//}
//
//func RecoveryMiddleware(next Handler) Handler {
//	return func(w http.ResponseWriter, r *http.Request) {
//		defer func() {
//			if err := recover(); err != nil {
//				w.WriteHeader(http.StatusInternalServerError)
//				log.Printf("[%s %s] Recovered from panic: %s", r.Method, r.URL.Path, err)
//			}
//		}()
//
//		next(w, r)
//	}
//}

// RequestAuthMiddleware middleware
// See RequestUserMiddleware middleware for auth details
//func RequestAuthMiddleware(c *gin.Context) {
//	if !c.GetBool(authSuccessKey) {
//		errText := c.GetString(authErrorKey)
//		if errText == "" {
//			errText = http.StatusText(http.StatusUnauthorized)
//		}
//
//		abort(c, app.ErrUnauthorized(errText))
//		return
//	}
//
//	c.Next()
//}

const authSuccessKey = "auth_success"
const authErrorKey = "auth_error"

// RequestUserMiddleware Middleware
// Attempts to resolve the user based on the provided authentication token.
// Some public routes rely on the current user for their response logic.
// Authorization also occurs here to ensure the authentication token is valid.
// However, authorization errors are not thrown in this middleware.
// Instead, they are added to the request context for potential use by the authorization middleware later, if needed.
//func RequestUserMiddleware(c *gin.Context) {
//	sig := c.Request.Header.Get("Authorization")
//	if sig == "" {
//		c.Next()
//		return
//	}
//
//	sig = strings.TrimPrefix(sig, "Bearer ")
//	if len(sig) < 128 {
//		c.Set(authErrorKey, "invalid auth token")
//		c.Next()
//		return
//	}
//
//	token := sig[:64]
//	tok, err := authtoken.FindByToken(token)
//	if err != nil {
//		c.Set(authErrorKey, "unknown auth token")
//		c.Next()
//		return
//	}
//
//	tok.ClientName = Client(c)
//	tok.ClientIP = c.Request.RemoteAddr
//	tok.UserAgent = c.Request.UserAgent()
//
//	if sig != authtoken.Sign(&tok) {
//		c.Set(authErrorKey, "invalid auth token signature")
//		c.Next()
//		return
//	}
//
//	if tok.ExpiresAt.Before(time.Now()) {
//		c.Set(authErrorKey, "auth token expired")
//		c.Next()
//		return
//	}
//
//	userEl := tok.User
//	err = user.SetActiveAt(userEl.ID)
//	if err != nil {
//		abort(c, err)
//		return
//	}
//
//	err = authtoken.Prolong(tok.ID)
//	if err != nil {
//		abort(c, err)
//		return
//	}
//
//	c.Set(authSuccessKey, true)
//	SetAuthUser(c, *userEl)
//	c.Next()
//}
