package databases

import "fmt"

type CombinedStore struct {
	postgres *PostgresStore
	mongodb  *MongoStore
}  

func NewCombinedStore(postgresConn string, mongoURI string, dbName string, collection []string) (*CombinedStore, error) {
	postgres, err := ConnectToPostgreSQL(postgresConn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %w", err)
	}
	if err := postgres.Init(); err != nil {
		return nil, fmt.Errorf("failed to init postgres: %w", err)
	}

	mongodb, err := ConnectToMongoDB(mongoURI, dbName, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mongodb: %w", err)
	}
	if err := mongodb.Init(); err != nil {
		return nil, fmt.Errorf("failed to init mongodb: %w", err)
	}

	return &CombinedStore{
		postgres: postgres,
		mongodb:  mongodb,
	}, nil
}

// for each methods define which database methods will be called
// Since we have two database each one of have it's own methods
// This allows us to add more databases sequentially

func (s *CombinedStore) SignUpAccount(hipinfo *HIPInfo) (int64, error) {
	return s.postgres.SignUpAccount(hipinfo)
}

func (s *CombinedStore) LoginUser(login *Login) (*HIPInfo, error) {
	return s.postgres.LoginUser(login)
}

func (s *CombinedStore) ChangePreferance(id string, pref map[string]interface{}) error {
	return s.postgres.ChangePreferance(id, pref)
}

func (s *CombinedStore) GetPreferance(id string) (*ChangePreferance, error) {
	return s.postgres.GetPreferance(id)
}

func (s *CombinedStore) GetAppointments(id string, list int) ([]*Appointments, error) {
	return s.mongodb.GetAppointments(id, list)
}

func (s *CombinedStore) CreatePatient_bioData(id string, details *PatientDetails) (*PatientDetails, error) {
	return s.mongodb.CreatePatient_bioData(id, details)
}

func (s *CombinedStore) GetPatient_bioData(healthID string) (*PatientDetails, error) {
	return s.mongodb.GetPatient_bioData(healthID)
}

func (s *CombinedStore) GetHealthcare_details(id string) (*HIPInfo, error) {
	return s.mongodb.GetHealthcare_details(id)
}

func (s *CombinedStore) CreatepatientRecords(healthID string, records *PatientRecords) (*PatientRecords, error) {
	return s.mongodb.CreatepatientRecords(healthID, records)
}

func (s *CombinedStore) GetPatientRecords(healthID string, limit int) (*[]PatientRecords, error) {
	return s.mongodb.GetPatientRecords(healthID, limit)
}

func (s *CombinedStore) UpdatePatientBioData(healthID string, updates map[string]interface{}) (*PatientDetails, error) {
	return s.mongodb.UpdatePatientBioData(healthID, updates)
}
func (s *CombinedStore) CreateHealthcare_details(healthcare_info *HIPInfo) (*HIPInfo, error) {
	return s.mongodb.CreateHealthcare_details(healthcare_info)
}
