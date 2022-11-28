package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gin-gonic/gin"
)

// adapted from tutorial https://go.dev/doc/tutorial/web-service-gin

type password struct {
	ShaHash  string `json:"shaHash"`
	Password string `json:"password"`
}

type crackstationAppError struct {
	Error   error
	Message string
	Code    int
}

// handful of passwords to mock
var passwords = []password{
	{ShaHash: "86f7e437faa5a7fce15d1ddcb9eaeaea377667b8", Password: "a"},
	{ShaHash: "e9d71f5ee7c92d6dc9e92ffdad17b8bd49418f98", Password: "b"},
	{ShaHash: "84a516841ba77a5b4648de2cd0dfcb30ea46dbb4", Password: "c"},
}

func postPasswords(c *gin.Context) {
	var newPassword password

	if err := c.BindJSON(&newPassword); err != nil {
		return
	}

	passwords = append(passwords, newPassword)
	c.IndentedJSON(http.StatusCreated, newPassword)
}

func getPassword(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "unexpected error"})
		}
	}()
	hash := c.Param("shahash")
	if hash == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "missing shahash parameter"})
		return
	} else {
		pwAsStruct, err := getPasswordByHash(hash)
		if err != nil { // err is *crackstationAppError, not os.Error
			c.IndentedJSON(err.Code, gin.H{"requestedShaHash": hash, "message": err.Message})
			return
		} else {
			pwAsJSON, _ := json.Marshal(pwAsStruct)
			c.Data(http.StatusOK, "application/json", pwAsJSON)
		}
	}

}

func getPasswordByHash(hash string) (*password, *crackstationAppError) {
	var db = dynamodb.New(session.New(), aws.NewConfig().WithRegion("us-east-1"))

	input := &dynamodb.GetItemInput{
		TableName: aws.String("rainbowlookup"),
		Key: map[string]*dynamodb.AttributeValue{
			"shaHash": {
				S: aws.String(hash),
			},
		},
	}

	result, err := db.GetItem(input)
	if err != nil {
		return nil, &crackstationAppError{err, "could not reach dynamoDB. GetItem failed.", 500}
	}

	if len(result.Item) == 0 {
		return nil, &crackstationAppError{fmt.Errorf("not found:  %s", hash), "no results found", 404}
	}

	pw := password{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &pw)

	if err != nil {
		panic(fmt.Sprintf("Failed to UnmarshalMap result.Item: %s", err.Error()))
	}

	return &pw, nil

}

func missingShaHash(c *gin.Context) {
	c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "missing shahash parameter"})
}

func main() {
	router := gin.Default()
	router.GET("/passwords/:shahash", getPassword)
	router.GET("/passwords", missingShaHash)
	router.POST("/passwords", postPasswords)

	router.Run("localhost:8080")
}
