package main

import (
	"mo_links/auth"
	"mo_links/db"
	"mo_links/http_server"
)

func main() {
	db.InitializeDB()
	auth.InitAuth()
	http_server.StartServer()
}
