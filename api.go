package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"time"

	mod "vaibhavyadav-dev/healthcareServer/databases"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/rs/cors"

	// for monitoring
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Client, Patient, User refer to same thing their names are used interchangebly !!!!

type contextKey string

const (
	contextKeyHealthCareID      = contextKey("healthcareID")
	contextKeyEmailHealthCareID = contextKey("healthcare_email")
	contextKeyHealthCareName    = contextKey("healthcare_name")
)

type Store interface {
	// PostgreSQL Methods goes here...
	SignUpAccount(*mod.HIPInfo) (int64, error)
	LoginUser(*mod.Login) (*mod.HIPInfo, error)
	ChangePreferance(string, map[string]interface{}) error
	GetPreferance(string) (*mod.Preferance, error)
	GetTotalRequestCount(string) (int, error)
	CreateClient_stats(string) error
	GetAppointments_postgres(health_id string, offset, limit int64) ([]*mod.Appointments, error)
	SetAppointments_postgres(healthcare_id, health_id, status string, id int64) (int64, error)
	Create_ClientProfile(*mod.PatientDetails) error
	Get_ClientProfile(string) (*mod.PatientDetails, error)
	Update_clientProfile(string, map[string]interface{}) (*mod.PatientDetails, error)
	GetHealthcare_details_postgres(string) (*mod.HIPInfo, error)

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// MongoDB methods goes here...

	// Some methods that are commented are deprecated and not used and maintained
	// they've shifted or migrated to other databases for optimizations (like postgres)

	CreatepatientRecords(string, *mod.PatientRecords) (*mod.PatientRecords, error)
	GetPatientRecords(string, string, int) (*[]mod.PatientRecords, error)
	// GetAppointments(string, int64) ([]*mod.Appointments, error)
	// SetAppointments(string, string, string, int64) (*mod.Appointments, error)
	// CreatePatient_bioData(string, *mod.PatientDetails) (*mod.PatientDetails, error)
	// GetPatient_bioData(string) (*mod.PatientDetails, error)
	// CreateHealthcare_details(*mod.HIPInfo) (*mod.HIPInfo, error)
	// GetHealthcare_details(string) (*mod.HIPInfo, error)
	// UpdatePatientBioData(string, map[string]interface{}) (*mod.PatientDetails, error)

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// Rabbitmq methods goes here...
	Push_logs(interface{}, interface{}, interface{}, interface{}, interface{}, interface{}) error
	Push_update_appointment(appointment map[string]interface{}) error
	Push_patient_records(map[string]interface{}) error
	Push_patientbiodata(map[string]interface{}) error
	Push_counters(string, string) error

	/////////////////////////////////////////////////////////////////////////////
	/////////////////////////////////////////////////////////////////////////////
	// Redis Implementation Goes here
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
	// Add Prometheus middleware to all routes
	router.Use(PrometheusMiddleware)
	router.Path("/metrics").Handler(promhttp.Handler())

	router.HandleFunc("/api/v1/healthcare/auth/register", (makeHTTPHandlerFunc(s.SignUp)))
	router.HandleFunc("/api/v1/healthcare/auth/login", (makeHTTPHandlerFunc(s.LoginUser)))

	// this one will serve from postgres
	router.HandleFunc("/api/v1/healthcare/preferance/get", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetPreferance))))
	router.HandleFunc("/api/v1/healthcare/preferance/change", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.Update_Preferance))))
	router.HandleFunc("/api/v1/healthcare/delete/account", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.DeleteAccount))))

	// this is will server from mongodb
	router.HandleFunc("/api/v1/healthcare/appointments/get", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetAppointments))))
	router.HandleFunc("/api/v1/healthcare/appointments/set", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.SetAppointments))))
	router.HandleFunc("/api/v1/healthcare/details", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetHealthcare_details))))

	router.HandleFunc("/api/v1/healthcare/client/records/create", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.CreatepatientRecords))))
	router.HandleFunc("/api/v1/healthcare/client/records/fetch", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.GetPatientRecords))))

	router.HandleFunc("/api/v1/healthcare/client/profile/create", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.Create_ClientProfile))))
	router.HandleFunc("/api/v1/healthcare/client/profile/get", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.Get_clientProfile))))
	router.HandleFunc("/api/v1/healthcare/client/profile/update", withJWTAuth(s.RateLimiter(makeHTTPHandlerFunc(s.UpdateClientProfile))))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap the router with CORS handler
	handler := c.Handler(router)

	log.Println("HealthCare Server running on Port: ", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, handler); err != nil {
		log.Fatal(err)
	}
}

