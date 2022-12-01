package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type password struct {
	ShaHash  string `json:"shaHash"`
	Password string `json:"password"`
}

type uncrackedResponse struct {
	RequestedShaHash string `json:"requestedShaHash"`
	Error            string `json:"Error"`
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	shaHash := request.PathParameters["shaHash"]
	if shaHash != "" {
		var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))
		// Prepare the input for the query.
		input := &dynamodb.GetItemInput{
			TableName: aws.String("rainbowlookup"),
			Key: map[string]*dynamodb.AttributeValue{
				"shaHash": {
					S: aws.String(shaHash),
				},
			},
		}

		// Retrieve the item from DynamoDB. If no matching item is found
		// return nil.
		result, err := db.GetItem(input)
		if err != nil {
			fmt.Println(err.Error())
			return events.APIGatewayProxyResponse{StatusCode: 500}, nil
		}
		if len(result.Item) == 0 {
			fmt.Println("No results found!")
			noresults := &uncrackedResponse{
				RequestedShaHash: shaHash,
				Error:            "Could not crack. This hash is unknown to CrackStation.",
			}
			body, _ := json.Marshal(noresults)
			return events.APIGatewayProxyResponse{Body: string(body), StatusCode: 404}, nil
		}

		// The result.Item object returned has the underlying type
		// map[string]*AttributeValue. We can use the UnmarshalMap helper
		// to parse this straight into the fields of a struct. Note:
		// UnmarshalListOfMaps also exists if you are working with multiple
		// items.
		pw := new(password)

		err = dynamodbattribute.UnmarshalMap(result.Item, pw)
		if err != nil {
			panic(fmt.Sprintf("Failed to UnmarshalMap result.Item: ", err))
		}

		//js, err := json.Marshal(pw)
		responseBody := fmt.Sprintf("{\"%s\": \"%s\"}", pw.ShaHash, pw.Password)
		

		return events.APIGatewayProxyResponse{Body: responseBody, StatusCode: 200}, nil

	} else {
		return events.APIGatewayProxyResponse{Body: "Error: Query Parameter name missing", StatusCode: 404}, nil
	}

}

func main() {
	lambda.Start(HandleRequest)
}
