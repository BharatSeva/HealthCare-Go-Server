package databases

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func ConnectToPostgreSQL(url string) (*PostgresStore, error) {
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
			Id SERIAL PRIMARY KEY,
			healthcare_id TEXT NOT NULL UNIQUE,
			healthcare_license TEXT NOT NULL UNIQUE,
			healthcare_name TEXT NOT NULL UNIQUE,
			email VARCHAR(100) NOT NULL UNIQUE,
			availability VARCHAR(15) NOT NULL,
			total_facilities INTEGER NOT NULL, 
			total_mbbs_doc INTEGER NOT NULL,
			total_worker INTEGER NOT NULL, 
			no_of_beds INTEGER NOT NULL,
			date_of_registration TIMESTAMP DEFAULT NOW(),
			password TEXT NOT NULL,
			about VARCHAR(300) NOT NULL,
			country VARCHAR(30) NOT NULL,
			state VARCHAR(20) NOT NULL,
			city VARCHAR(30) NOT NULL,
			landmark VARCHAR(45) NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS HealthCare_pref (
			Id SERIAL PRIMARY KEY,
			healthcare_id TEXT NOT NULL,
			scheduled_deletion VARCHAR(20),
			biodata_viewed_count INTEGER,
			healthID_created_count INTEGER NOT NULL,
			account_locked VARCHAR(15) NOT NULL,
			records_created_count INTEGER NOT NULL,
			recordsviewed_count INTEGER NOT NULL,
			totalrequest_count INTEGER NOT NULL,
			appointmentFee INTEGER NOT NULL,
			isAvailable VARCHAR(20) NOT NULL,
			FOREIGN KEY (healthcare_id) REFERENCES HIP_TABLE(healthcare_id) ON DELETE CASCADE
		);`,

		`CREATE TABLE IF NOT EXISTS Appointments (
			Id SERIAL PRIMARY KEY,
			healthcare_id TEXT NOT NULL,
			appointment_date DATE NOT NULL,
			appointment_time TIME NOT NULL,
			health_id VARCHAR(10) NOT NULL,
			department VARCHAR(50),
			note VARCHAR(500),
			fname VARCHAR(50) NOT NULL,
			middlename VARCHAR(50),
			lname VARCHAR(50) NOT NULL,
			name VARCHAR(50) NOT NULL,
			FOREIGN KEY (healthcare_id) REFERENCES HIP_TABLE(healthcare_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS client_stats (
			health_id VARCHAR PRIMARY KEY UNIQUE,
			account_status VARCHAR CHECK (account_status IN ('Trial', 'Testing', 'Beta', 'Premium')) NOT NULL DEFAULT 'Trial',
			available_money VARCHAR NOT NULL DEFAULT '5000',
			profile_viewed INTEGER NOT NULL DEFAULT 0,
			profile_updated INTEGER NOT NULL DEFAULT 0,
			records_viewed INTEGER NOT NULL DEFAULT 0,
			records_created INTEGER NOT NULL DEFAULT 0
		);`,
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
	query := `INSERT INTO HIP_TABLE (healthcare_id, healthcare_license, 
		healthcare_name, email, availability, total_facilities, 
		total_mbbs_doc, total_worker, no_of_beds, password, about, country, 
		state, city, landmark)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) RETURNING healthcare_id`

	query1 := `INSERT INTO HealthCare_pref (healthcare_id, scheduled_deletion, biodata_viewed_count, 
			  healthID_created_count, account_locked, records_created_count, recordsviewed_count, 
			  totalRequest_count, appointmentFee, isAvailable)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	// Check if email already exists
	exists, err := checkEmailExists(s.db, hip.Email)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, fmt.Errorf("email %s already exists", hip.Email)
	}

	// Insert into HIP_TABLE and get the generated healthcare_id
	var healthcareID string
	err = s.db.QueryRow(query, hip.HealthcareID, hip.HealthcareLicense, hip.HealthcareName, hip.Email, hip.Availability, hip.TotalFacilities, hip.TotalMBBSDoc, hip.TotalWorker, hip.NoOfBeds, hip.Password, hip.About, hip.Address.Country, hip.Address.State, hip.Address.City, hip.Address.Landmark).Scan(&healthcareID)
	if err != nil {
		return 0, err
	}

	// Insert into HealthCare_Logs using the healthcare_id
	Id, err := s.db.Exec(query1, healthcareID, "false", 0, 0, "false", 0, 0, 100, 100, "true")
	if err != nil {
		return 0, err
	}
	Inserted_id, _ := Id.LastInsertId()
	return Inserted_id, nil
}

