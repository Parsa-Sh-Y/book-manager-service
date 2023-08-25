package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"

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

type tableOfContents struct {
	Contents []string `json:"table_of_contents"`
}

type bookCollection struct {
	Books *[]models.Book `json:"books"`
}

type updateRequestBody struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type respone struct {
	Message string `json:"message"`
}

func (r *respone) json() []byte {
	res, _ := json.Marshal(r.Message)
	return res
}

func CreateNewServer(conf config.Config) *Server {

	// Setup the logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})

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

	// check if request body is empty
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error reading the request body")
		return
	}

	var cred auth.UserCredentials
	// parse the request body
	err = json.Unmarshal(body, &cred)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("could not parse the request body")
		return
	}

	token, err := s.auth.Login(&cred)
	if err == db.ErrUserNotFound {
		respone, err := json.Marshal(map[string]interface{}{"message": "no such username exists"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.WithError(err).Error("error trying to marshal the respone message")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respone)
	} else if err == auth.ErrIncorrectPassword {
		respone, err := json.Marshal(map[string]interface{}{"message": "incorrect password"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.WithError(err).Error("error trying to marshal the respone message")
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(respone)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error while logging in the user")
		return
	}

	respone, err := json.Marshal(map[string]string{"access_token": token})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error trying to marshal respone message")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respone)
}

func (s *Server) HandleBooksRoot(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		s.HandleCreateBook(w, r)
	case http.MethodGet:
		s.HandleGetAllBooks(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (s *Server) HandleBooksSubtree(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodDelete:
		s.HandleDelete(w, r)
	case http.MethodGet:
		s.HandleGetBook(w, r)
	case http.MethodPut:
		s.HandleUpdate(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}

func (s *Server) HandleCreateBook(w http.ResponseWriter, r *http.Request) {

	// check if method is POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")

	username, err := s.auth.GetUsernameByToken(token)
	if err != nil {
		if err == auth.ErrCanNotValidateToken {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.WithError(err).Error("error validating user token")
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			s.logger.WithError(err).Warn("the was a problem with the token provided")
			return
		}
	}

	account, err := s.db.GetUserByUsername(username)
	if err == db.ErrUserNotFound {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error retrieving the user from database")
		return
	}

	var book models.Book
	var table tableOfContents
	// check if request body is empty
	if r.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error reading request body")
		return
	}

	err = json.Unmarshal(reqData, &table)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("there was an error when parsing the request body(table of contents)")
		return
	}
	err = json.Unmarshal(reqData, &book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("there was an error when parsing the request body(book)")
		return
	}
	book.UserID = account.ID // set the use who made the request as the owner of the book

	// add each content to the book instance
	for _, content := range table.Contents {
		book.TableOfContents = append(book.TableOfContents, models.Content{ContentName: content})
	}

	err = s.db.CreateBook(&book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	message := map[string]interface{}{
		"message": "book was created successfully",
	}

	respone, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error trying to marshal the respone message")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respone)
}

func (s *Server) HandleGetBook(w http.ResponseWriter, r *http.Request) {

	// check if method is get
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// check if user is logged in
	token := r.Header.Get("Authorization")
	_, err := s.auth.GetUsernameByToken(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("could not log in the user")
		return
	}

	bookID, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the book from the database
	book, err := s.db.GetBook(bookID)
	if err != nil {
		// TODO : check for different errors
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Populate TableOfContentsJson field
	for _, content := range book.TableOfContents {
		book.TableOfContentsJson = append(book.TableOfContentsJson, content.ContentName)
	}

	// Create the response
	response, err := json.Marshal(&book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error while trying to marshal a requested book")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func (s *Server) HandleGetAllBooks(w http.ResponseWriter, r *http.Request) {

	// check if method is GET
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// check if user is logged in
	token := r.Header.Get("Authorization")
	_, err := s.auth.GetUsernameByToken(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("could not log in the user")
		return
	}

	var books bookCollection
	// get all books from the database
	books.Books, err = s.db.GetAllBooks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error retrieving all the books from the database")
		return
	}

	// create the respone
	respone, err := json.Marshal(&books)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error trying to marshal the respone")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respone)
}

func (s *Server) HandleDelete(w http.ResponseWriter, r *http.Request) {

	// check if method is delete
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// check if user is logged in
	token := r.Header.Get("Authorization")
	username, err := s.auth.GetUsernameByToken(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("could not log in the user")
		return
	}

	// get the book id
	bookID, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// delete the book
	err = s.db.DeleteUserBook(username, uint(bookID))
	var res respone
	if err == db.ErrBookNotFound || err == db.ErrUserNotFound || err == db.ErrPermissionDenied {
		res.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.json())
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error deleting a book")
		return
	} else {
		res.Message = "Book was deleted successfully"
		w.WriteHeader(http.StatusOK)
		w.Write(res.json())
	}
}

func (s *Server) HandleUpdate(w http.ResponseWriter, r *http.Request) {

	// check if method is PUT
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// check if user is logged in
	token := r.Header.Get("Authorization")
	username, err := s.auth.GetUsernameByToken(token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.WithError(err).Warn("could not log in the user")
		return
	}

	// get book id
	bookID, err := strconv.Atoi(path.Base(r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get the request body

	var reqBody updateRequestBody

	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("can not read request body")
		return
	}
	err = json.Unmarshal(reqData, &reqBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error trying to parse request body")
		return
	}

	// update the book
	err = s.db.UpdateUserBook(username, uint(bookID), reqBody.Name, reqBody.Category)
	var res respone
	if err == db.ErrBookNotFound || err == db.ErrUserNotFound || err == db.ErrPermissionDenied {
		res.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
		w.Write(res.json())
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.WithError(err).Error("error updating a book")
		return
	} else {
		res.Message = "Book was updated successfully"
		w.WriteHeader(http.StatusOK)
		w.Write(res.json())
	}
}
