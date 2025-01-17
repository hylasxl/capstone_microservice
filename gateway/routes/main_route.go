package routes

import (
	"github.com/gorilla/mux"
)

func InitializeRoutes(router *mux.Router, clients *ServiceClients) {
	InitializeAuthRoutes(router, clients)
	InitializeUserRoutes(router, clients)
	InitializeFriendRoutes(router, clients)
	InitializePostRoutes(router, clients)
	InitializePingRoutes(router)
}
