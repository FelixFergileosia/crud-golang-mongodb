package userService

import (
	"context"
	"crud-golang/database"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	// v2 path

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	dbName         = "felix-cluster" // <-- change to your DB name
	collectionName = "users"
)

func usersColl(client *mongo.Client) *mongo.Collection {
	return client.Database(dbName).Collection(collectionName)
}

type User struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string        `bson:"name"        json:"name"`
	Email string        `bson:"email"       json:"email"`
}
type ErrorResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}
type DeleteUserRequest struct {
	ID string `json:"id"`
}

// CreateUser inserts a user into the database.
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser User
	if err := json.NewDecoder(r.Body).Decode(&newUser); handleErr(w, "invalid request body", err) {
		return
	}
	client, err := database.DbConnection()
	fmt.Println(client, err)

	if handleErr(w, "DB connection failed", err) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)

	res, err := usersColl(client).InsertOne(ctx, newUser)
	if handleErr(w, "insert failed", err) {
		return
	}

	newUser.ID = res.InsertedID.(bson.ObjectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newUser)
}

// GetUsers retrieves all users from the database.
func GetUsers(w http.ResponseWriter, r *http.Request) {
	client, err := database.DbConnection()
	if handleErr(w, "DB connection failed", err) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)

	cursor, err := usersColl(client).Find(ctx, bson.M{})
	if handleErr(w, "query failed", err) {
		return
	}
	defer cursor.Close(ctx)
	fmt.Print(cursor)

	var users []User
	if err = cursor.All(ctx, &users); handleErr(w, "cursor decode failed", err) {
		return
	}
	fmt.Print(users)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// GetUserByID retrieves a user from the database by ID.
func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if username == "" {
		http.Error(w, "missing username", http.StatusBadRequest)
		return
	}
	client, err := database.DbConnection()
	if handleErr(w, "DB connection failed", err) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	var user User
	filter := bson.M{"name": username}
	err2 := usersColl(client).FindOne(ctx, filter).Decode(&user)
	if err2 != nil {
		if err2 == mongo.ErrNoDocuments {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			handleErr(w, "query failed", err2)
		}
		return
	}
	// 4) Return the single user as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// // UpdateUser updates the data of a user in the database.
// func UpdateUser(w http.ResponseWriter, r *http.Request) {

// }

// DeleteUser deletes a user from the database.
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   0,
			Message: "Missing ID in the URL",
		})
		return
	}
	objectId, err := bson.ObjectIDFromHex(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   0,
			Message: "Invalid ID format",
		})
		return
	}
	client, err := database.DbConnection()
	if handleErr(w, "DB connection failed", err) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer client.Disconnect(ctx)
	filter := bson.M{"_id": objectId}
	res, err2 := usersColl(client).DeleteOne(ctx, filter)
	if handleErr(w, "deletion failed", err2) {
		return
	}
	if res.DeletedCount == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   1,
			Message: "User Not Found",
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// func handleGenericError(w http.ResponseWriter, errorMessage string, err error) bool {

// }

/* --------------------------------------------------------------------- *
 * ERROR HELPER                                                          *
 * --------------------------------------------------------------------- */
func handleErr(w http.ResponseWriter, msg string, err error) bool {
	if err != nil {
		http.Error(w, msg+": "+err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}
