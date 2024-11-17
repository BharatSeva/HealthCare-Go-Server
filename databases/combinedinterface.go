package databases

import (
	"fmt"
	"time"
	mq "vaibhavyadav-dev/healthcareServer/rabbitmq"
	rd "vaibhavyadav-dev/healthcareServer/redis"

	// "go.mongodb.org/mongo-driver/bson/primitive"
)

type CombinedStore struct {
	postgres  *PostgresStore
	mongodb   *MongoStore
	rabbitmq  *mq.Rabbitmq
	redisconn *rd.Redisconn
}

// redis will contain url, limit -> no request allowed in window time
func Combinedstore(redisURL string, limit int64, window time.Duration, rabbitMqURL, postgresConn, mongoURI string, dbName string, collection []string) (*CombinedStore, error) {
	postgres, err := ConnectToPostgreSQL(postgresConn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %s", err.Error())
	}
	if err := postgres.Init(); err != nil {
		return nil, fmt.Errorf("failed to init postgres: %s", err.Error())
	}

	mongodb, err := ConnectToMongoDB(mongoURI, dbName, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize mongodb: %s", err.Error())
	}
	if err := mongodb.Init(); err != nil {
		return nil, fmt.Errorf("failed to init mongodb: %s", err.Error())
	}

	// this one for rabbitmq
	rabbitmqconn, err := mq.Connect2rabbitmq(rabbitMqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres: %s", err.Error())
	}

	// this one for redis
	redisconn, err := rd.Connect2Redis(redisURL, limit, window)
	if err != nil {
		return nil, fmt.Errorf("failed to init postgres: %s", err.Error())
	}

	return &CombinedStore{
		postgres:  postgres,
		mongodb:   mongodb,
		rabbitmq:  rabbitmqconn,
		redisconn: redisconn,
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

func (s *CombinedStore) GetPreferance(id string) (*Preferance, error) {
	return s.postgres.GetPreferance(id)
}

func (s *CombinedStore) GetTotalRequestCount(healthcare_id string) (int, error) {
	return s.postgres.GetTotalRequestCount(healthcare_id)
}

func (s *CombinedStore) CreateClient_stats(health_id string) error {
	return s.postgres.CreateClient_stats(health_id)
}
func (s *CombinedStore) GetAppointments_postgres(health_id string, offset, limit int64) ([]*Appointments, error) {
	return s.postgres.GetAppointments(health_id, offset, limit)
}
func (s *CombinedStore) SetAppointments_postgres(healthcare_id, health_id, status string, id int64) (int64, error) {
	return s.postgres.SetAppointments(healthcare_id, health_id, status, id)
}
// Get Healthcare_Profile
func (s *CombinedStore) GetHealthcare_details_postgres(healthcare_id string) (*HIPInfo, error){
	return s.postgres.GetHealthcare_details(healthcare_id)
}

// Create Client_Profile
func (s *CombinedStore) Create_ClientProfile(client *PatientDetails) error {
	return s.postgres.Create_ClientProfile(client)
}

// Get Client_Profile
func (s *CombinedStore) Get_ClientProfile(health_id string) (*PatientDetails, error) {
	return s.postgres.Get_ClientProfile(health_id)
}

// Update Client_Profile
func (s *CombinedStore) Update_clientProfile(health_id string, update map[string]interface{}) (*PatientDetails, error){
	return s.postgres.UpdateClientProfile(health_id, update);
}


// mongodb methods goes here.....
func (s *CombinedStore) GetAppointments(id string, list int64) ([]*Appointments, error) {
	return s.mongodb.GetAppointments(id, list)
}

func (s *CombinedStore) SetAppointments(healthcare_id, health_id, status string, id int64) (*Appointments, error) {
	return s.mongodb.SetAppointments(healthcare_id, health_id, status, id)
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

func (s *CombinedStore) GetPatientRecords(healthID, severity string, limit int) (*[]PatientRecords, error) {
	return s.mongodb.GetPatientRecords(healthID, severity, limit)
}

func (s *CombinedStore) UpdatePatientBioData(healthID string, updates map[string]interface{}) (*PatientDetails, error) {
	return s.mongodb.UpdatePatientBioData(healthID, updates)
}
func (s *CombinedStore) CreateHealthcare_details(healthcare_info *HIPInfo) (*HIPInfo, error) {
	return s.mongodb.CreateHealthcare_details(healthcare_info)
}

// /// ///////////////////////////////////////////
// ////////////////////// 	These are counters

// func (s *CombinedStore) Recordsviewed_counter(healthcare_id string) error {
// 	return s.postgres.Recordsviewed_counter(healthcare_id)
// }
// func (s *CombinedStore) Recordscreated_counter(healthcare_id string) error {
// 	return s.postgres.Recordscreated_counter(healthcare_id)
// }
// func (s *CombinedStore) Patientbiodata_created_counter(healthcare_id string) error {
// 	return s.postgres.Patientbiodata_created_counter(healthcare_id)
// }
// func (s *CombinedStore) Patientbiodata_viewed_counter(healthcare_id string) error {
// 	return s.postgres.Patientbiodata_viewed_counter(healthcare_id)
// }
// //////////////////////////////////////////////////////////
///////////////////////////////////////////////////////

// rabbitmq implementation goes here
func (s *CombinedStore) Push_counters(category, healthcare_id string) error {
	return s.rabbitmq.Push_counters(category, healthcare_id)
}
func (s *CombinedStore) Push_logs(category, name, email, health_id, healthcare_name, healthcare_id interface{}) error {
	return s.rabbitmq.Push_logs(category, name, email, health_id, healthcare_name, healthcare_id)
}
func (s *CombinedStore) Push_update_appointment(appointment map[string]interface{}) error {
	return s.rabbitmq.Push_update_appointment(appointment)
}

func (s *CombinedStore) Push_patient_records(record map[string]interface{}) error {
	return s.rabbitmq.Push_patient_records(record)
}

func (s *CombinedStore) Push_patientbiodata(biodata map[string]interface{}) error {
	return s.rabbitmq.Push_patientbiodata(biodata)
}

// Redis implementation
func (s *CombinedStore) Set(key string, value interface{}) error {
	return s.redisconn.Set(key, value)
}

func (s *CombinedStore) Get(key string) (interface{}, error) {
	return s.redisconn.Get(key)
}

//	RATE LIMITER GOES HERE...
//
// this one is for rate limiting (rate limiter)
func (s *CombinedStore) IsAllowed(healthcare_id string) (bool, error) {
	return s.redisconn.IsAllowed(healthcare_id)
}

func (s *CombinedStore) IsAllowed_leaky_bucket(healthcare_id string) (bool, error) {
	return s.redisconn.IsAllowed_leaky_bucket(healthcare_id)
}

func (s *CombinedStore) Close() error {
	return s.redisconn.Close()
}
