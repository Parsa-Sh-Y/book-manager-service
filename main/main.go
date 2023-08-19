package main

import (
	"log"

	"github.com/Parsa-Sh-Y/book-manager-service/config"
	"github.com/Parsa-Sh-Y/book-manager-service/db"
	"github.com/ilyakaznacheev/cleanenv"
)

func main() {

	var cfg config.Config
	cleanenv.ReadEnv(&cfg)

	db, err := db.CreateNewGormDB(cfg)
	if err != nil {
		log.Fatalln(err.Error())
	}

	err = db.CreateSchema()
	if err != nil {
		log.Fatalln(err.Error())
	}

}
