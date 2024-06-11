package common

const (
	ServiceName = "vh-srv-events"

	CtxRequestID = "REQUEST_ID"
	CtxLogger    = "LOGGER"
)

// This gets set at build time via `-ldflags "-X ..."`
var GitSHA string = "local"
