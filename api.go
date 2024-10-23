package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	mod "vaibhavyadav-dev/healthcareServer/databases"
)

type contextKey string

const (
	contextKeyHealthCareID = contextKey("healthcareID")
)

// type postgres interface {
// 	SignUpAccount(*mod.HIPInfo) (int64, error)
// 	LoginUser(*mod.Login) (*mod.HIPInfo, error)
// 	ChangePreferance(int, *mod.ChangePreferance) error
// 	GetPreferance(int) (*mod.ChangePreferance, error)
// }

/////////////////////////////////////////
///// APIServer Server ///////

// type mongodb interface {
// 	GetAppointments(int) ([]*mod.Appointments, error)
// 	CreatePatient_bioData(int, *mod.PatientDetails) (*mod.PatientDetails, error)
// 	GetPatient_bioData(string) (*mod.PatientDetails, error)
// 	GetHealthcare_details(int) (*mod.HIPInfo, error)
// 	CreatepatientRecords(string, *mod.PatientRecords) (*mod.PatientRecords, error)
// 	GetPatientRecords(string, int) (*[]mod.PatientRecords, error)
// 	UpdatePatientBioData(string, map[string]interface{}) (*mod.PatientDetails, error)
// }
// type MONGODB struct {
// 	listenAddr string
// 	store      mongodb
// }

// func NewMONOGODB_SERVER(listen string, store mongodb) *MONGODB {
// 	return &MONGODB{
// 		listenAddr: listen,
// 		store:      store,
// 	}
// }

// func (m *MONGODB) Run() {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/api/v1/healthcare/getappointments", withJWTAuth(makeHTTPHandlerFunc(m.GetAppointments)))
// 	router.HandleFunc("/api/v1/healthcare/createpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(m.CreatePatient_bioData)))
// 	router.HandleFunc("/api/v1/healthcare/getpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(m.GetpatientBioData)))
// 	router.HandleFunc("/api/v1/healthcare/getdetails", withJWTAuth(makeHTTPHandlerFunc(m.GetDetails)))
// 	router.HandleFunc("/api/v1/healthcare/createrecords", withJWTAuth(makeHTTPHandlerFunc(m.CreatepatientRecords)))
// 	router.HandleFunc("/api/v1/healthcare/getpatientrecords", withJWTAuth(makeHTTPHandlerFunc(m.GetPatientRecords)))
// 	router.HandleFunc("/api/v1/healthcare/updatepatientbiodata", withJWTAuth(makeHTTPHandlerFunc(m.UpdatePatientBioData)))

// 	log.Println("MONGODB_HealthCare Server running on Port: ", m.listenAddr)
// 	http.ListenAndServe(m.listenAddr, router)
// }

// ///////////////////////////////////////5
// MongoDB Server ///////

type Store interface {
	// PostgreSQL Methods goes here...
	SignUpAccount(*mod.HIPInfo) (int64, error)
	LoginUser(*mod.Login) (*mod.HIPInfo, error)
	ChangePreferance(string, map[string]interface{}) error
	GetPreferance(int) (*mod.ChangePreferance, error)

	// MongoDB methods goes here...
	GetAppointments(string, int) ([]*mod.Appointments, error)
	CreatePatient_bioData(int, *mod.PatientDetails) (*mod.PatientDetails, error)
	GetPatient_bioData(string) (*mod.PatientDetails, error)
	CreateHealthcare_details(*mod.HIPInfo) (*mod.HIPInfo, error)
	GetHealthcare_details(string) (*mod.HIPInfo, error)
	CreatepatientRecords(string, *mod.PatientRecords) (*mod.PatientRecords, error)
	GetPatientRecords(string, int) (*[]mod.PatientRecords, error)
	UpdatePatientBioData(string, map[string]interface{}) (*mod.PatientDetails, error)
}

type APIServer struct {
	listenAddr string
	store      Store
}

