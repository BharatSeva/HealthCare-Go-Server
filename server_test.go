package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Address struct {
	Country  string `json:"country"`
	Landmark string `json:"landmark"`
	City     string `json:"city"`
	State    string `json:"state"`
}

type RegisterRequest struct {
	Name            string  `json:"name"`
	Availability    string  `json:"availability"`
	TotalFacilities int     `json:"total_facilities"`
	TotalMbbsDoc    int     `json:"total_mbbs_doc"`
	TotalWorker     int     `json:"total_worker"`
	NoOfBeds        int     `json:"no_of_beds"`
	Email           string  `json:"email"`
	AppointmentFee  int     `json:"appointment_fee"`
	About           string  `json:"about"`
	Password        string  `json:"password"`
	Address         Address `json:"address"`
}

type HealthcareDetails struct {
	HealthcareID      string `json:"healthcare_id"`
	HealthcareLicense string `json:"healthcare_license"`
}

type RegisterResponse struct {
	Status            string            `json:"status"`
	HealthcareDetails HealthcareDetails `json:"healthcare_details"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := RegisterResponse{
		Status: "Successfully Created",
		HealthcareDetails: HealthcareDetails{
			HealthcareID:      "HCID12345",
			HealthcareLicense: "HCID123",
		},
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func TestRegisterEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		request          RegisterRequest
		expectedStatus   int
		validateResponse func(*testing.T, []byte)
	}{
		{
			name: "Successful Registration",
			request: RegisterRequest{
				Name:            "Test Hospital",
				Availability:    "Yes",
				TotalFacilities: 200,
				TotalMbbsDoc:    58,
				TotalWorker:     400,
				NoOfBeds:        200,
				Email:           "test@hospital.com",
				AppointmentFee:  300,
				About:           "Test Hospital Description",
				Password:        "11",
				Address: Address{
					Country:  "India",
					Landmark: "Test Landmark",
					City:     "Test City",
					State:    "Test State",
				},
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, body []byte) {
				var response RegisterResponse
				err := json.Unmarshal(body, &response)
				fmt.Println("RESPONSE***********", response)
				assert.NoError(t, err)
				assert.Equal(t, "Successfully Created", response.Status)
				assert.NotEmpty(t, response.HealthcareDetails.HealthcareID)
				assert.NotEmpty(t, response.HealthcareDetails.HealthcareLicense)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "http://localhost:3000/api/v1/healthcareauth/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			RegisterHandler(w, req) // Call the handler function

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w.Body.Bytes())
			}
		})
	}
}

type LoginRequest struct {
	HealthcareID      string `json:"healthcare_id"`
	HealthcareLicense string `json:"healthcare_license"`
	Password          string `json:"password"`
}

type LoginResponse struct {
	Token          string `json:"token"`
	ExpiresIn      string `json:"Expires In"`
	HealthcareId   string `json:"healthcare_id"`
	HealthcareName string `json:"healthcare_name"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.HealthcareID == "HCID123" && req.HealthcareLicense == "HCID123" && req.Password == "11" {
		response := LoginResponse{
			Token:     "tokengoeshere",
			ExpiresIn: "5d",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

func TestLoginEndpoint(t *testing.T) {
	tests := []struct {
		name             string
		request          LoginRequest
		expectedStatus   int
		validateResponse func(*testing.T, []byte)
	}{
		{
			name: "Successful Login",
			request: LoginRequest{
				HealthcareID:      "HCID123",
				HealthcareLicense: "LICENSE123",
				Password:          "11",
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, body []byte) {
				var response LoginResponse
				err := json.Unmarshal(body, &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "5d", response.ExpiresIn)
			},
		},
		{
			name: "Failed Login - Invalid Credentials",
			request: LoginRequest{
				HealthcareID:      "INVALID_ID",
				HealthcareLicense: "INVALID_LICENSE",
				Password:          "wrong_password",
			},
			expectedStatus: http.StatusUnauthorized,
			validateResponse: func(t *testing.T, body []byte) {
				assert.Empty(t, body) // Expecting no body on unauthorized
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "http://localhost:3000/api/v1/healthcareauth/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			LoginHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w.Body.Bytes())
			}
		})
	}
}
