package common

const (
	ServiceName = "vh-srv-events"

	CtxRequestID   = "REQUEST_ID"
	CtxLogger      = "LOGGER"
	CtxTokenSource = "TOKEN_SOURCE"
	CtxAuthClaims  = "AUTH_CLAIMS"
	RoleRoot       = "vh_root" // kong service clients has this role as well to allow inter-service communication
	RoleAdmin      = "vh_admin"
)

// This gets set at build time via `-ldflags "-X ..."`
var GitSHA string = "local"
