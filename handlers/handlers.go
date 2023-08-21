package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Parsa-Sh-Y/book-manager-service/auth"
	"github.com/Parsa-Sh-Y/book-manager-service/config"
	"github.com/Parsa-Sh-Y/book-manager-service/db"
	"github.com/Parsa-Sh-Y/book-manager-service/db/models"
	"github.com/sirupsen/logrus"
)

type Server struct {
	db     *db.GormDB
	logger *logrus.Logger
	auth   *auth.Auth
}

func CreateNewServer(conf config.Config) *Server {

	// Setup the logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true)
	logger.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	//logger.SetFormatter(&logrus.JSONFormatter{})

	// Create new instance of dg.DB
	gormDB, err := db.CreateNewGormDB(conf)
	if err != nil {
		logger.WithError(err).Fatal("error in connecting to the database")
	}
	logger.Infof("connected to the %s database", conf.Database.Name)

	// Create schema
	// Create any tables if they don't exits
	err = gormDB.CreateSchema()
	if err != nil {
		logger.WithError(err).Fatal("error in database migration")
	}
	logger.Infoln("migrate tables and models successfully")

	// Create authenticate
	auth, err := auth.NewAuth(gormDB, conf.JwtExpirationInMinutes)
	if err != nil {
		logger.WithError(err).Fatal("can not create the authenticate instance")
	}

	return &Server{
		db:     gormDB,
		logger: logger,
		auth:   auth,
	}

}

func (s *Server) HandleSignup(w http.ResponseWriter, r *http.Request) {

	// check if method is post
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// parse the request body
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.WithError(err).Warn("Can not read the request body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var user models.User

	err = json.Unmarshal(reqData, &user)
	if err != nil {
		s.logger.WithError(err).Warn("can not unmarshal the request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// add the user to the database
	// TODO : handle different errors individually
	err = s.db.CreateUser(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("can not create a new user")
		return
	}

	response := map[string]interface{}{
		"message": "user has been created",
	}
	resBody, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(resBody)

}

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {

	// check if mehtod is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var cred auth.UserCredentials
	// parse the request body
	err = json.Unmarshal(body, &cred)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := s.auth.Login(&cred)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	respone, err := json.Marshal(map[string]string{"access_token": token})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respone)
}
