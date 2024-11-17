package databases

import (
	"database/sql"
	"fmt"
	"strings"

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
			profile_viewed INTEGER,
			profile_updated INTEGER NOT NULL,
			account_locked VARCHAR(15) NOT NULL,
			records_created INTEGER NOT NULL,
			records_viewed INTEGER NOT NULL,
			totalrequest_count INTEGER NOT NULL,
			appointmentFee INTEGER NOT NULL,
			isAvailable VARCHAR(20) NOT NULL,
			FOREIGN KEY (healthcare_id) REFERENCES HIP_TABLE(healthcare_id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS client_stats (
			health_id VARCHAR PRIMARY KEY UNIQUE,
			account_status VARCHAR CHECK (account_status IN ('Trial', 'Testing', 'Beta', 'Premium')) NOT NULL DEFAULT 'Trial',
			available_money VARCHAR NOT NULL DEFAULT '5000',
			profile_viewed INTEGER NOT NULL DEFAULT 0,
			profile_updated INTEGER NOT NULL DEFAULT 0,
			records_viewed INTEGER NOT NULL DEFAULT 0,
			records_created INTEGER NOT NULL DEFAULT 0,
			FOREIGN KEY (health_id) REFERENCES client_profile(health_id) ON DELETE CASCADE
		);`,

		// create client_profile
		`CREATE TABLE IF NOT EXISTS client_profile (
			id SERIAL PRIMARY KEY,
			health_id VARCHAR(150) NOT NULL,
			first_name VARCHAR(150) NOT NULL,
			middle_name VARCHAR(150),
			last_name VARCHAR(150) NOT NULL, 
			sex VARCHAR(150) NOT NULL,
			healthcare_id VARCHAR NOT NULL,
			dob VARCHAR(150) NOT NULL, -- Increased length here
			blood_group VARCHAR(150) NOT NULL,
			bmi VARCHAR(150) NOT NULL,
			marriage_status VARCHAR(150) NOT NULL,
			weight VARCHAR(150) NOT NULL, 
			email VARCHAR(150) NOT NULL,
			mobile_number VARCHAR(150) NOT NULL,
			aadhaar_number VARCHAR(150) NOT NULL,
			primary_location VARCHAR(150) NOT NULL,
			sibling VARCHAR(150) NOT NULL,
			twin VARCHAR(150) NOT NULL,
			father_name VARCHAR(150) NOT NULL,
			mother_name VARCHAR(150) NOT NULL,
			emergency_number VARCHAR(150) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			country VARCHAR(150) NOT NULL,
			city VARCHAR(150) NOT NULL,
			state VARCHAR(150) NOT NULL, 
			landmark VARCHAR(150) NOT NULL
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

	query1 := `INSERT INTO HealthCare_pref (healthcare_id, scheduled_deletion, profile_viewed, 
		profile_updated, account_locked, records_created, records_viewed, 
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

func (s *PostgresStore) GetPreferance(healthcareId string) (*Preferance, error) {
	query := `
			SELECT 
				HIP_TABLE.email, 
				HealthCare_pref.isavailable, 
				HealthCare_pref.scheduled_deletion, 
				HealthCare_pref.profile_updated, 
				HealthCare_pref.profile_viewed, 
				HealthCare_pref.records_created, 
				HealthCare_pref.records_viewed 
			FROM 
				HIP_TABLE 
			INNER JOIN 
				HealthCare_pref 
			ON 
				HIP_TABLE.healthcare_id = HealthCare_pref.healthcare_id 
			WHERE 
				HIP_TABLE.healthcare_id = $1;
		`

	preferance := &Preferance{}
	err := s.db.QueryRow(query, healthcareId).Scan(&preferance.Email, &preferance.IsAvailable, &preferance.Scheduled_deletion, &preferance.Profile_updated, &preferance.Profile_viewed, &preferance.Records_created, &preferance.Records_viewed)
	if err != nil {
		return nil, err
	}
	return preferance, nil
}

func (s *PostgresStore) GetHealthcare_details(healthcare_id string) (*HIPInfo, error) {
	query := `SELECT 
		healthcare_id, healthcare_license, healthcare_name, email, availability, 
		total_facilities, total_mbbs_doc, total_worker, no_of_beds, 
		date_of_registration, password, about, country, state, city, landmark
		FROM HIP_TABLE
		WHERE healthcare_id = $1;`

	row := s.db.QueryRow(query, healthcare_id)

	var hip HIPInfo
	err := row.Scan(
		&hip.HealthcareID, &hip.HealthcareLicense, &hip.HealthcareName, &hip.Email, &hip.Availability,
		&hip.TotalFacilities, &hip.TotalMBBSDoc, &hip.TotalWorker, &hip.NoOfBeds,
		&hip.DateOfRegistration, &hip.Password, &hip.About, &hip.Address.Country,
		&hip.Address.State, &hip.Address.City, &hip.Address.Landmark,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no healthcare provider found with ID: %s", healthcare_id)
		}
		return nil, err
	}
	return &hip, nil
}

// create client_profile
func (s *PostgresStore) Create_ClientProfile(client *PatientDetails) error {
	query := `INSERT INTO client_profile (
		health_id, first_name, middle_name, last_name, sex, healthcare_id, 
		dob, blood_group, bmi, marriage_status, weight, email, 
		mobile_number, aadhaar_number, primary_location, sibling, twin, 
		father_name, mother_name, emergency_number, created_at, updated_at, country, city, state, landmark
	) VALUES (
		$1, $2, $3, $4, $5, $6, 
		$7, $8, $9, $10, $11, $12, 
		$13, $14, $15, $16, $17, 
		$18, $19, $20, $21, $22, $23, $24, $25, $26
	);`

	_, err := s.db.Exec(query, client.HealthID, client.FirstName, client.MiddleName, client.LastName, client.Sex,
		client.HealthcareID, client.DOB, client.BloodGroup, client.BMI,
		client.MarriageStatus, client.Weight, client.Email, client.MobileNumber,
		client.AadhaarNumber, client.PrimaryLocation, client.Sibling, client.Twin,
		client.FatherName, client.MotherName, client.EmergencyNumber, client.CreatedAt, client.UpdatedAt,
		client.Address.Country, client.Address.City, client.Address.State, client.Address.Landmark)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) Get_ClientProfile(health_id string) (*PatientDetails, error) {
	query := `SELECT health_id, first_name, middle_name, last_name, sex, healthcare_id, 
	dob, blood_group, bmi, marriage_status, weight, email, 
	mobile_number, aadhaar_number, primary_location, sibling, twin, 
	father_name, mother_name, emergency_number, created_at, updated_at, country, city, state, landmark
	FROM client_profile
	WHERE health_id = $1;`

	row := s.db.QueryRow(query, health_id)

	var client PatientDetails
	err := row.Scan(
		&client.HealthID, &client.FirstName, &client.MiddleName, &client.LastName, &client.Sex, &client.HealthcareID,
		&client.DOB, &client.BloodGroup, &client.BMI, &client.MarriageStatus, &client.Weight, &client.Email,
		&client.MobileNumber, &client.AadhaarNumber, &client.PrimaryLocation, &client.Sibling, &client.Twin,
		&client.FatherName, &client.MotherName, &client.EmergencyNumber, &client.CreatedAt, &client.UpdatedAt,
		&client.Address.Country, &client.Address.City, &client.Address.State, &client.Address.Landmark,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no client found with health ID: %s", health_id)
		}
		return nil, err
	}
	return &client, nil
}

func (s *PostgresStore) UpdateClientProfile(healthID string, updates map[string]interface{}) (*PatientDetails, error) {
	setClause := []string{}
	values := []interface{}{}
	counter := 1

	// Iterate over the updates map to prepare the SET clause
	for key, value := range updates {
		if value != "" && value != "N/A" && key != "healthcare_id" && key != "health_id" {
			setClause = append(setClause, fmt.Sprintf("%s = $%d", key, counter))
			values = append(values, value)
			counter++
		}
	}

	// If no valid fields to update, return an error
	if len(setClause) == 0 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	// Append the updated_at field to always update the timestamp
	setClause = append(setClause, fmt.Sprintf("updated_at = NOW()"))

	// Add the health_id as the last parameter for the WHERE clause
	values = append(values, healthID)

	// Construct the final SQL query
	query := fmt.Sprintf(`
		UPDATE client_profile
		SET %s
		WHERE health_id = $%d
		RETURNING *;
	`, strings.Join(setClause, ", "), counter)

	// Execute the update query
	row := s.db.QueryRow(query, values...)
	var updatedClient PatientDetails
	err := row.Scan(
		&updatedClient.ID, &updatedClient.HealthID, &updatedClient.FirstName, &updatedClient.MiddleName,
		&updatedClient.LastName, &updatedClient.Sex, &updatedClient.HealthcareID, &updatedClient.DOB,
		&updatedClient.BloodGroup, &updatedClient.BMI, &updatedClient.MarriageStatus, &updatedClient.Weight,
		&updatedClient.Email, &updatedClient.MobileNumber, &updatedClient.AadhaarNumber, &updatedClient.PrimaryLocation,
		&updatedClient.Sibling, &updatedClient.Twin, &updatedClient.FatherName, &updatedClient.MotherName,
		&updatedClient.EmergencyNumber, &updatedClient.CreatedAt, &updatedClient.UpdatedAt, &updatedClient.Address.Country,
		&updatedClient.Address.City, &updatedClient.Address.State, &updatedClient.Address.Landmark)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no client profile found with health_id %s", healthID)
		}
		return nil, err
	}
	return &updatedClient, nil
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

// get and set appointments for user
func (s *PostgresStore) GetAppointments(healthcare_id string, offset, limit int64) ([]*Appointments, error) {
	query := `SELECT id, health_id, status, appointment_date, appointment_time, healthcare_id, department, note, fullname, healthcare_name 
              FROM appointments WHERE healthcare_id = $1 `
	rows, err := s.db.Query(query, healthcare_id)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var appointments []*Appointments
	for rows.Next() {
		var appointment Appointments
		err := rows.Scan(
			&appointment.ID,
			&appointment.HealthID,
			&appointment.Status,
			&appointment.AppointmentDate,
			&appointment.AppointmentTime,
			&appointment.HealthcareID,
			&appointment.Department,
			&appointment.Note,
			&appointment.FullName,
			&appointment.HealthcareName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		appointments = append(appointments, &appointment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return appointments, nil
}

// Update appointment Status
func (s *PostgresStore) SetAppointments(healthcare_id, healthID, status string, id int64) (int64, error) {
	query := `UPDATE appointments SET status = $1 WHERE health_id = $2 AND healthcare_id = $3`
	result, err := s.db.Exec(query, status, healthcare_id, healthID)
	if err != nil {
		return 0, fmt.Errorf("failed to update appointments: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch rows affected: %w", err)
	}
	return rowsAffected, nil
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
