package authctx

const LocalsKey = "authenticated_request"

type Context struct {
	UserID    string
	SessionID string
	Role      string
}
