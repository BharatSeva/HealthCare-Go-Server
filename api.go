package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

type Storage interface {
	SignUp(*User) error
	LoginUser(*Login) (*User, error)
}

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listen string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listen,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/signup", (makeHTTPHandlerFunc(s.SignUp)))
	router.HandleFunc("/login", (makeHTTPHandlerFunc(s.LoginUser)))
	router.HandleFunc("/uploadresume", withJWTAuth(makeHTTPHandlerFunc(s.UploadResume), s.store))

	log.Println("HealthCare_Server running on port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	req := User{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	user, err := SignUpAccount(req.Name, req.Email, req.Address, req.UserType, req.ProfileHeadline, req.PasswordHash)
	if err != nil {
		return err
	}
	if err := s.store.SignUp(user); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, user)
}

func (s *APIServer) LoginUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}

	login := &Login{}
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		return err
	}

	if _, err := s.store.LoginUser(login); err != nil {
		return err
	}

	// Create New Token everytime user login
	tokenString, err := createJWT(login)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) UploadResume(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed")
	}
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return err
	}

	// Retrieve the file from the form
	file, handler, err := r.FormFile("resume")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return err
	}
	defer file.Close()

	// Define the path where the file will be saved
	uploadPath := "./uploads"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		http.Error(w, "Unable to create upload directory", http.StatusInternalServerError)
		return err
	}

	// Create the file on the server
	dst, err := os.Create(filepath.Join(uploadPath, handler.Filename))
	if err != nil {
		http.Error(w, "Unable to create the file", http.StatusInternalServerError)
		return err
	}
	defer dst.Close()

	// Copy the uploaded file data to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Unable to save the file", http.StatusInternalServerError)
		return err
	}

	// Respond to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s", handler.Filename)
	return nil
}
///////////////////////////////////////////
////////  	JSON Token 	/////////////////////////

func createJWT(account *Login) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":    1500,
		"accountEmail": account.Email,
	}
	signKey := "PASSWORD"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Calling JWT handler")

		tokenString := r.Header.Get("auth")
		token, err := validateJWT(tokenString)
		if err != nil {
			writeJSON(w, http.StatusNotAcceptable, apiError{Error: "invalid token"})
			return
		}

		if !token.Valid {
			writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
			return
		}

		// userId, err := getId(r)
		// if err != nil {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }

		// account, err := s.GetAccountByID(userId)
		// if err != nil {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }

		// claims := token.Claims.(jwt.MapClaims)
		// fmt.Println(claims, account.Number)
		// if account.Number != claims["accountEmail"] {
		// 	writeJSON(w, http.StatusForbidden, apiError{Error: "invalid token"})
		// 	return
		// }
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := "PASSWORD"
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

// Helper One
// //////////////////////////////////////////
func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error
type apiError struct {
	Error string `json:"error"`
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return id, fmt.Errorf("invalid id given %s", idStr)
	}
	return id, nil
}
