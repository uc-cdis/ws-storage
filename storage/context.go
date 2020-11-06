package storage


// AppContext runtime context
type SessionContext struct {
	User    string
}

func NewSessionContext(user string) (cx *SessionContext) {
	cx = &SessionContext {
		User: user,
	};
	return cx
}
