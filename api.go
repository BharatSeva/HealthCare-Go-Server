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
	mod "vaibhavyadav-dev/healthcareServer/databases"
)

type contextKey string

const (
	contextKeyHealthCareID = contextKey("healthcareID")
)

type Storage interface {
	SignUpAccount(*mod.HIPInfo) (int64, error)
	LoginUser(*mod.Login) (*mod.HIPInfo, error)
	ChangePreferance(int, *mod.ChangePreferance) error
	GetPreferance(int) (*mod.ChangePreferance, error)
}

/////////////////////////////////////////
///// MongoDB Server ///////

type mongodb interface {
	GetAppointments(int) ([]*mod.Appointments, error)
	CreatePatient_bioData(int, *mod.PatientDetails) (*mod.PatientDetails, error)
	GetPatient_bioData(string) (*mod.PatientDetails, error)
}
type MONGODB struct {
	listenAddr string
	store      mongodb
}

func NewMONOGODB_SERVER(listen string, store mongodb) *MONGODB {
	return &MONGODB{
		listenAddr: listen,
		store:      store,
	}
}

func (m *MONGODB) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/healthcare/getappointments", withJWTAuth(makeHTTPHandlerFunc(m.GetAppointments)))
	router.HandleFunc("/api/v1/healthcare/createpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(m.CreatePatient_bioData)))
	router.HandleFunc("/api/v1/healthcare/getpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(m.GetpatientBioData)))
	log.Println("MONGODB_HealthCare Server running on Port: ", m.listenAddr)
	http.ListenAndServe(m.listenAddr, router)
}

/////////////////////////////////////////5
// MongoDB Server ///////

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
	router.HandleFunc("/api/v1/healthcare/changepreferance", withJWTAuth(makeHTTPHandlerFunc(s.ChangePreferance)))
	router.HandleFunc("/api/v1/healthcare/getpreferance", withJWTAuth(makeHTTPHandlerFunc(s.GetPreferance)))
	router.HandleFunc("/api/v1/healthcare/deleteaccount", withJWTAuth(makeHTTPHandlerFunc(s.DeleteAccount)))

	log.Println("HealthCare Server running on Port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := mod.HIPInfo{}
	// validate := validator.New()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	// err := validate.Struct(req)
	// if err != nil {
	// 	return fmt.Errorf("validation failed")
	// }

	user, err := mod.SignUpAccount(&req)
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

	login := &mod.Login{}
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
	req := &mod.ChangePreferance{}
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
	req := &mod.ChangePreferance{}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	req, err := s.store.GetPreferance(int(healthcareID))
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, req)
}

func (s *APIServer) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "DELETE" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := &mod.ChangePreferance{}
	req.Email = ""
	req.IsAvailable = false
	req.Scheduled_deletion = true
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}

	err := s.store.ChangePreferance(int(healthcareID), req)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]string{"status": "account deletion scheduled, contact to tron21vaibhav@gmail.com to remove deletion ASAP."})
}

/////////////////////////////// MONGODB METHODS GOES HERE //////////////////////////////////

func (m *MONGODB) GetAppointments(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "1HealthID not found in token"})
	}
	appointments, err := m.store.GetAppointments(int(healthcareID))
	if err != nil {
		log.Fatal("Error retrieving appointments:", err)
	}
	// Print retrieved appointments
	return writeJSON(w, http.StatusOK, appointments)
}

func (m *MONGODB) CreatePatient_bioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	patient := &mod.PatientDetails{}
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		return err
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "1HealthID not found in token"})
	}

	patientDetails, err := m.store.CreatePatient_bioData(int(healthcareID), patient)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, patientDetails)
}

func (m *MONGODB) GetpatientBioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	query := r.URL.Query()

    // Get the healthID from the query parameters
    healthID := query.Get("healthID")
	if healthID == "" {
		http.Error(w, "Missing healthID in URL", http.StatusBadRequest)
		return fmt.Errorf("missing healthID in URL")
	}
	patientDetails, err := m.store.GetPatient_bioData(healthID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, patientDetails)
}

// ///////////////////////////// ///////////////////// ///////////////// //////////// /////////////// ////////////// /
/////////////////////////// ///  	 Utility   	///////////////////////// ////////////////// ///////////// ///////

func createJWT(account *mod.Login) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":    1500,
		"healthcareID": account.HealthcareID,
	}
	signKey := "PASSWORD"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

func withJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
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
