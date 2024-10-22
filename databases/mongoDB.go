package databases

import (
	"context"
	"fmt"
	"strconv"

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
	collection string
}

type Server_message struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Connect to MongoDB
func ConnectToMongoDB(url, database, collection string) (*MongoStore, error) {
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

func (m *MongoStore) GetAppointments(healthcareID int) ([]*Appointments, error) {
    coll := m.db.Database(m.database).Collection(m.collection)
    
	healthcareIDStr := strconv.Itoa(healthcareID)
    filter := bson.D{{"healthcareID", healthcareIDStr}}
    
    cursor, err := coll.Find(context.TODO(), filter)
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
	coll := m.db.Database(m.database).Collection(m.collection)
	_, err = coll.InsertOne(context.TODO(), patientDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to insert patient details: %v", err)
	}

	return patientDetails, nil
}

func (m *MongoStore) GetPatient_bioData(patient_healthcareID string) (*PatientDetails, error){
	coll := m.db.Database(m.database).Collection(m.collection)
    
    filter := bson.D{{"health_id", patient_healthcareID}}
    cursor, err := coll.Find(context.TODO(), filter)
    if err != nil {
        return nil, fmt.Errorf("error finding Patient with given id: %w", err)
    }
    defer cursor.Close(context.TODO())

    var patientdetails []PatientDetails
    if err = cursor.All(context.TODO(), &patientdetails); err != nil {
        return nil, fmt.Errorf("error decoding appointments: %w", err)
    }
	if len(patientdetails)==0 {
		return nil, fmt.Errorf("no patient found with given patient_id %s, please create a new one", patient_healthcareID)
	}

    return &patientdetails[0], nil
}

// func (m *MongoStore) DeleteAccount(id int) error {

// 	coll := m.db.Database(m.database).Collection(m.collection)
// 	// idNumber, _ := primitive.ObjectIDFromHex(id)
// 	filter := bson.D{{"id", id}}

// 	DeletedCount, err := coll.DeleteOne(context.TODO(), filter)
// 	if err != nil {
// 		return err
// 	}

// 	if DeletedCount.DeletedCount != 1 {
// 		return fmt.Errorf("no account found with this Id: %d", id)
// 	}
// 	return nil
// }

// func (m *MongoStore) UpdateAccount(acc *Account) error {

// 	coll := m.db.Database(m.database).Collection(m.collection)
// 	// idNumber, _ := primitive.ObjectIDFromHex(id)
// 	filter := bson.D{{"id", acc.ID}}

// 	update := bson.D{{"$set", bson.D{{"firstname", acc.FirstName}, {"lastname", acc.LastName}, {"balance", acc.Balance}}}}
// 	UpdateResult, err := coll.UpdateOne(context.TODO(), filter, update)
// 	if err != nil {
// 		return err
// 	}
// 	if UpdateResult.MatchedCount != 1 {
// 		return fmt.Errorf("no user found with id: %d", acc.ID)
// 	}
// 	return nil
// }

// func (m *MongoStore) GetAccounts() ([]*Account, error) {
// 	coll := m.db.Database(m.database).Collection(m.collection)
// 	filter := bson.D{{}}

// 	cursor, err := coll.Find(context.TODO(), filter)
// 	if err != nil {
// 		panic(err)
// 	}

// 	acc := []*Account{}
// 	if err = cursor.All(context.TODO(), &acc); err != nil {
// 		panic(err)
// 	}

// 	return acc, nil
// }

// func (m *MongoStore) GetAccountByID(id int) (*Account, error) {
// 	coll := m.db.Database(m.database).Collection(m.collection)
// 	// idNumber, _ := primitive.ObjectIDFromHex(id)
// 	filter := bson.D{{"id", id}}

// 	account := Account{}
// 	err := coll.FindOne(context.TODO(), filter).Decode(&account)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return nil, err
// 		}
// 		panic(err)
// 	}
// 	return &account, nil
// }

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
