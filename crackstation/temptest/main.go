package main

import (
	"fmt"

	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type password struct {
	shaHash  string `json:"shaHash"`
	password string `json:"password"`
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Obtain the QueryStringParameter
	shaHash := request.PathParameters["shaHash"]
	fmt.Println(shaHash)
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
			return events.APIGatewayProxyResponse{StatusCode: 404}, nil
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

		marshalledItem, err := json.Marshal(pw)

		return events.APIGatewayProxyResponse{Body: string(marshalledItem), StatusCode: 200}, nil

	} else {
		return events.APIGatewayProxyResponse{Body: "Error: Query Parameter name missing", StatusCode: 422}, nil
	}

}

func main() {
	lambda.Start(HandleRequest)
}
