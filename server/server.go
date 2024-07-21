package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/____reserved/privacy_policy", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I will never look at your content. The only thing this chrome extension will do is expand mo/ to https://www.molinks.me/. Everything else is managed within https://www.molinks.me/.")
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This site is a WIP, it not yet ready for public use.")
	})
	http.ListenAndServe(":3003", nil)
}