func (s *APIServer) SignUp(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "Method Not Allowed",
		})
	}

	req := mod.HIPInfo{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
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

	// store in postgres !!
	_, err = s.store.SignUpAccount(user)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "User already exists",
			"err":     err.Error(),
		})
	}

	// store in mongoDB also !!
	// _, err = s.store.CreateHealthcare_details(user)
	// if err != nil {
	// 	return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
	// 		"message": "User already exists",
	// 		"err":     err.Error(),
	// 	})
	// }

	// GET IP Addrress of user
	// for logging and monitering purpose only, this will help to
	// moniter account
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
	err = s.store.Push_logs("hip_accountCreated", user.HealthcareName, user.Email, ip, user.HealthcareName, user.HealthcareID)
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

	// check for total_request
	ok, err := s.store.IsAllowed(login.HealthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something went wrong from our side",
		})
	}

	// block request if limit exceeded
	if !ok {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "Your request quota has been exhausted",
			"message": "Mail 21vaibhav11@gmail.com with your Id to increase your quota",
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
	err = s.store.Push_logs("hip_accountLogin", hip.HealthcareName, hip.Email, ip, hip.HealthcareName, hip.HealthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something went wrong from our side",
		})
	}
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

func (s *APIServer) Update_Preferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPatch {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}

	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}

	// Decode the request body into a map
	var req map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"message": "Error reading request body: " + err.Error(),
		})
	}

	// Define valid fields and their types
	validFields := map[string]func(interface{}) error{
		"email": func(value interface{}) error {
			str, ok := value.(string)
			if !ok || !isValidEmail(str) {
				return fmt.Errorf("invalid email format")
			}
			return nil
		},
		"isAvailable": func(value interface{}) error {
			_, ok := value.(bool)
			if !ok {
				return fmt.Errorf("isAvailable must be a boolean")
			}
			return nil
		},
		"scheduled_deletion": func(value interface{}) error {
			_, ok := value.(bool)
			if !ok {
				return fmt.Errorf("scheduled_deletion must be a boolean")
			}
			return nil
		},
	}

	// Validate and filter the request fields
	updates := make(map[string]interface{})
	for field, validator := range validFields {
		if value, exists := req[field]; exists {
			if err := validator(value); err != nil {
				return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
					"message": fmt.Sprintf("Validation error on field '%s': %s", field, err.Error()),
				})
			}
			updates[field] = value
		}
	}

	// No fields has been provided
	if len(updates) == 0 {
		return writeJSON(w, http.StatusBadRequest, map[string]string{
			"message": "No valid fields to update",
		})
	}

	// Perform the update in the postgresDB
	err = s.store.ChangePreferance(healthcareID, updates)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to update preferences: " + err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "Preferences updated successfully",
		"preferances": updates,
	})
}

func (s *APIServer) GetPreferance(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	pref := &mod.Preferance{}
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
				fetchedData, ok := fetched.(struct {
					Value string        `json:"value"`
					TTL   time.Duration `json:"ttl"`
				})

				if !ok {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to convert data to expected format",
					})
				}

				var jsonBody *mod.Preferance
				err = json.Unmarshal([]byte(fetchedData.Value), &jsonBody)
				if err != nil {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to parse the data",
					})
				}

				return writeJSON(w, http.StatusOK, map[string]interface{}{
					"preferance":         jsonBody,
					"refreshIn(seconds)": fetchedData.TTL.Seconds(),
				})
			}
		}
	}

	// fetch from database
	pref, err := s.store.GetPreferance(healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	// Store into redis
	err = s.store.Set("hip:pref:"+healthcareID, pref)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something mishappened from our side :)",
			"error":   err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"preferance": pref,
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
	email_healthcareID, ok := r.Context().Value(contextKeyEmailHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}
	err := s.store.ChangePreferance(healthcareID, req)
	if err != nil {
		return writeJSON(w, http.StatusNotImplemented, map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Send email to user
	err = s.store.Push_logs("hip_deleteAccount", healthcare_name, email_healthcareID, nil, healthcare_name, healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusNotImplemented, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "Account deletion scheduled",
		"message": "mail tron21vaibhav@gmail.com to remove deletion ASAP.",
	})
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
	listStr := query.Get("limit")
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
	appointments, err := s.store.GetAppointments_postgres(healthcareID, 0, int64(list))
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "Something went wrong from our side",
			"error":  "error: " + err.Error(),
		})
	}

	// print [] array always if appointis empty
	if appointments == nil {
		appointments = []*mod.Appointments{}
	}
	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"appointments": appointments,
		"fetched":      len(appointments),
	})
}

