//MongoDB - shadinachat@csu.fullerton.edu
//AWS - Ticketlyy

package main

import (
	"context"
	"encoding/json"
	"log"

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
	Score string `json:"Get_Scores"`
}

// This struct will be used to send back a message to the client
type Response struct {
	Body string `json:"message"`
}

// This struct will be used to decode the data from the database
type Score_Save struct {
	User  string `bson:"user,omitempty"`
	Score string `bson:"recent_score,omitempty"`
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

	//This is the database that we are using

	coll := s.client.Database("School").Collection("Scores")

	opts := options.Find().SetLimit(75)

	cursor, err := coll.Find(context.TODO(), bson.D{{}}, opts)
	if err != nil {
		return ResponseReturn("database error - find", 500), nil
	}
	// end find

	var resultst []Score_Save
	if err = cursor.All(context.TODO(), &resultst); err != nil {
		return ResponseReturn("database erro - curse", 500), nil
	}

	for _, result := range resultst {
		cursor.Decode(&result)
	}

	output, err := json.Marshal(resultst)
	if err != nil {
		return ResponseReturn("database error - marshal", 500), nil
	}
	//This is the response that will be sent back to the client
	return ResponseReturn(string(output), 200), nil
}

func main() {
	//This is the connection string to the database
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://School:Shadi1973@school.ix1gx2g.mongodb.net/?retryWrites=true&w=majority"))
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

	response := events.APIGatewayProxyResponse{
		StatusCode: code,
		Body:       message,
	}
	return response
}