func NewAPIServer(listen string, store Store) *APIServer {
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

	router.HandleFunc("/api/v1/healthcare/getappointments", withJWTAuth(makeHTTPHandlerFunc(s.GetAppointments)))
	router.HandleFunc("/api/v1/healthcare/createpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(s.CreatePatient_bioData)))
	router.HandleFunc("/api/v1/healthcare/getpatientbiodata", withJWTAuth(makeHTTPHandlerFunc(s.GetpatientBioData)))
	router.HandleFunc("/api/v1/healthcare/details", withJWTAuth(makeHTTPHandlerFunc(s.GetHealthcare_details)))
	router.HandleFunc("/api/v1/healthcare/createrecords", withJWTAuth(makeHTTPHandlerFunc(s.CreatepatientRecords)))
	router.HandleFunc("/api/v1/healthcare/getpatientrecords", withJWTAuth(makeHTTPHandlerFunc(s.GetPatientRecords)))
	router.HandleFunc("/api/v1/healthcare/updatepatientbiodata", withJWTAuth(makeHTTPHandlerFunc(s.UpdatePatientBioData)))

	log.Println("HealthCare Server running on Port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	req := mod.HIPInfo{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	user, err := mod.SignUpAccount(&req)
	if err != nil {
		return err
	}
	_, err = s.store.SignUpAccount(user)
	if err != nil {
		return err
	}

	_, err = s.store.CreateHealthcare_details(user)
	if err != nil {
		return err
	}
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
	if r.Method != "PATCH" {
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	req := make(map[string]interface{})
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	healthcareID_int := int(healthcareID)
	healthcareID_str := string(healthcareID_int)
	err = s.store.ChangePreferance(healthcareID_str, req)
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
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	req := make(map[string]interface{})
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	healthcareID_int := int(healthcareID)
	healthcareID_str := string(healthcareID_int)
	err := s.store.ChangePreferance(healthcareID_str, req)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]string{"status": "account deletion scheduled, contact to tron21vaibhav@gmail.com to remove deletion ASAP."})
}

/////////////////////////////// MONGODB METHODS GOES HERE //////////////////////////////////

func (s *APIServer) GetAppointments(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(float64)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	query := r.URL.Query()
	listStr := query.Get("list")
	list := 5
	if listStr != "" {
		var err error
		list, err = strconv.Atoi(listStr)
		if err != nil {
			log.Fatalf("Failed to convert list to int: %v pagination %d", err, list)
		}
		fmt.Println(list)
	}
	healthcareID_int := int(healthcareID)
	healthcareID_str := string(healthcareID_int)
	appointments, err := s.store.GetAppointments(healthcareID_str, list)
	if err != nil {
		log.Fatal("Error retrieving appointments:", err)
	}
	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "fetch successful",
		"appointments": appointments,
		"pagination":   list,
	})
}

func (s *APIServer) CreatePatient_bioData(w http.ResponseWriter, r *http.Request) error {
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

	patientDetails, err := s.store.CreatePatient_bioData(int(healthcareID), patient)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusCreated, patientDetails)
}

func (s *APIServer) GetpatientBioData(w http.ResponseWriter, r *http.Request) error {
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
	patientDetails, err := s.store.GetPatient_bioData(healthID)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, patientDetails)
}

func (s *APIServer) GetHealthcare_details(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	fmt.Println(healthcareID)
	if !ok {
		return writeJSON(w, http.StatusBadRequest, map[string]string{"HealthCareID": "HealthCareID not found in token"})
	}
	hipdetails, err := s.store.GetHealthcare_details(healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"message":    "Not Found!",
			"healthcare": hipdetails,
		})
	}
	return writeJSON(w, http.StatusOK, hipdetails)
}

func (s *APIServer) CreatepatientRecords(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method is not allowed %s", r.Method)
	}
	patientrecords := &mod.PatientRecords{}
	err := json.NewDecoder(r.Body).Decode(&patientrecords)
	if err != nil {
		return writeJSON(w, http.StatusNoContent, map[string]interface{}{
			"message": err,
		})
	}
	healthcareId, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "StatusUnauthorized"})
	}
	patientrecords_created, err := s.store.CreatepatientRecords(healthcareId, patientrecords)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": err,
		})
	}
	return writeJSON(w, http.StatusCreated, patientrecords_created)
}

func (s *APIServer) GetPatientRecords(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("%s method not allowed", r.Method)
	}
	query := r.URL.Query()
	health_id := query.Get("healthID")
	if health_id == "" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Health Id not Provided",
		})
	}
	listStr := query.Get("list")
	list := 5
	if listStr != "" {
		var err error
		list, err = strconv.Atoi(listStr)
		if err != nil {
			return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"message": "Could not fetch Records",
			})
		}
	}
	patientRecords, err := s.store.GetPatientRecords(health_id, list)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Could not fetch Records",
		})
	}
	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "successfull",
		"pagination":      list,
		"patient_records": patientRecords,
	})
}

func (s *APIServer) UpdatePatientBioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "PATCH" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	healthID := r.URL.Query().Get("healthID")
	if healthID == "" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Provide health Id",
		})
	}

	updates := make(map[string]interface{})
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error: could not get data",
		})
	}

	updatedPatient, err := s.store.UpdatePatientBioData(healthID, updates)
	if err != nil {
		return writeJSON(w, http.StatusConflict, map[string]interface{}{
			"message": "No Chage Detected or No Patient Found",
		})
	}
	return writeJSON(w, http.StatusAccepted, updatedPatient)
}

// ///////////////////////////// ///////////////////// ///////////////// //////////// /////////////// ////////////// /
/////////////////////////// ///  	 Utility Functions  	///////////////////////// ////////////////// ///////////// ///////

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
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			healthID, _ := claims["healthcareID"].(string)
			ctx := context.WithValue(r.Context(), contextKeyHealthCareID, healthID)
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
