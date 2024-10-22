package databases

import (
	// "github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	HealthcareID    int    `json:"healthcare_id" bson:"healthcare_id" validate:"required"`
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
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	HealthID        string             `bson:"health_id" json:"health_id" validate:"required,min=10,max=10"`
	FirstName       string             `bson:"first_name" json:"fname" validate:"required,min=3,max=10"`
	MiddleName      string             `bson:"middle_name" json:"middlename" validate:"max=10"`
	LastName        string             `bson:"last_name" json:"lname" validate:"required,min=3,max=10"`
	Sex             string             `bson:"sex" json:"sex" validate:"required,min=1,max=5"`
	HealthcareID    int                `bson:"healthcare_id" json:"healthcare_id" validate:"required"`
	HealthcareName  string             `bson:"healthcare_name" json:"healthcare_name" validate:"required,min=1,max=30"`
	DOB             string             `bson:"dob" json:"dob" validate:"required,min=1,max=15"`
	BloodGroup      string             `bson:"blood_group" json:"bloodgrp" validate:"required,min=1,max=20"`
	BMI             string             `bson:"bmi" json:"bmi" validate:"required,min=1,max=10"`
	MarriageStatus  string             `bson:"marriage_status" json:"marriage_status" validate:"required,min=1,max=20"`
	Weight          string             `bson:"weight" json:"weight" validate:"required,min=1,max=10"`
	Email           string             `bson:"email" json:"email" validate:"required,email,max=50"`
	MobileNumber    string             `bson:"mobile_number" json:"mobilenumber" validate:"required,min=1,max=10"`
	AadhaarNumber   string             `bson:"aadhaar_number" json:"aadhar_number" validate:"required,min=1,max=20"`
	PrimaryLocation string             `bson:"primary_location" json:"primary_location" validate:"required,min=1,max=50"`
	Sibling         string             `bson:"sibling" json:"sibling" validate:"required,min=1,max=10"`
	Twin            string             `bson:"twin" json:"twin" validate:"required,min=1,max=10"`
	FatherName      string             `bson:"father_name" json:"fathername" validate:"required,min=1,max=10"`
	MotherName      string             `bson:"mother_name" json:"mothername" validate:"required,min=1,max=10"`
	EmergencyNumber string             `bson:"emergency_number" json:"emergencynumber" validate:"required,min=1,max=10"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
	Address         Address            `bson:"address" json:"address"`
}

func CreatePatient_bioData(HealthcareID int, patient *PatientDetails) (*PatientDetails, error) {
	newPatient := &PatientDetails{
		HealthID:        patient.HealthID,        
		FirstName:       patient.FirstName,       
		MiddleName:      patient.MiddleName,      
		LastName:        patient.LastName,     
		Sex:             patient.Sex,            
		HealthcareID:    HealthcareID,            
		HealthcareName:  patient.HealthcareName,  
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
