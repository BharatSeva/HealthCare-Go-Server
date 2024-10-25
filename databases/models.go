package databases

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Address struct {
	Country  string `json:"country" validate:"required,max=30"`  // Required, max 30 characters
	State    string `json:"state" validate:"required,max=20"`    // Required, max 20 characters
	City     string `json:"city" validate:"required,max=30"`     // Required, max 30 characters
	Landmark string `json:"landmark" validate:"required,max=45"` // Required, max 45 characters
}

type HIPInfo struct {
	HealthcareID      string `bson:"healthcare_id,omitempty" json:"healthcare_id"`                                  // MongoDB auto-generated ID
	HealthcareLicense string `bson:"healthcare_license" json:"healthcare_license" validate:"required,min=4,max=25"` // Required, min 4, max 20, unique
	HealthcareName    string `bson:"name" json:"name" validate:"required,min=5,max=20"`                             // Required, min 5, max 20, unique
	Email             string `bson:"email" json:"email" validate:"required,email"`                                  // Required, valid email, unique
	Availability      string `bson:"availability" json:"availability" validate:"required,min=2,max=15"`             // Required, min 2, max 15
	TotalFacilities   int    `bson:"total_facilities" json:"total_facilities" validate:"required,min=4,max=15"`     // Required, min 4, max 15
	TotalMBBSDoc      int    `bson:"total_mbbs_doc" json:"total_mbbs_doc" validate:"required,min=4,max=15"`         // Required, min 4, max 15
	TotalWorker       int    `bson:"total_worker" json:"total_worker" validate:"required,min=4,max=15"`             // Required, min 4, max 15
	NoOfBeds          int    `bson:"no_of_beds" json:"no_of_beds" validate:"required,min=4,max=15"`                 // Required, min 4, max 15
	// postgres accept time.Time and
	// mongo db accept primitive.Datetime
	About              string    `bson:"about" json:"about" validate:"required,min=5,max=200"`
	DateOfRegistration time.Time `bson:"date_of_registration" json:"date_of_registration"`   // Default to current time
	Password           string    `bson:"password" json:"password" validate:"required,min=3"` // Required, min 3
	Address            Address   `bson:"address" json:"address" validate:"required"`
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
	HealthcareID      string `json:"healthcare_id" validate:"required"`
	HealthcareLicense string `json:"healthcare_license" validate:"required,min=4,max=20"`
	Password          string `json:"password" validate:"required,min=3"`
}

