package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	mod "vaibhavyadav-dev/healthcareServer/databases"

	_ "github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
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
	GetPreferance(string) (*mod.ChangePreferance, error)
	GetTotalRequestCount(string) (int, error)
	// counters goes here.....
	// Recordsviewed_counter(string) error
	// Recordscreated_counter(string) error
	// Patientbiodata_created_counter(string) error
	// Patientbiodata_viewed_counter(string) error

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// MongoDB methods goes here...
	GetAppointments(string, int) ([]*mod.Appointments, error)
	CreatePatient_bioData(string, *mod.PatientDetails) (*mod.PatientDetails, error)
	GetPatient_bioData(string) (*mod.PatientDetails, error)
	CreateHealthcare_details(*mod.HIPInfo) (*mod.HIPInfo, error)
	GetHealthcare_details(string) (*mod.HIPInfo, error)
	CreatepatientRecords(string, *mod.PatientRecords) (*mod.PatientRecords, error)
	GetPatientRecords(string, int) (*[]mod.PatientRecords, error)
	UpdatePatientBioData(string, map[string]interface{}) (*mod.PatientDetails, error)

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// rabbitmq methods goes here...
	Push_logs(interface{}, interface{}, interface{}, interface{}, interface{}) error
	Push_appointment(category string) error
	Push_patient_records(map[string]interface{}) error
	Push_patientbiodata(map[string]interface{}) error
	Push_counters(string, string) error

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// redis implementation Goes here
	Set(string, interface{}) error
	Get(string) (interface{}, error)
	Close() error
	// rate limiter goes here...
	IsAllowed(string) (bool, error)
	IsAllowed_leaky_bucket(string) (bool, error)
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

	// this one will serve from postgres
	router.HandleFunc("/api/v1/healthcare/getpreferance", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetPreferance))))
	router.HandleFunc("/api/v1/healthcare/changepreferance", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.ChangePreferance))))
	router.HandleFunc("/api/v1/healthcare/deleteaccount", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.DeleteAccount))))

	// this is will server from mongodb
	router.HandleFunc("/api/v1/healthcare/getappointments", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetAppointments))))
	router.HandleFunc("/api/v1/healthcare/createpatientbiodata", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.CreatePatient_bioData))))
	router.HandleFunc("/api/v1/healthcare/getpatientbiodata", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetpatientBioData))))
	router.HandleFunc("/api/v1/healthcare/details", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetHealthcare_details))))
	router.HandleFunc("/api/v1/healthcare/createrecords", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.CreatepatientRecords))))
	router.HandleFunc("/api/v1/healthcare/getpatientrecords", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetPatientRecords))))
	router.HandleFunc("/api/v1/healthcare/updatepatientbiodata", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.UpdatePatientBioData))))

	log.Println("HealthCare Server running on Port: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "Method Not Allowed",
		})
	}

	req := mod.HIPInfo{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}
	user, err := mod.SignUpAccount(&req)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	_, err = s.store.SignUpAccount(user)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "User already exists",
		})
	}

	_, err = s.store.CreateHealthcare_details(user)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "User already exists",
		})
	}

	// GET IP Addrress of user
	// for logging and monitering purpose only, this will help you to
	// moniter your account
	ip := r.Header.Get("X-Forwarded-For")
	// get from header if empty
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	// get from rmote address
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	// send Email to healthcare that his account has been created now
	err = s.store.Push_logs("hip:account_created", user.HealthcareName, user.Email, ip, user.HealthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	return writeJSON(w, http.StatusCreated, map[string]interface{}{
		"status": "Successfully Created",
		"Healthcare_details": map[string]interface{}{
			"healthcare_id":      user.HealthcareID,
			"healthcare_license": user.HealthcareLicense,
			"name":               user.HealthcareName,
			"email":              user.Email,
		},
	})
}

func (s *APIServer) LoginUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "Method Not Allowed",
		})
	}

	login := &mod.Login{}
	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "could not process your request please check your schema",
		})
	}

	hip, err := s.store.LoginUser(login)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "No user Found!",
		})
	}
	// GET IP Addrress of user
	// for logging and monitering purpose only, this will help you to
	// moniter your account
	ip := r.Header.Get("X-Forwarded-For")
	// get from header if empty
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	// get from rmote address
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	// Notify user everytime user login !
	s.store.Push_logs("hip:account_login", hip.HealthcareName, hip.Email, ip, hip.HealthcareID)
	// check quota limit
	// from sql database first
	count, err := s.store.GetTotalRequestCount(login.HealthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}
	// if count of request limit reached don't allow user to login
	// limit the user
	if count <= 0 {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "Your Request Quota Has been reached",
			"status":  "Quota Limit Reached (mail 21vaibhav11@gmail.com to increase the limit)",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hip.Password), []byte(login.Password)); err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "password mismatched",
		})
	}

	// create token everytime user login !!
	tokenString, err := createJWT(hip)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"Expires In":      "5d",
		"token":           tokenString,
		"healthcare_id":   hip.HealthcareID,
		"healthcare_name": hip.HealthcareName,
	})
}