func (s *PostgresStore) LoginUser(acc *Login) (*HIPInfo, error) {
	var hip HIPInfo
	query := `SELECT healthcare_id, healthcare_license, healthcare_name, email, availability, total_facilities, total_mbbs_doc, total_worker, no_of_beds, date_of_registration, password, country, state, city, landmark
	          FROM HIP_TABLE WHERE healthcare_id = $1`

	err := s.db.QueryRow(query, acc.HealthcareID).Scan(&hip.HealthcareID, &hip.HealthcareLicense, &hip.HealthcareName, &hip.Email, &hip.Availability, &hip.TotalFacilities, &hip.TotalMBBSDoc, &hip.TotalWorker, &hip.NoOfBeds, &hip.DateOfRegistration, &hip.Password, &hip.Address.Country, &hip.Address.State, &hip.Address.City, &hip.Address.Landmark)
	if err != nil {
		return nil, fmt.Errorf("error : %w", err)
	}
	return &hip, nil
}

func (s *PostgresStore) ChangePreferance(healthcareId string, preferance map[string]interface{}) error {
	for key, value := range preferance {
		if key == "email" && value != "" {
			_, err := s.db.Exec("UPDATE HIP_TABLE set email = $1 WHERE healthcare_id = $2", value, healthcareId)
			if err != nil {
				return err
			}
		}
	}
	for key, value := range preferance {
		if key == "scheduled_deletion" && value != "" {
			_, err := s.db.Exec("UPDATE HealthCare_pref set scheduled_deletion = $1 WHERE healthcare_id = $2", value, healthcareId)
			if err != nil {
				return err
			}
		}
	}
	for key, value := range preferance {
		if key == "isAvailable" && value != "" {
			_, err := s.db.Exec("UPDATE HealthCare_pref set isAvailable = $1 WHERE healthcare_id = $2", value, healthcareId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PostgresStore) GetPreferance(healthcareId string) (*ChangePreferance, error) {
	query := `
		SELECT 
			HIP_TABLE.email, 
			HealthCare_pref.isavailable, 
			HealthCare_pref.scheduled_deletion 
		FROM 
			HIP_TABLE 
		INNER JOIN 
		HealthCare_pref 
		ON 
			HIP_TABLE.healthcare_id = HealthCare_pref.healthcare_id 
		WHERE 
			HIP_TABLE.healthcare_id = $1;
	`
	preferance := &ChangePreferance{}
	err := s.db.QueryRow(query, healthcareId).Scan(&preferance.Email, &preferance.IsAvailable, &preferance.Scheduled_deletion)
	if err != nil {
		return nil, err
	}
	return preferance, nil
}

// Get totalRequest from database
func (s *PostgresStore) GetTotalRequestCount(healthcare_id string) (int, error) {
	var count int
	query := `
		SELECT totalrequest_count 
		FROM HealthCare_pref 
		WHERE healthcare_id = $1;
	`
	// Execute the query and scan the result into the 'count' variable
	err := s.db.QueryRow(query, healthcare_id).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve totalrequest_count: %w", err)
	}

	return count, nil
}

func (s *PostgresStore) CreateClient_stats(health_id string) error {
	query := `INSERT INTO client_stats (health_id, account_status, 
		available_money, profile_viewed, profile_updated, records_viewed, 
		records_created) VALUES ($1, $2, $3, $4, $5, $6, $7);`
	_, err := s.db.Exec(query, health_id, "Trial", 5000, 0, 0, 0, 0)
	if err != nil {
		return err
	}
	return nil
}

// Utility Functions
func checkEmailExists(db *sql.DB, email string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM HIP_TABLE WHERE email = $1)"
	err := db.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// func (s *PostgresStore) Recordsviewed_counter(healthcare_id string) error {
// 	query := `
// 	UPDATE HealthCare_Logs
// 	SET recordsviewed_count = recordsviewed_count + 1
// 	WHERE healthcare_id = $1;
// `
// 	_, err := s.db.Exec(query, healthcare_id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update recordsviewed_count: %w", err)
// 	}
// 	return nil
// }
// func (s *PostgresStore) Recordscreated_counter(healthcare_id string) error {
// 	query := `
// 	UPDATE HealthCare_Logs
// 	SET records_created_count = records_created_count + 1
// 	WHERE healthcare_id = $1;
// `
// 	_, err := s.db.Exec(query, healthcare_id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update records_created_count: %w", err)
// 	}
// 	return nil
// }
// func (s *PostgresStore) Patientbiodata_created_counter(healthcare_id string) error {
// 	query := `
// 	UPDATE HealthCare_Logs
// 	SET healthID_created_count = healthID_created_count + 1
// 	WHERE healthcare_id = $1;
// `
// 	_, err := s.db.Exec(query, healthcare_id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update healthID_created_count: %w", err)
// 	}
// 	return nil
// }
// func (s *PostgresStore) Patientbiodata_viewed_counter(healthcare_id string) error {
// 	query := `
// 	UPDATE HealthCare_Logs
// 	SET biodata_viewed_count = biodata_viewed_count + 1
// 	WHERE healthcare_id = $1;
// `
// 	_, err := s.db.Exec(query, healthcare_id)
// 	if err != nil {
// 		return fmt.Errorf("failed to update biodata_viewed_count: %w", err)
// 	}
// 	return nil
// }
