package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-playground/validator/v10"
)

type contextKey string

const (
	contextKeyHealthCareID = contextKey("healthcareID")
)

type Storage interface {
	SignUpAccount(*HIPInfo) (int64, error)
	LoginUser(*Login) (*HIPInfo, error)
	ChangePreferance(int, *ChangePreferance) error
	GetPreferance(int) (*ChangePreferance, error)
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
	router.HandleFunc("/api/v1/healthcareauth/register", (makeHTTPHandlerFunc(s.SignUp)))
	router.HandleFunc("/api/v1/healthcareauth/login", (makeHTTPHandlerFunc(s.LoginUser)))
	router.HandleFunc("/api/v1/healthcare/changepreferance", withJWTAuth(makeHTTPHandlerFunc(s.ChangePreferance), s.store))
	router.HandleFunc("/api/v1/healthcare/getpreferance", withJWTAuth(makeHTTPHandlerFunc(s.GetPreferance), s.store))

	log.Println("HealthCare Server running on Port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := HIPInfo{}
	// validate := validator.New()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	// err := validate.Struct(req)
	// if err != nil {
	// 	return fmt.Errorf("validation failed")
	// }

	fmt.Println(req)
	user, err := SignUpAccount(&req)
	if err != nil {
		return err
	}
	id, err := s.store.SignUpAccount(user)
	if err != nil {
		return err
	}
	user.HealthcareID = int(id)
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

	hip, err := s.store.LoginUser(login)
	if err != nil {
		return err
	}

	// Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(hip.Password), []byte(login.Password)); err != nil {
		return fmt.Errorf("password mismatched your id %s", hip.HealthcareLicense)
	}

	// create token everytime user login !!
	tokenString, err := createJWT(login)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, map[string]string{"token": tokenString})
}

func (s *APIServer) ChangePreferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := &ChangePreferance{}
	req.Email = ""
	req.IsAvailable = true
	req.Scheduled_deletion = false
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	err = s.store.ChangePreferance(int(healthcareID), req)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, req)
}

func (s *APIServer) GetPreferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := &ChangePreferance{}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	req, err := s.store.GetPreferance(int(healthcareID)); if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, req);
}

// ///////////////////////////// ///////////////////// ///////////////// //////////// /////////////// ////////////// /
/////////////////////////// ///  	 Utility   	///////////////////////// ////////////////// ///////////// ///////

func createJWT(account *Login) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":    1500,
		"healthcareID": account.HealthcareID,
	}
	signKey := "PASSWORD"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		// this will extract token from Bearer keyword
		if tokenString == "" || len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			writeJSON(w, http.StatusNotAcceptable, apiError{Error: "Authorization header format must be Bearer <token>"})
			return
		}
		tokenString = tokenString[7:]
		token, err := validateJWT(tokenString)
		if err != nil {
			writeJSON(w, http.StatusNotAcceptable, apiError{Error: fmt.Sprintf("Token Not Valid: %v", err)})
			return
		}

		if !token.Valid {
			writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid token"})
			return
		}
		// Extract claims and add to request context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			healthID, _ := claims["healthcareID"].(float64)
			// Add claims to request context
			ctx := context.WithValue(r.Context(), contextKeyHealthCareID, healthID)
			// Pass the new context with claims to the handler
			handlerFunc(w, r.WithContext(ctx))
		} else {
			writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid token claims"})
			return
		}
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
