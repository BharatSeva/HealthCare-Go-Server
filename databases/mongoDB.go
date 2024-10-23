package databases

import (
	"context"
	"fmt"

	// "strconv"

	// "errors"
	// "fmt"

	// "time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	db         *mongo.Client
	database   string
	collection []string
}

type Server_message struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Connect to MongoDB
func ConnectToMongoDB(url, database string, collection []string) (*MongoStore, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to ensure connection is established
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("could not connect to MongoDB: %v", err)
	}

	return &MongoStore{
		db:         client,
		database:   database,
		collection: collection,
	}, nil
}

func (m *MongoStore) GetAppointments(healthcareID string, list int) ([]*Appointments, error) {
	coll := m.db.Database(m.database).Collection("appointments")

	filter := bson.D{{Key: "healthcare_id", Value: healthcareID}}
	findOptions := options.Find().SetLimit(int64(list))
	cursor, err := coll.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error finding appointments: %w", err)
	}
	defer cursor.Close(context.TODO())

	var appointments []*Appointments
	if err = cursor.All(context.TODO(), &appointments); err != nil {
		return nil, fmt.Errorf("error decoding appointments: %w", err)
	}
	return appointments, nil
}

func (m *MongoStore) CreatePatient_bioData(healthcareID int, patient *PatientDetails) (*PatientDetails, error) {
	patientDetails, err := CreatePatient_bioData(healthcareID, patient)
	if err != nil {
		return nil, err
	}
	patientDetails.ID = primitive.NewObjectID()
	coll := m.db.Database(m.database).Collection("patient_details")
	_, err = coll.InsertOne(context.TODO(), patientDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to insert patient details: %v", err)
	}

	return patientDetails, nil
}

func (m *MongoStore) GetPatient_bioData(patient_healthcareID string) (*PatientDetails, error) {
	coll := m.db.Database(m.database).Collection("patient_details")

	filter := bson.D{{Key: "health_id", Value: patient_healthcareID}}
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("error finding Patient with given id: %w", err)
	}
	defer cursor.Close(context.TODO())

	var patientdetails []PatientDetails
	if err = cursor.All(context.TODO(), &patientdetails); err != nil {
		return nil, fmt.Errorf("error decoding appointments: %w", err)
	}
	if len(patientdetails) == 0 {
		return nil, fmt.Errorf("no patient found with given patient_id %s, please create a new one", patient_healthcareID)
	}

	return &patientdetails[0], nil
}

func (m *MongoStore) CreateHealthcare_details(HIPInfo *HIPInfo) (*HIPInfo, error) {
	coll := m.db.Database(m.database).Collection("healthcare_info")
	_, err := coll.InsertOne(context.TODO(), HIPInfo)
	if err != nil {
		return nil, err
	}
	return HIPInfo, nil
}

func (m *MongoStore) GetHealthcare_details(healthcareId string) (*HIPInfo, error) {
	coll := m.db.Database(m.database).Collection("healthcare_info")
	filter := bson.D{{Key: "healthcare_id", Value: healthcareId}} // Match the BSON field name with struct tag
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("no healthcare found with given id: %s", healthcareId)
	}
	defer cursor.Close(context.TODO())

	var hipdetails []HIPInfo
	if err = cursor.All(context.TODO(), &hipdetails); err != nil {
		return nil, fmt.Errorf("error decoding HIP details: %w", err)
	}

	if len(hipdetails) == 0 {
		return nil, fmt.Errorf("no healthcare found with given id: %s", healthcareId)
	}

	// No manual date parsing required if MongoDB is storing it as ISODate
	return &hipdetails[0], nil
}

func (m *MongoStore) CreatepatientRecords(healthcare_id string, patientrecords *PatientRecords) (*PatientRecords, error) {
	coll := m.db.Database(m.database).Collection("patient_records")
	patientrecords, err := CreatePatientRecords(healthcare_id, patientrecords)
	if err != nil {
		return nil, err
	}
	id, err := coll.InsertOne(context.TODO(), patientrecords)
	if err != nil {
		return nil, err
	}
	if objectID, ok := id.InsertedID.(primitive.ObjectID); ok {
		patientrecords.ID = objectID
	} else {
		return nil, fmt.Errorf("failed to convert inserted ID to ObjectID")
	}
	return patientrecords, nil
}

func (m *MongoStore) GetPatientRecords(health_id string, list int) (*[]PatientRecords, error) {
	coll := m.db.Database(m.database).Collection("patient_records")
	filter := bson.D{{Key: "health_id", Value: health_id}}
	findOptions := options.Find().SetLimit(int64(list))
	cursor, err := coll.Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("error in database")
	}
	defer cursor.Close(context.TODO())
	var patientRecords []PatientRecords
	if err = cursor.All(context.TODO(), &patientRecords); err != nil {
		return nil, fmt.Errorf("error decoding patient records: %w", err)
	}
	return &patientRecords, nil
}

func (m *MongoStore) UpdatePatientBioData(healthID string, updates map[string]interface{}) (*PatientDetails, error) {
	coll := m.db.Database(m.database).Collection("patient_details")

	cleanedUpdates := map[string]interface{}{}
	for key, value := range updates {
		if value != "" && value != "N/A" {
			cleanedUpdates[key] = value
		}
	}

	if len(cleanedUpdates) == 0 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	filter := bson.D{{Key: "health_id", Value: healthID}}
	update := bson.D{{Key: "$set", Value: cleanedUpdates}}

	// Execute the update operation
	result, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("no document found with health_id %s", healthID)
	}

	if result.ModifiedCount == 0 {
		return nil, fmt.Errorf("no fields were updated for health_id %s", healthID)
	}

	// Retrieve and return the updated patient details
	var updatedPatient PatientDetails
	err = coll.FindOne(context.Background(), filter).Decode(&updatedPatient)
	if err != nil {
		return nil, err
	}

	return &updatedPatient, nil
}

// func (m *MongoStore) TransferAmount(fromID, toID, amount int) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	session, err := m.db.StartSession()
// 	if err != nil {
// 		return err
// 	}
// 	defer session.EndSession(ctx)

// 	err = session.StartTransaction()
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err != nil {
// 			if abortErr := session.AbortTransaction(ctx); abortErr != nil {
// 				fmt.Println("Error aborting transaction:", abortErr)
// 			}
// 		}
// 	}()

// 	coll := m.db.Database(m.database).Collection(m.collection)

// 	// Update sender balance (with check for sufficient funds)
// 	updateResult, err := coll.UpdateOne(ctx, bson.M{"id": fromID, "balance": bson.M{"$gte": amount}}, bson.M{"$inc": bson.M{"balance": -amount}})
// 	if err != nil {
// 		return err
// 	}

// 	if updateResult.ModifiedCount == 0 {
// 		return errors.New("insufficient funds")
// 	}

// 	// Update receiver balance
// 	_, err = coll.UpdateOne(ctx, bson.M{"id": toID}, bson.M{"$inc": bson.M{"balance": amount}})
// 	if err != nil {
// 		return err
// 	}

// 	err = session.CommitTransaction(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("Transfer successful!")
// 	return nil
// }
