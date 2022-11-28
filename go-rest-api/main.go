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

type PostRequest struct {
	ShaHash string `json:"sha_hash"`
}

func postPasswords(c *gin.Context) {
	var shaHashRequest PostRequest

	if err := c.BindJSON(&shaHashRequest); err != nil || shaHashRequest.ShaHash == "" {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON or no sha_hash present. See documentation."})
		return
	}

	pwAsStruct, err := getPasswordByHash(shaHashRequest.ShaHash)
	if err != nil { // err is *crackstationAppError, not os.Error
		c.IndentedJSON(err.Code, gin.H{"requestedShaHash": shaHashRequest.ShaHash, "message": err.Message})
		return
	} else {
		pwAsJSON, _ := json.Marshal(pwAsStruct)
		c.Data(http.StatusOK, "application/json", pwAsJSON)
	}
}

func getPassword(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected error"})
		}
	}()
	hash := c.Param("shahash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "missing shahash parameter"})
		return
	} else {
		pwAsStruct, err := getPasswordByHash(hash)
		if err != nil { // err is *crackstationAppError, not os.Error
			c.JSON(err.Code, gin.H{"requestedShaHash": hash, "message": err.Message})
			return
		} else {
			pwAsJSON, _ := json.Marshal(pwAsStruct)
			c.Data(http.StatusOK, "application/json", pwAsJSON)
		}
	}

}

func getPasswordByHash(hash string) (*password, *crackstationAppError) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, &crackstationAppError{err, "could not reach dynamoDB. GetItem failed.", 500}
	}
	var db = dynamodb.New(sess, aws.NewConfig().WithRegion("us-east-1"))

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
	c.JSON(http.StatusBadRequest, gin.H{"message": "missing shahash parameter"})
}

func main() {
	router := gin.Default()
	router.GET("/passwords/:shahash", getPassword)
	router.GET("/passwords", missingShaHash)
	router.POST("/decrypt", postPasswords)

	router.Run("localhost:8080")
}
