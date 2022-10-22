package main

import (
	"context"
	"encoding/json"
	"log"
	"net/mail"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/thanhpk/randstr"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// This struct is for what data will be saved to the database
type User struct {
	UUID         string `bson:"uuid,omitempty"`
	Username     string `bson:"username,omitempty"`
	Email        string `bson:"email,omitempty"`
	Password     string `bson:"password,omitempty"`
	Created_Date string `bson:"created_date,omitempty"`
}

// This is the post data to the Lambda function from the client
type Request struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// This is the response data back to the client
type Response struct {
	Body string `json:"message"`
}

// This is to allow Mongodb to be used by other functions easily
type Server struct {
	client *mongo.Client
}

// Creates the background context for Lambda
var ctx = context.Background()

// This is the main handler for the Lambda function
func (s *Server) handler(request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Creates a copy of the Request Struct
	var req Request
	//Saves all the data from the request to the Request Struct
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	//Checks if password is at a minimum length of 14 characters
	if len(req.Password) < 5 {
		//Returns a error response to the client
		return ResponseReturn("password must be longer than 14 characters", 422), nil
	}

	//Checks if name atleast has 2 characters
	if len(req.Username) < 1 {
		//Returns a error response to the client
		return ResponseReturn("first name must be longer than 1 character", 422), nil
	}

	//This is the database we are using
	apiDatabase := s.client.Database("Taplyy")
	userDB := apiDatabase.Collection("users")

	//Parses the Email from the post request to check if its valid
	_, err = mail.ParseAddress(strings.ToUpper(req.Email))
	if err != nil {
		//Returns a error response to the client
		return ResponseReturn("invalid email input", 400), nil
	}

	//Creates a bson.M object to save the query data
	var result bson.M
	//Checks to see if the email is already being used
	err = userDB.FindOne(ctx, bson.D{{Key: "email", Value: strings.ToUpper(req.Email)}}, nil).Decode(&result)
	if err == nil {
		//Returns a error response to the client
		return ResponseReturn("email already being used", 409), nil
	}
	//Checks to see if the phone number is already being used
	errs := userDB.FindOne(ctx, bson.D{{Key: "username", Value: req.Username}}, nil).Decode(&result)
	if errs == nil {
		//Returns a error response to the client
		return ResponseReturn("email already being used", 409), nil
	}

	//Hashes the password
	hash, _ := HashPassword(req.Password)
	//Creates the UUID
	UUID_Generated := randstr.Hex(16)

	//Gets the time stamp for the created date
	dt := time.Now()
	Time_NOW := dt.Format("01-02-2006 15:04:05")

	//Creates a copy of the User structs and popualtes it
	user := User{
		UUID:         UUID_Generated,
		Email:        strings.ToUpper(req.Email),
		Password:     hash,
		Created_Date: Time_NOW,
	}

	//Saves the user to the database
	_, err = userDB.InsertOne(ctx, user)
	if err != nil {
		//Returns a error response to the client
		return ResponseReturn("database error", 500), nil
	}
	//Returns a success response to the client
	return ResponseReturn("user is now registered", 200), nil

}

func main() {
	//Connection to the MongoDB database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://Taplyy:V0zYmLJR4XLDIyTo@taplyy.jfwdxnu.mongodb.net/?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	s := &Server{client}
	lambda.Start(s.handler)
}

// Hashes the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func ResponseReturn(message string, code int) events.APIGatewayProxyResponse {
	responseBody := Response{
		Body: message,
	}

	jbytes, err := json.Marshal(responseBody)
	if err != nil {
		return events.APIGatewayProxyResponse{}
	}

	response := events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       string(jbytes),
	}
	return response
}
