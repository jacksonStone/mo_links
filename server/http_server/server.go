package http_server

import (
	"log"
	"net/http"
)

func StartServer() {
	initializeRoutes()
	err := http.ListenAndServe(":3003", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func initializeRoutes() {
	initializeInvitesRoute()
	initializeStaticRoutes()
	initializeAuthRoutes()
	initializeOrganizationRoutes()
	initializeUserDetailRoutes()
	initializeLinkRoutes()
}
