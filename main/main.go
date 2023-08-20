package main

import (
	"net/http"

	"github.com/Parsa-Sh-Y/book-manager-service/config"
	"github.com/Parsa-Sh-Y/book-manager-service/handlers"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {

	var cfg config.Config
	cleanenv.ReadEnv(&cfg)

	server := handlers.CreateNewServer(cfg)

	http.HandleFunc("/api/v1/auth/signup", server.HandleSignup)

	http.ListenAndServe(":8080", nil)

}