// Set status of appointments
func (s *APIServer) SetAppointments(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("%s method is not allowed", r.Method)
	}
	healthcareID, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}

	// append healthcare ID
	update := &mod.UpdateAppointment{}
	update.HealthcareID = healthcareID
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"err":     err.Error(),
			"message": "Internal Server Error: could not process data",
		})
	}

	if update.Status != "Confirmed" && update.Status != "Rejected" && update.Status != "Pending" && update.Status != "Not Available" {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "Invalid status. Status must be one of [\"Pending\", \"Confirmed\", \"Rejected\", \"Not Available\"]",
		})
	}

	// Validate the struct fields
	validate := validator.New()
	err = validate.Struct(update)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"status":  "Constraints Violeted",
			"message": "Invalid data provided, please check your payload",
		})
	}

	// Check if struct fields are populated (non-zero values)
	val := reflect.ValueOf(update).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)
		if value.IsZero() {
			return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
				"message": fmt.Errorf("field %s is required but not provided", field.Name),
			})
		}
	}

	// list := 5
	// if update.ID != "" {
	// 	var err error
	// 	list, err = strconv.Atoi(update.ID)
	// 	if err != nil {
	// 		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 			"message": "Server error: " + err.Error(),
	// 		})
	// 	}
	// }

	// Set Appointment directly into posgres database, but it is time consuming...
	// appointments, err := s.store.SetAppointments_postgres(healthcareID, update.HealthID, update.Status, int64(list))
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"error": err.Error(),
	// 	})
	// }

	//push into queue for processing
	notify_appointment := map[string]interface{}{
		"update": update,
	}
	err = s.store.Push_update_appointment(notify_appointment)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Server error: " + err.Error(),
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "Updation Queued",
		"message":      "appointment will be updated shortly",
		"appointments": update,
	})
}

func (s *APIServer) Create_ClientProfile(w http.ResponseWriter, r *http.Request) error {
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

	// healthcare Name for logs
	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "HealthID not found in token"})
	}

	// create client_profile using function
	client_profile, err := mod.Create_clientProfile(healthcareID, patient)
	if err != nil {
		return err
	}

	// store into posgres directly
	err = s.store.Create_ClientProfile(client_profile)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"err":     err.Error(),
			"message": "User Already exists",
		})
	}

	// create stats for this patient also
	err = s.store.CreateClient_stats(client_profile.HealthID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something Went Wrong from our side :(",
			"err":     err.Error(),
		})
	}

	err = s.store.Push_logs("profile_updated", client_profile.FirstName, client_profile.Email, client_profile.HealthID, healthcare_name, healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something Went Wrong from our side :(",
			"err":     err.Error(),
		})
	}

	return writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "data has been successfully created!",
		"status":    "created",
		"email":     client_profile.Email,
		"health_id": client_profile.HealthID,
		"fullname":  client_profile.FirstName + " " + client_profile.LastName,
	})
}

