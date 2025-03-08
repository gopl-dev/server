package server

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
