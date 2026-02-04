package session

import (
	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/routewrap"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

// Session represents the user session
type Session struct {
	UserID types.UserID
}

// Get implements routewrap.ISession
func (s *Session) Get(key string) (interface{}, bool, error) {
	if s == nil {
		return nil, false, nil
	}
	switch key {
	case "UserID":
		return s.UserID, true, nil
	default:
		return nil, false, nil
	}
}

// SessionFactoryImpl implements the session factory interface
type SessionFactoryImpl struct{}

// SessionKeys implements routewrap.SessionFactory
func (SessionFactoryImpl) SessionKeys() []*routewrap.SessionKey {
	return routewrap.ExtractStructSessionKeys(Session{})
}

// GetSession implements routewrap.SessionFactory
func (SessionFactoryImpl) GetSession(ctx *gin.Context) routewrap.ISession {
	sess := Get(ctx)
	if sess == nil {
		return nil
	}
	return sess
}

// GetUserID returns the user ID from session
func (s *Session) GetUserID() types.UserID {
	if s == nil {
		return 0
	}
	return s.UserID
}

const sessionKey = "session"

// SetGin sets the session in gin context
func SetGin(ctx *gin.Context, sess *Session) {
	ctx.Set(sessionKey, sess)
}

// Get retrieves the session from gin context
func Get(ctx *gin.Context) *Session {
	val, exists := ctx.Get(sessionKey)
	if !exists {
		return nil
	}
	sess, ok := val.(*Session)
	if !ok {
		return nil
	}
	return sess
}