func (s *APIServer) ChangePreferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "PATCH" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	req := make(map[string]interface{})
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Error reading request body: " + err.Error(),
		})
	}
	err = s.store.ChangePreferance(healthcareID, req)
	if err != nil {
		return writeJSON(w, http.StatusGatewayTimeout, map[string]interface{}{
			"error": err.Error(),
		})
	}
	return writeJSON(w, http.StatusOK, req)
}

func (s *APIServer) GetPreferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	req := &mod.ChangePreferance{}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}

	// check if cache are needed or not
	query := r.URL.Query()
	cache := query.Get("cache")
	if cache == "" {
		cache = "true"
	}
	// Fetch from redis server first
	if cache == "true" {
		fetched, err := s.store.Get("hip:pref:" + healthcareID)
		if err != redis.Nil {
			if err != nil {
				return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"error": err.Error(),
				})
			}
			if fetched != nil {
				fetchedStr, ok := fetched.(string)
				if !ok {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to convert data to string",
					})
				}
				var jsonBody *mod.ChangePreferance
				err = json.Unmarshal([]byte(fetchedStr), &jsonBody)
				if err != nil {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to parse the data",
					})
				}
				return writeJSON(w, http.StatusOK, map[string]interface{}{
					"preferance": jsonBody,
				})
			}
		}
	}

	// fetch from database
	req, err := s.store.GetPreferance(healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	// Store into redis
	err = s.store.Set("hip:pref:"+healthcareID, req)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"preferance": req,
	})
}

func (s *APIServer) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "DELETE" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	req := map[string]interface{}{
		"scheduled_deletion": true,
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	err := s.store.ChangePreferance(healthcareID, req)
	if err != nil {
		return writeJSON(w, http.StatusNotImplemented, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// send email to user
	s.store.Push_logs("hip:delete_account", nil, nil, nil, healthcareID)

	return writeJSON(w, http.StatusOK, map[string]string{"status": "account deletion scheduled, contact to tron21vaibhav@gmail.com to remove deletion ASAP."})
}

/////////////////////////////// MONGODB METHODS GOES HERE //////////////////////////////////

func (s *APIServer) GetAppointments(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
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
			return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"message": "Server error: " + err.Error(),
			})
		}
	}
	appointments, err := s.store.GetAppointments(healthcareID, list)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"error": "error: " + err.Error(),
		})
	}
	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"appointments": appointments,
		"pagination":   list,
	})
}

func (s *APIServer) CreatePatient_bioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	patient := &mod.PatientDetails{}
	err := json.NewDecoder(r.Body).Decode(&patient)
	if err != nil {
		return err
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	patientDetails, err := mod.CreatePatient_bioData(healthcareID, patient)
	if err != nil {
		return err
	}

	patientDetails, err = s.store.CreatePatient_bioData(healthcareID, patientDetails)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "User Already exists",
		})
	}

	// Push this into rabbitmq instead of directly into database (mongodDB)
	// Push into rabbitmq
	// body := map[string]interface{}{
	// 	"biodata": patientDetails,
	// }
	// err = s.store.Push_patientbiodata(body)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }
	// Notify user via email
	s.store.Push_logs("hip:patient_biodata_created", patientDetails.FirstName, patientDetails.Email, patientDetails.HealthcareID, healthcareID)

	// counters
	// err = s.store.Push_counters("hip:patientbiodata_created_counter", patientDetails.HealthcareID)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }

	return writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":          "data has been successfully created!",
		"status":           "created",
		"patient_healthID": patientDetails.HealthID,
		"Full name":        patientDetails.FirstName + " " + patientDetails.MiddleName + " " + patientDetails.LastName,
	})
}

func (s *APIServer) GetpatientBioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
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
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "something went wrong from our side :(",
		})
	}
	// Notify user via email
	s.store.Push_logs("hip:patient_biodata_viewed", patientDetails.FirstName, patientDetails.Email, patientDetails.HealthID, healthcareID)
	// counters
	// err = s.store.Push_counters("hip:patientbiodata_viewed_counter", patientDetails.HealthcareID)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }

	return writeJSON(w, http.StatusOK, patientDetails)
}

