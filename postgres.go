package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(url string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateTable()
}

func (s *PostgresStore) CreateTable() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS HIP_TABLE (
			healthcare_id SERIAL PRIMARY KEY,
			healthcare_license VARCHAR(20) NOT NULL UNIQUE,
			healthcare_name VARCHAR(20) NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			availability VARCHAR(15) NOT NULL,
			total_facilities INTEGER NOT NULL,
			total_mbbs_doc INTEGER NOT NULL,
			total_worker INTEGER NOT NULL,
			no_of_beds INTEGER NOT NULL,
			date_of_registration TIMESTAMP DEFAULT NOW(),
			password TEXT NOT NULL,
			country VARCHAR(30) NOT NULL,
			state VARCHAR(20) NOT NULL,
			city VARCHAR(30) NOT NULL,
			landmark VARCHAR(45) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS HealthCare_Logs (
			Id SERIAL PRIMARY KEY,
			healthcare_id INTEGER NOT NULL,
			scheduled_deletion VARCHAR(20),
			biodata_viewed_count INTEGER,
			healthID_created_count INTEGER NOT NULL,
			account_locked VARCHAR(15) NOT NULL,
			records_created_count INTEGER NOT NULL,
			recordsViewed_count INTEGER NOT NULL,
			totalnoOfviews_count INTEGER NOT NULL,
			totalAppointments_count INTEGER NOT NULL,
			totalRequest_count INTEGER NOT NULL,
			about TEXT NOT NULL,
			appointmentFee INTEGER NOT NULL,
			isAvailable VARCHAR(20) NOT NULL,
		--	// name VARCHAR(30) NOT NULL,
		--	// country VARCHAR(30) NOT NULL,
		--	// state VARCHAR(20) NOT NULL,
		--	// city VARCHAR(30) NOT NULL,
		--	// landmark VARCHAR(45) NOT NULL,
			FOREIGN KEY (healthcare_id) REFERENCES HIP_TABLE(healthcare_id) ON DELETE CASCADE  -- Foreign Key linking HIP_TABLE
		)`,
	}
	for _, query := range queries {
		_, err := s.db.Exec(query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) SignUpAccount(hip *HIPInfo) (int64, error) {
	query := `INSERT INTO HIP_TABLE (healthcare_license, 
		healthcare_name, email, availability, total_facilities, 
		total_mbbs_doc, total_worker, no_of_beds, password, country, 
		state, city, landmark)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING healthcare_id`

	query1 := `INSERT INTO HealthCare_Logs (healthcare_id, scheduled_deletion, biodata_viewed_count, 
			  healthID_created_count, account_locked, records_created_count, recordsViewed_count, 
			  totalnoOfviews_count, totalAppointments_count, totalRequest_count, 
			  about, appointmentFee, isAvailable)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	// Check if email already exists
	exists, err := checkEmailExists(s.db, hip.Email)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("email %s already exists", hip.Email)
	}

	// Insert into HIP_TABLE and get the generated healthcare_id
	var healthcareID int64
	err = s.db.QueryRow(query, hip.HealthcareLicense, hip.HealthcareName, hip.Email, hip.Availability, hip.TotalFacilities, hip.TotalMBBSDoc, hip.TotalWorker, hip.NoOfBeds, hip.Password, hip.Address.Country, hip.Address.State, hip.Address.City, hip.Address.Landmark).Scan(&healthcareID)
	if err != nil {
		return 0, err
	}

	// Insert into HealthCare_Logs using the healthcare_id
	_, err = s.db.Exec(query1, healthcareID, "false", 0, 0, "false", 0, 0, 0, 0, 0, "TestingPhase - 1", 100, "true")
	if err != nil {
		return 0, err
	}
	return healthcareID, nil
}

func (s *PostgresStore) LoginUser(acc *Login) (*HIPInfo, error) {
	var hip HIPInfo
	query := `SELECT healthcare_id, healthcare_license, healthcare_name, email, availability, total_facilities, total_mbbs_doc, total_worker, no_of_beds, date_of_registration, password, country, state, city, landmark
	          FROM HIP_TABLE WHERE healthcare_license = $1`

	err := s.db.QueryRow(query, acc.HealthcareLicense).Scan(&hip.HealthcareID, &hip.HealthcareLicense, &hip.HealthcareName, &hip.Email, &hip.Availability, &hip.TotalFacilities, &hip.TotalMBBSDoc, &hip.TotalWorker, &hip.NoOfBeds, &hip.DateOfRegistration, &hip.Password, &hip.Address.Country, &hip.Address.State, &hip.Address.City, &hip.Address.Landmark)
	if err != nil {
		return nil, fmt.Errorf("%s no account found with given HealthcareLicense: %s", err, acc.HealthcareLicense)
	}
	return &hip, nil
}

func (s *PostgresStore) ChangePreferance(healthcareId int, preferance *ChangePreferance) error {
	if len(preferance.Email) > 1 {
		_, err := s.db.Exec("UPDATE HIP_TABLE set email = $1 WHERE healthcare_id = $2", preferance.Email , healthcareId)
		if err != nil {
			return err
		}
	}

	if !preferance.IsAvailable {
		_, err := s.db.Exec("UPDATE HealthCare_Logs set isAvailable = $1 WHERE healthcare_id = $2", preferance.IsAvailable, healthcareId)
		if err != nil {
			return err
		}
	}

	if preferance.Scheduled_deletion {
		id, err := s.db.Exec("UPDATE HealthCare_Logs set scheduled_deletion = $1 WHERE healthcare_id = $2", preferance.Scheduled_deletion, healthcareId)
		if err != nil {
			return err
		}
		ID, _ := id.LastInsertId();
		fmt.Println(ID)
	}
	return nil
}

func (s *PostgresStore) GetPreferance(healthcareId int) (*ChangePreferance, error) {
	query := `
		SELECT 
			HIP_TABLE.email, 
			HealthCare_Logs.isavailable, 
			HealthCare_Logs.scheduled_deletion 
		FROM 
			HIP_TABLE 
		INNER JOIN 
			HealthCare_Logs 
		ON 
			HIP_TABLE.healthcare_id = HealthCare_Logs.healthcare_id 
		WHERE 
			HIP_TABLE.healthcare_id = $1;
	`
	preferance := &ChangePreferance{}
	fmt.Println(healthcareId)
	err := s.db.QueryRow(query, healthcareId).Scan(&preferance.Email, &preferance.IsAvailable, &preferance.Scheduled_deletion)
	fmt.Println(preferance)
	if err != nil {
		return nil, err
	}
	return preferance, nil
}

// Utility Functions
func checkEmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM HIP_TABLE WHERE email = $1)"
	err := db.QueryRow(query, email).Scan(&exists)
	return exists, err
}
