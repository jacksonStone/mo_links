package main

import (
	"log"
	"mo_links/auth"
	"mo_links/db"
	"mo_links/routes"
	"net/http"
)

func main() {

	db.InitializeDB()
	auth.InitAuth()

	routes.InitializeInvitesRoute()
	routes.InitStaticRoutes()
	routes.InitAuthRoutes()
	routes.InitOrganizationRoutes()
	routes.InitUserDetailRoutes()
	routes.InitLinkRoutes()

	err := http.ListenAndServe(":3003", nil)
	if err != nil {
		log.Fatal(err)
	}
}