func (s *APIServer) GetHealthcare_details(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusBadRequest, map[string]string{"HealthCareID": "HealthCareID not found in token"})
	}

	// check if cache are needed or not
	query := r.URL.Query()
	cache := query.Get("cache")
	if cache == "" {
		cache = "true"
	}
	// Fetch from redis server first
	if cache == "true" {
		// Fetch from redis server first
		fetched, err := s.store.Get("hip:details:" + healthcareID)
		if err != redis.Nil {
			if err != nil {
				return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"error": err.Error(),
				})
			}
			if fetched != nil {
				fetchedStr, ok := fetched.(string)
				if !ok {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to convert data to string",
					})
				}
				var jsonBody *mod.HIPInfo
				err = json.Unmarshal([]byte(fetchedStr), &jsonBody)
				if err != nil {
					return err
				}
				return writeJSON(w, http.StatusOK, map[string]interface{}{
					"preferance": jsonBody,
				})
			}
		}
	}

	// fetch from database now!!
	hipdetails, err := s.store.GetHealthcare_details(healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"message":    "Not Found!",
			"healthcare": hipdetails,
		})
	}

	// Store into redis!!!
	err = s.store.Set("hip:details:"+healthcareID, hipdetails)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"healthcare": hipdetails,
	})
}

func (s *APIServer) CreatepatientRecords(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
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

	// pushing into database
	// Leave this for now
	// patientrecords_created, err := s.store.CreatepatientRecords(healthcareId, patientrecords)
	// if err != nil {
	// 	return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
	// 		"message": err,
	// 	})
	// }

	// Convert into body format
	body := map[string]interface{}{
		"record": patientrecords,
	}
	// Push it intoRabbitMq
	err = s.store.Push_patient_records(body)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
			"status":  "Server Could not Process your Request",
			"err":     err.Error(),
		})
	}

	// Notify user via email
	err = s.store.Push_logs("hip:patient_record_created", patientrecords.ID, nil, patientrecords.HealthID, healthcareId)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
			"status":  "Server Could not Process your Request",
			"err":     err.Error(),
		})
	}
	// counters
	// err = s.store.Push_counters("hip:recordscreated_counter", healthcareId)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }

	return writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "successfully Created",
		"status":  "pending ",
	})
}

func (s *APIServer) GetPatientRecords(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
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

	// Notify user via email
	healthcareId, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "StatusUnauthorized"})
	}

	s.store.Push_logs("hip:patient_record_viewed", nil, nil, health_id, healthcareId)

	// counters (Will be removed soon)
	// err = s.store.Push_counters("hip:recordsviewed_counter", healthcareId)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "successfull",
		"pagination":      list,
		"patient_records": patientRecords,
	})
}

func (s *APIServer) UpdatePatientBioData(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "PATCH" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
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
			"message": "Internal Server Error: could not process data",
		})
	}

	updatedPatient, err := s.store.UpdatePatientBioData(healthID, updates)
	if err != nil {
		return writeJSON(w, http.StatusConflict, map[string]interface{}{
			"message": "No Chage Detected or No Patient Found",
		})
	}
	// Get HealthCareId and update the patient
	s.store.Push_logs("hip:patient_biodata_updated", updatedPatient.FirstName, updatedPatient.Email, updatedPatient.HealthID, healthcareID)

	return writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"updated_details": updatedPatient,
	})
}

// ///////////////////////////// ///////////////////// ///////////////// //////////// /////////////// ////////////// /
/////////////////////////// ///  	 Utility Functions  	///////////////////////// ////////////////// ///////////// ///////

// rate limiter goes here...
func (s *APIServer) RateLimiter(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
		if !ok {
			writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid token"})
			return
		}
		allowed_fixed_window, err := s.store.IsAllowed(healthcareID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "Something bad happened from our side :("})
			return
		}
		if !allowed_fixed_window {
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"status":  "Request Blocked",
				"message": "Too many request from your side, please login again",
			})
			return
		}
		// only check for at max 10,000 request per second at any given time
		// this one checks for leaky bucket rate-limiting
		allowed_leaky_bucket, err := s.store.IsAllowed_leaky_bucket(healthcareID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, apiError{Error: "Something bad happened from our side :("})
			return
		}
		if !allowed_leaky_bucket {
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"status":  "Request Blocked",
				"message": "Too many request from your side, we've suspended all of your for 5 minutes",
			})
			return
		}

		// request counters
		/////////////////////////////////////////////////////////////////////////////
		/////////////////////////////////////////////////////////////////////////////
		// err = s.store.Push_counters("hip:requestcounter", healthcareID)
		// if err != nil {
		// 	writeJSON(w, http.StatusInternalServerError, apiError{Error: "Something bad happened from our side :("})
		// 	return
		// }
		/////////////////////////////////////////////////////////////////////////////
		/////////////////////////////////////////////////////////////////////////////

		handlerFunc(w, r)
	}
}

func createJWT(account *mod.HIPInfo) (string, error) {
	claims := jwt.MapClaims{
		"expiresAt":    time.Now().Add(5 * 24 * time.Hour).Unix(), //setting it to 5days from now
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
			healthcareID, _ := claims["healthcareID"].(string)
			// block the request if token tempered
			if healthcareID == "" {
				writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid Token"})
				return
			}

			// validate ratelimiter

			ctx := context.WithValue(r.Context(), contextKeyHealthCareID, healthcareID)
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
