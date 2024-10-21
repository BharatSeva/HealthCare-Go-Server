package main

import (
	// "github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Address struct {
	Country  string `json:"country" validate:"required,max=30"`  // Required, max 30 characters
	State    string `json:"state" validate:"required,max=20"`    // Required, max 20 characters
	City     string `json:"city" validate:"required,max=30"`     // Required, max 30 characters
	Landmark string `json:"landmark" validate:"required,max=45"` // Required, max 45 characters
}

type HIPInfo struct {
	HealthcareID       int       `json:"healthcare_id" validate:"required"`                   // Required, unique
	HealthcareLicense  string    `json:"healthcare_license" validate:"required,min=4,max=20"` // Required, min 4, max 20, unique
	HealthcareName     string    `json:"name" validate:"required,min=5,max=20"`               // Required, min 5, max 20, unique
	Email              string    `json:"email" validate:"required,email"`                     // Required, valid email, unique
	Availability       string    `json:"availability" validate:"required,min=2,max=15"`       // Required, min 2, max 15
	TotalFacilities    int       `json:"total_facilities" validate:"required,min=4,max=15"`   // Required, min 4, max 15
	TotalMBBSDoc       int       `json:"total_mbbs_doc" validate:"required,min=4,max=15"`     // Required, min 4, max 15
	TotalWorker        int       `json:"total_worker" validate:"required,min=4,max=15"`       // Required, min 4, max 15
	NoOfBeds           int       `json:"no_of_beds" validate:"required,min=4,max=15"`         // Required, min 4, max 15
	DateOfRegistration time.Time `json:"date_of_registration"`                                // Default to current time
	Password           string    `json:"password" validate:"required,min=3"`                  // Required, min 3
	Address            Address   `json:"address" validate:"required"`
}

type HealthCarePortal struct {
	ID                      int     `json:"id"`
	ScheduledDeletion       bool    `json:"scheduled_deletion"`
	BiodataViewed_count     int     `json:"biodata_viewed_count"`
	HealthidCreated_count   int     `json:"healthID_created_count"`
	AccountLocked           bool    `json:"account_locked"`
	RecordsCreated_count    int     `json:"records_created_count"`
	RecordsViewed_count     int     `json:"recordsViewed_count"`
	TotalnoOfviews_count    int     `json:"totalnoOfviews_count"`
	TotalAppointments_count int     `json:"totalAppointments_count"`
	TotalRequest_count      int     `json:"totalRequest_count"`
	About                   string  `json:"about"`
	AppointmentFee          int     `json:"appointmentFee"`
	Isavailable             bool    `json:"isavailable"`
	Email                   string  `json:"email" validate:"required,email"`
	Name                    string  `json:"name"`
	Rating                  string  `json:"rating"`
	Address                 Address `json:"address" validate:"required"`
}

type Login struct {
	HealthcareID      int    `json:"healthcare_id" validate:"required"`
	HealthcareLicense string `json:"healthcare_license" validate:"required,min=4,max=20"`
	Password          string `json:"password" validate:"required,min=3"`
}

func SignUpAccount(hip *HIPInfo) (*HIPInfo, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(hip.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &HIPInfo{
		HealthcareID:       hip.HealthcareID,
		HealthcareLicense:  hip.HealthcareLicense,
		HealthcareName:     hip.HealthcareName,
		Email:              hip.Email,
		Availability:       hip.Availability,
		TotalFacilities:    hip.TotalFacilities,
		TotalMBBSDoc:       hip.TotalMBBSDoc,
		TotalWorker:        hip.TotalWorker,
		NoOfBeds:           hip.NoOfBeds,
		DateOfRegistration: hip.DateOfRegistration,
		Password:           string(encpw),
		Address: Address{
			Country:  hip.Address.Country,
			State:    hip.Address.State,
			City:     hip.Address.City,
			Landmark: hip.Address.Landmark,
		},
	}, nil
}

// this is for appointments sections
type Appointments struct {
	HealthcareID    int    `json:"healthcare_id" validate:"required"`
	AppointmentDate string `json:"appointment_date"`
	AppointmentTime string `json:"appointment_time"`
	HealthID        string `json:"health_id" validate:"required,min=10,max=10"`
	Department      string `json:"department"`
	Note            string `json:"note" validate:"max=500"`
	FirstName       string `json:"fname" validate:"required,min=3,max=50"`
	MiddleName      string `json:"middlename" validate:"max=50"`
	LastName        string `json:"lname" validate:"required,min=1,max=50"`
	HealthcareName  string `json:"name" validate:"required,min=5,max=50"`
}

type PatientDetails struct {
	HealthID        string    `json:"health_id" validate:"required,min=10,max=10"`       // Required, min 10, max 10
	FirstName       string    `json:"fname" validate:"required,min=3,max=10"`            // Required, min 3, max 10
	MiddleName      string    `json:"middlename" validate:"max=10"`                      // Optional, max 10
	LastName        string    `json:"lname" validate:"required,min=3,max=10"`            // Required, min 3, max 10
	Sex             string    `json:"sex" validate:"required,min=1,max=5"`               // Required, min 1, max 5
	HealthcareID    int       `json:"healthcare_id" validate:"required"`                 // Required
	HealthcareName  string    `json:"healthcare_name" validate:"required,min=1,max=30"`  // Required, min 1, max 30
	DOB             string    `json:"dob" validate:"required,min=1,max=15"`              // Required, min 1, max 15
	BloodGroup      string    `json:"bloodgrp" validate:"required,min=1,max=20"`         // Required, min 1, max 20
	BMI             string    `json:"bmi" validate:"required,min=1,max=10"`              // Required, min 1, max 10
	MarriageStatus  string    `json:"marriage_status" validate:"required,min=1,max=20"`  // Required, min 1, max 20
	Weight          string    `json:"weight" validate:"required,min=1,max=10"`           // Required, min 1, max 10
	Email           string    `json:"email" validate:"required,email,max=50"`            // Required, valid email, max 50
	MobileNumber    string    `json:"mobilenumber" validate:"required,min=1,max=10"`     // Required, min 1, max 10
	AadhaarNumber   string    `json:"aadhar_number" validate:"required,min=1,max=20"`    // Required, min 1, max 20
	PrimaryLocation string    `json:"primary_location" validate:"required,min=1,max=50"` // Required, min 1, max 50
	Sibling         string    `json:"sibling" validate:"required,min=1,max=10"`          // Required, min 1, max 10
	Twin            string    `json:"twin" validate:"required,min=1,max=10"`             // Required, min 1, max 10
	FatherName      string    `json:"fathername" validate:"required,min=1,max=10"`       // Required, min 1, max 10
	MotherName      string    `json:"mothername" validate:"required,min=1,max=10"`       // Required, min 1, max 10
	EmergencyNumber string    `json:"emergencynumber" validate:"required,min=1,max=10"`  // Required, min 1, max 10
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type PatientProblem struct {
	PProblem        string    `json:"p_problem" validate:"required,min=3,max=20"`   // Required, min 3, max 20 characters
	Description     string    `json:"description" validate:"required,min=3,max=50"` // Required, min 3, max 50 characters
	HealthID        int       `json:"health_id" validate:"required"`
	HealthcareName  string    `json:"healthcare_name" validate:"required"`
	MedicalSeverity string    `json:"medical_severity" validate:"required"`
	CreatedAt       time.Time `json:"created_at"`
}

// Utility structs
type ChangePreferance struct {
	Email              string `json:"email"`
	IsAvailable        bool   `json:"isAvailable"`
	Scheduled_deletion bool   `json:"scheduled_deletion"`
}

