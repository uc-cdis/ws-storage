package storage


// AppContext runtime context
type SessionContext struct {
	User    string
}

func NewSessionContext(user string) (cx *SessionContext) {
	cx = nil
	return cx
}
