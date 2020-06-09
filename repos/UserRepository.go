package repos

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dataBase *mongo.Database

func Init() {
	var localMasterKey []byte // This must be the same master key that was used to create the encryption key.
	kmsProviders := map[string]map[string]interface{}{
		"local": {
			"key": localMasterKey,
		},
	}

	// The MongoDB namespace (db.collection) used to store the encryption data keys.
	keyVaultDBName, keyVaultCollName := "encryption", "testKeyVault"
	keyVaultNamespace := keyVaultDBName + "." + keyVaultCollName

	// Create the Client for reading/writing application data. Configure it with BypassAutoEncryption=true to disable
	// automatic encryption but keep automatic decryption. Setting BypassAutoEncryption will also bypass spawning
	// mongocryptd in the driver.
	autoEncryptionOpts := options.AutoEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace).
		SetBypassAutoEncryption(true)
	clientOpts := options.Client().
		ApplyURI("mongodb://localhost:27017").
		SetAutoEncryptionOptions(autoEncryptionOpts)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		panic(err)
	}
	defer func() { client.Disconnect(context.TODO()) }()

	// Get a handle to the application collection and clear existing data.
	//dataBase = client.Database("MosqueOnline")
	coll := client.Database("test").Collection("coll")
	_ = coll.Drop(context.TODO())

	// Set up the key vault for this example.
	keyVaultColl := client.Database(keyVaultDBName).Collection(keyVaultCollName)
	_ = keyVaultColl.Drop(context.TODO())
	// Ensure that two data keys cannot share the same keyAltName.
	keyVaultIndex := mongo.IndexModel{
		Keys: bson.D{{"keyAltNames", 1}},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.D{
				{"keyAltNames", bson.D{
					{"$exists", true},
				}},
			}),
	}
	if _, err = keyVaultColl.Indexes().CreateOne(context.TODO(), keyVaultIndex); err != nil {
		panic(err)
	}

	// Create the ClientEncryption object to use for explicit encryption/decryption. The Client passed to
	// NewClientEncryption is used to read/write to the key vault. This can be the same Client used by the main
	// application.
	clientEncryptionOpts := options.ClientEncryption().
		SetKmsProviders(kmsProviders).
		SetKeyVaultNamespace(keyVaultNamespace)
	clientEncryption, err := mongo.NewClientEncryption(client, clientEncryptionOpts)
	if err != nil {
		panic(err)
	}
	defer func() { _ = clientEncryption.Close(context.TODO()) }()

	// Create a new data key for the encrypted field.
	dataKeyOpts := options.DataKey().SetKeyAltNames([]string{"go_encryption_example"})
	dataKeyID, err := clientEncryption.CreateDataKey(context.TODO(), "local", dataKeyOpts)
	if err != nil {
		panic(err)
	}

	// Create a bson.RawValue to encrypt and encrypt it using the key that was just created.
	rawValueType, rawValueData, err := bson.MarshalValue("123456789")
	if err != nil {
		panic(err)
	}
	rawValue := bson.RawValue{Type: rawValueType, Value: rawValueData}
	encryptionOpts := options.Encrypt().
		SetAlgorithm("AEAD_AES_256_CBC_HMAC_SHA_512-Deterministic").
		SetKeyID(dataKeyID)
	encryptedField, err := clientEncryption.Encrypt(context.TODO(), rawValue, encryptionOpts)
	if err != nil {
		panic(err)
	}

	// Insert a document with the encrypted field and then find it. The FindOne call will automatically decrypt the
	// field in the document.
	if _, err = coll.InsertOne(context.TODO(), bson.D{{"encryptedField", encryptedField}}); err != nil {
		panic(err)
	}
	var foundDoc bson.M
	if err = coll.FindOne(context.TODO(), bson.D{}).Decode(&foundDoc); err != nil {
		panic(err)
	}
	fmt.Printf("Decrypted document: %v\n", foundDoc)
}

// TODO: istead integers use constants like const admin for getting admin collection
func GetDBCollection(i int) (*mongo.Collection, error) {
	// Define Address of Database
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	// Try to connect to Database, save error if one is thrown
	client, err := mongo.Connect(context.TODO(), clientOptions)
	// If there was an error connecting to the DB (DB not running, wrong URI, ...) return the error
	if err != nil {
		return nil, err
	}

	// Check if connection could be established to running DB

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	// Define the name of the Database as PPT, change this if you want to name your DB otherwise
	db := client.Database("MosqueOnline")
	// Working with int for extensibility, you can just add another else if and check for another value if you want to add another table
	// Get the Users Table
	if i == 0 {
		return db.Collection("users"), nil
	} else if i == 1 {
		// Get The Mosques Table with the entries of the Mosques
		return db.Collection("mosques"), nil
	} else if i == 2 {
		// Get The Mosques Table with the entries of the Mosques
		return db.Collection("admins"), nil
	}
	return nil, errors.New("Not a valid Databasequery")

}
