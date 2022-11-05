//MongoDB - shadinachat@csu.fullerton.edu
//AWS - Ticketlyy

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/mail"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This is to make MongoDB work in diffrent functions easily
type Server struct {
	client *mongo.Client
}

// This will take the input Post request and decode it into a struct
type Requests struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// This struct will be used to send back a message to the client
type Response struct {
	Body string `json:"message"`
}

// This struct will be used to decode the data from the database
type Email_Subscribe struct {
	User       string `bson:"user,omitempty"`
	Email      string `bson:"email,omitempty"`
	Password   string `bson:"password,omitempty"`
	High_Score int    `bson:"high_score,omitempty"`
}

// Creates a context for the Lambda
var ctx = context.Background()

// This is the main handler that is called when the Lambda is executed
func (s *Server) handler(request events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	//Creates a copy of the "Requests" structs
	var req Requests

	//Unmarshals all the data to the Requests struct
	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	//Parses the Email from the post request to check if its valid
	_, err = mail.ParseAddress(strings.ToUpper(req.Email))
	if err != nil {
		//Returns a error response to the client
		return ResponseReturn("invalid email input", 400), nil
	}

	//This is the database that we are using
	apiDatabase := s.client.Database("School")
	mailingDB := apiDatabase.Collection("Users")

	email := Email_Subscribe{
		User:       req.User,
		Email:      req.Email,
		Password:   req.Password,
		High_Score: 0,
	}

	var result_user Email_Subscribe
	err = mailingDB.FindOne(ctx, bson.D{{Key: "email", Value: req.Email}}, nil).Decode(&result_user)
	if err == nil {
		return ResponseReturn("email already exists", 409), nil
	}

	_, err = mailingDB.InsertOne(context.TODO(), email)
	if err == nil {

	}
	return ResponseReturn("account created", 200), nil

	//If password or email is incorrect return an error
}

func main() {
	//This is the connection string to the database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://School:Shadi1973@school.q6mrlky.mongodb.net/?retryWrites=true&w=majorityy"))
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

	//This is the Lambda handler
	lambda.Start(s.handler)
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