func (s *APIServer) Get_clientProfile(w http.ResponseWriter, r *http.Request) error {
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
	// healthcare_name
	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "healthcare_name not found in token"})
	}

	patientDetails, err := s.store.Get_ClientProfile(healthID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"message": "No Patient Found :(",
		})
	}

	// Notify user via email
	err = s.store.Push_logs("profile_viewed", patientDetails.FirstName, patientDetails.Email, patientDetails.HealthID, healthcare_name, healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"err":     err.Error(),
			"message": "something went wrong from our side :(",
		})
	}

	return writeJSON(w, http.StatusOK, map[string]interface{}{"client_profile": patientDetails})
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
				fetchedData, ok := fetched.(struct {
					Value string        `json:"value"`
					TTL   time.Duration `json:"ttl"`
				})

				if !ok {
					return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
						"error": "Failed to convert data to expected format",
					})
				}
				var jsonBody *mod.HIPInfo
				err = json.Unmarshal([]byte(fetchedData.Value), &jsonBody)
				if err != nil {
					return err
				}
				return writeJSON(w, http.StatusOK, map[string]interface{}{
					"preferance":         jsonBody,
					"refreshIn(seconds)": fetchedData.TTL.Seconds(),
				})
			}
		}
	}

	// fetch from database now!!
	hipdetails, err := s.store.GetHealthcare_details_postgres(healthcareID)
	if err != nil {
		return writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"message": "Healthcare Not Found!",
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

	if patientrecords.MedicalSeverity != "High" && patientrecords.MedicalSeverity != "Low" && patientrecords.MedicalSeverity != "Severe" && patientrecords.MedicalSeverity != "Normal" {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"message": "medical_severity value not acceptable must be one of [High, Low, Severe, Normal]",
		})
	}

	healthcareId, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "StatusUnauthorized"})
	}

	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "healthcare_name not found in token"})
	}

	// assign healthcareId
	patientrecords.Createdby_ = healthcareId
	patientrecords.HealthcareName = healthcare_name

	// pushing into database
	// Leave this for now
	// patientrecords_created, err := s.store.CreatepatientRecords(healthcareId, patientrecords)
	// if err != nil {
	// 	return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
	// 		"message": err,
	// 	})
	// }

	patientrecords, err = mod.CreatePatientRecords(healthcareId, patientrecords)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"err":     err.Error(),
			"message": "Wrong Payload provided by User!",
		})
	}

	// Convert into body format
	body := map[string]interface{}{
		"record": patientrecords,
	}
	// Push it intoRabbitMq
	err = s.store.Push_patient_records(body)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Something bad (Please Mail 21vaibhav11@gmail.com for this issue)",
			"status":  "Server Could not Process your Request",
			"err":     err.Error(),
		})
	}

	// Notify user via email
	err = s.store.Push_logs("records_created", nil, nil, patientrecords.HealthID, healthcare_name, healthcareId)
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

	return writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "successfully processed, within few hours records will be created",
		"status":  "pending",
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

	// fetch according to severity
	// if not present then fetch all the medical records
	severity := query.Get("severity")
	patientRecords, err := s.store.GetPatientRecords(health_id, severity, list)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "patient not found :(",
		})
	}
	healthcareId, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "StatusUnauthorized"})
	}
	// healthcare_name
	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "healthcare_name not found in token"})
	}

	// push logs that your records_has been viewed and send notifications
	err = s.store.Push_logs("records_viewed", nil, nil, health_id, healthcare_name, healthcareId)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error: could not process data",
		})
	}

	// counters (Will be removed soon)
	// err = s.store.Push_counters("hip:recordsviewed_counter", healthcareId)
	// if err != nil {
	// 	return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
	// 		"message": "Something Mishappened (Please Mail 21vaibhav11@gmail.com for this issue)",
	// 		"status":  "Server Could not Process your Request",
	// 		"err":     err.Error(),
	// 	})
	// }
	if severity == "" {
		severity = "N/A"
	}
	return writeJSON(w, http.StatusOK, map[string]interface{}{
		// "message": "successfull",
		// "fetch":           patientRecords.,
		"patient_records": patientRecords,
		"severity":        severity,
	})
}

func (s *APIServer) UpdateClientProfile(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "PATCH" {
		return writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error": r.Method + " method not allowed",
		})
	}
	// healthcare_name
	healthcareId, ok := r.Context().Value(contextKeyHealthCareID).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "StatusUnauthorized"})
	}
	// healthcare_name
	healthcare_name, ok := r.Context().Value(contextKeyHealthCareName).(string)
	if !ok {
		return writeJSON(w, http.StatusUnauthorized, map[string]string{"HealthID": "healthcare_name not found in token"})
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

	// Update client directly in postgres database
	updatedPatient, err := s.store.Update_clientProfile(healthID, updates)
	if err != nil {
		return writeJSON(w, http.StatusNotAcceptable, map[string]interface{}{
			"err":     err.Error(),
			"message": "Wrong Payload Provided by User",
		})
	}

	// push the logs into queue
	err = s.store.Push_logs("profile_updated", updatedPatient.FirstName, updatedPatient.Email, updatedPatient.HealthID, healthcare_name, healthcareId)
	if err != nil {
		return writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"message": "Internal Server Error: could not process data",
		})
	}

	return writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"updated_details": updatedPatient,
	})
}

// ///////////////////////////// ///////////////////// ///////////////// //////////// /////////////// ////////////// /
/////////////////////////// ///  	 Utility Functions  	///////////////////////// ////////////////// ///////////// ///////

// Rate limiter goes here...
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
		"expiresAt":        time.Now().Add(5 * 24 * time.Hour).Unix(), //setting it to 5days from now
		"healthcareID":     account.HealthcareID,
		"healthcare_email": account.Email,
		"healthcare_name":  account.HealthcareName,
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
			emailHealthcareID, _ := claims["healthcare_email"].(string)
			nameHealthcare, _ := claims["healthcare_name"].(string)

			// Block the request if healthcareID is missing or invalid
			if healthcareID == "" {
				writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid Token: healthcareID missing"})
				return
			}

			// Block the request if emailHealthcareID is missing or invalid
			if emailHealthcareID == "" {
				writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid Token: healthcare_email missing"})
				return
			}
			if nameHealthcare == "" {
				writeJSON(w, http.StatusForbidden, apiError{Error: "Invalid Token: healthcare name missing"})
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyHealthCareID, healthcareID)
			ctx = context.WithValue(ctx, contextKeyEmailHealthCareID, emailHealthcareID)
			ctx = context.WithValue(ctx, contextKeyHealthCareName, nameHealthcare)

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

// isvalid email
// isValidEmail validates the email format
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}