func SignUpAccount(hip *HIPInfo) (*HIPInfo, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(hip.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	uniquehealthID := uuid.New().String()[:20]
	return &HIPInfo{
		HealthcareID:       "HCID" + uniquehealthID,
		HealthcareLicense:  uniquehealthID,
		HealthcareName:     hip.HealthcareName,
		Email:              hip.Email,
		Availability:       hip.Availability,
		TotalFacilities:    hip.TotalFacilities,
		TotalMBBSDoc:       hip.TotalMBBSDoc,
		TotalWorker:        hip.TotalWorker,
		NoOfBeds:           hip.NoOfBeds,
		DateOfRegistration: hip.DateOfRegistration,
		About:              hip.About,
		Password:           string(encpw),
		Address: Address{
			Country:  hip.Address.Country,
			State:    hip.Address.State,
			City:     hip.Address.City,
			Landmark: hip.Address.Landmark,
		},
	}, nil
}

type Appointments struct {
	HealthcareID    string `json:"healthcare_id" bson:"healthcare_id" validate:"required"`
	AppointmentDate string `json:"appointment_date" bson:"appointment_date"`
	AppointmentTime string `json:"appointment_time" bson:"appointment_time"`
	HealthID        string `json:"health_id" bson:"health_id" validate:"required,min=10,max=10"`
	Department      string `json:"department" bson:"department"`
	Note            string `json:"note" bson:"note" validate:"max=500"`
	FirstName       string `json:"fname" bson:"fname" validate:"required,min=3,max=50"`
	MiddleName      string `json:"middlename" bson:"middlename" validate:"max=50"`
	LastName        string `json:"lname" bson:"lname" validate:"required,min=1,max=50"`
	HealthcareName  string `json:"name" bson:"name" validate:"required,min=5,max=50"`
}

type PatientDetails struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	HealthID        string             `bson:"health_id" json:"health_id" validate:"required,min=10,max=10"`
	FirstName       string             `bson:"fname" json:"fname" validate:"required,min=3,max=10"`
	MiddleName      string             `bson:"middlename" json:"middlename" validate:"max=10"`
	LastName        string             `bson:"lname" json:"lname" validate:"required,min=3,max=10"`
	Sex             string             `bson:"sex" json:"sex" validate:"required,min=1,max=5"`
	HealthcareID    string             `bson:"healthcare_id" json:"healthcare_id" validate:"required"`
	DOB             string             `bson:"dob" json:"dob" validate:"required,min=1,max=15"`
	BloodGroup      string             `bson:"bloodgrp" json:"bloodgrp" validate:"required,min=1,max=20"`
	BMI             string             `bson:"bmi" json:"bmi" validate:"required,min=1,max=10"`
	MarriageStatus  string             `bson:"marriage_status" json:"marriage_status" validate:"required,min=1,max=20"`
	Weight          string             `bson:"weight" json:"weight" validate:"required,min=1,max=10"`
	Email           string             `bson:"email" json:"email" validate:"required,email,max=50"`
	MobileNumber    string             `bson:"mobilenumber" json:"mobilenumber" validate:"required,min=1,max=10"`
	AadhaarNumber   string             `bson:"aadhaar_number" json:"aadhar_number" validate:"required,min=1,max=20"`
	PrimaryLocation string             `bson:"primary_location" json:"primary_location" validate:"required,min=1,max=50"`
	Sibling         string             `bson:"sibling" json:"sibling" validate:"required,min=1,max=10"`
	Twin            string             `bson:"twin" json:"twin" validate:"required,min=1,max=10"`
	FatherName      string             `bson:"fathername" json:"fathername" validate:"required,min=1,max=10"`
	MotherName      string             `bson:"mothername" json:"mothername" validate:"required,min=1,max=10"`
	EmergencyNumber string             `bson:"emergencynumber" json:"emergencynumber" validate:"required,min=1,max=10"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	Address         Address            `bson:"address" json:"address"`
}

func CreatePatient_bioData(HealthcareID string, patient *PatientDetails) (*PatientDetails, error) {
	uniquehealthID := uuid.New().String()[:20]
	newPatient := &PatientDetails{
		HealthID:        "HID" + uniquehealthID,
		FirstName:       patient.FirstName,
		MiddleName:      patient.MiddleName,
		LastName:        patient.LastName,
		Sex:             patient.Sex,
		HealthcareID:    HealthcareID,
		DOB:             patient.DOB,
		BloodGroup:      patient.BloodGroup,
		BMI:             patient.BMI,
		MarriageStatus:  patient.MarriageStatus,
		Weight:          patient.Weight,
		Email:           patient.Email,
		MobileNumber:    patient.MobileNumber,
		AadhaarNumber:   patient.AadhaarNumber,
		PrimaryLocation: patient.PrimaryLocation,
		Sibling:         patient.Sibling,
		Twin:            patient.Twin,
		FatherName:      patient.FatherName,
		MotherName:      patient.MotherName,
		EmergencyNumber: patient.EmergencyNumber,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	newPatient.Address = Address{
		Country:  patient.Address.Country,
		State:    patient.Address.State,
		City:     patient.Address.City,
		Landmark: patient.Address.Landmark,
	}
	return newPatient, nil
}

type PatientRecords struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"-"`                              // MongoDB ID field
	Issue           string             `bson:"issue" json:"issue" validate:"required,min=3,max=20"` // Required, min 3, max 20 characters
	Createdby_      string             `bson:"Createdby_" json:"Createdby_" validate:"required"`
	Description     string             `bson:"description" json:"description" validate:"required,min=3,max=50"` // Required, min 3, max 50 characters
	HealthID        string             `bson:"health_id" json:"health_id" validate:"required"`
	MedicalSeverity string             `bson:"medical_severity" json:"medical_severity" validate:"required"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

func CreatePatientRecords(healthcare_id string, patientRecords *PatientRecords) (*PatientRecords, error) {
	return &PatientRecords{
		Issue:           patientRecords.Issue,
		Createdby_:      healthcare_id,
		Description:     patientRecords.Description,
		HealthID:        patientRecords.HealthID,
		MedicalSeverity: patientRecords.MedicalSeverity,
	}, nil
}

// Utility structs
type ChangePreferance struct {
	Email              string `json:"email"`
	IsAvailable        bool   `json:"isAvailable"`
	Scheduled_deletion bool   `json:"scheduled_deletion"`
}
