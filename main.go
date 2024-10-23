package main

import (
	"fmt"
	"log"
	db "vaibhavyadav-dev/healthcareServer/databases"
	mod "vaibhavyadav-dev/healthcareServer/databases"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

type CombinedStore struct {
	postgres *db.PostgresStore
	mongodb  *db.MongoStore
}

func (s *CombinedStore) SignUpAccount(hipinfo *mod.HIPInfo) (int64, error) {
	return s.postgres.SignUpAccount(hipinfo)
}

func (s *CombinedStore) LoginUser(login *mod.Login) (*mod.HIPInfo, error) {
	return s.postgres.LoginUser(login)
}

func (s *CombinedStore) ChangePreferance(id string, pref map[string]interface{}) error {
	return s.postgres.ChangePreferance(id, pref)
}

func (s *CombinedStore) GetPreferance(id int) (*mod.ChangePreferance, error) {
	return s.postgres.GetPreferance(id)
}

func (s *CombinedStore) GetAppointments(id string, list int) ([]*mod.Appointments, error) {
	return s.mongodb.GetAppointments(id, list)
}

func (s *CombinedStore) CreatePatient_bioData(id int, details *mod.PatientDetails) (*mod.PatientDetails, error) {
	return s.mongodb.CreatePatient_bioData(id, details)
}

func (s *CombinedStore) GetPatient_bioData(healthID string) (*mod.PatientDetails, error) {
	return s.mongodb.GetPatient_bioData(healthID)
}

func (s *CombinedStore) GetHealthcare_details(id string) (*mod.HIPInfo, error) {
	return s.mongodb.GetHealthcare_details(id)
}

func (s *CombinedStore) CreatepatientRecords(healthID string, records *mod.PatientRecords) (*mod.PatientRecords, error) {
	return s.mongodb.CreatepatientRecords(healthID, records)
}

func (s *CombinedStore) GetPatientRecords(healthID string, limit int) (*[]mod.PatientRecords, error) {
	return s.mongodb.GetPatientRecords(healthID, limit)
}

func (s *CombinedStore) UpdatePatientBioData(healthID string, updates map[string]interface{}) (*mod.PatientDetails, error) {
	return s.mongodb.UpdatePatientBioData(healthID, updates)
}
func (s *CombinedStore) CreateHealthcare_details(healthcare_info *mod.HIPInfo) (*mod.HIPInfo, error) {
	return s.mongodb.CreateHealthcare_details(healthcare_info)
}

func NewCombinedStore(postgresConn string, mongoURI string, dbName string, collection []string) (*CombinedStore, error) {
	postgres, err := db.NewPostgresStore(postgresConn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %w", err)
	}
	if err := postgres.Init(); err != nil {
		return nil, fmt.Errorf("failed to init postgres: %w", err)
	}

	mongodb, err := db.ConnectToMongoDB(mongoURI, dbName, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mongodb: %w", err)
	}

	return &CombinedStore{
		postgres: postgres,
		mongodb:  mongodb,
	}, nil
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	mongoURI := "mongodb+srv://vaibhavwateam:OpgbfoVf55RObbJN@bharatsevacluster.sbo5r.mongodb.net/?retryWrites=true&w=majority&appName=BharatSevaCluster"

	store, err := NewCombinedStore(psqlInfo, mongoURI, "db", []string{"golang1", "golang2", "golang3", "golang4"})
	if err != nil {
		log.Fatal("Failed to initialize store:", err)
	}
	server := NewAPIServer(":3000", store)

	// Start the Server
	server.Run()
}
