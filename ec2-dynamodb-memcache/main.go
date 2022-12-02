package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/bradfitz/gomemcache/memcache"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

// adapted from tutorial https://go.dev/doc/tutorial/web-service-gin

var mc = memcache.New("password-cache.4fxm87.cfg.use1.cache.amazonaws.com:11211")

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

	pwresponse, err := getPasswordByHash(shaHashRequest.ShaHash)
	if err != nil { // err is *crackstationAppError, not os.Error
		c.IndentedJSON(err.Code, gin.H{"requestedShaHash": shaHashRequest.ShaHash, "message": err.Message})
		return
	} else {
		// pwAsJSON, _ := json.Marshal(pwAsStruct)
		c.Data(http.StatusOK, "application/json", pwresponse)
	}
}

func getPasswordByHash(hash string) ([]byte, *crackstationAppError) {
	if hash == "" {
		return nil, &crackstationAppError{fmt.Errorf("missing shahash parameter"), "missing sha hash parameter", http.StatusBadRequest}
	} else {
		pwFromCache, err := checkCache(hash)
		if err != nil { // not cached

			// CHECK DYNAMODB
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

			// ADD TO CACHE

			cacheErr := mc.Set(&memcache.Item{Key: hash, Value: []byte(pw.Password), Expiration: 60})

			if cacheErr != nil {
				fmt.Printf("there was an error adding to cache")
				print(cacheErr)
			} else {
				fmt.Printf("{%s:%s} added to cache\n", hash, pw.Password)
			}

			responseBody := fmt.Sprintf("{\"%s\": \"%s\"}", pw.ShaHash, pw.Password)
			return []byte(responseBody), nil

		} else { // CACHED
			responseBody := fmt.Sprintf("{\"%s\": \"%s\"}", hash, fmt.Sprintf("%s", pwFromCache))

			return []byte(responseBody), nil
		}

	}
}

func getPassword(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in getPassword", r)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "unexpected error"})
		}
	}()
	hash := c.Param("shahash")

	pwresponse, err := getPasswordByHash(hash)
	if err != nil { // err is *crackstationAppError, not os.Error
		c.IndentedJSON(err.Code, gin.H{"requestedShaHash": hash, "message": err.Message})
		return
	} else {
		//pwAsJSON, _ := json.Marshal(pwAsStruct)
		c.Data(http.StatusOK, "application/json", pwresponse)
	}

}

func checkCache(hash string) ([]byte, error) {
	val, err := mc.Get(hash)
	if err != nil {
		fmt.Println("Cache Miss")
		return nil, err
	}
	fmt.Printf("yay! Cache Hit. Answer: %s\n", val.Value)
	return val.Value, nil

}

func missingShaHash(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"message": "missing shahash parameter"})
}

func main() {
	r := gin.Default()

	// Ping handler
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/password/:shahash", getPassword)
	r.GET("/password", missingShaHash)
	r.POST("/decrypt", postPasswords)

	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("api.v3.thecrackstation.com", "www.api.v3.thecrackstation.com"),
		Cache:      autocert.DirCache("/var/www/.cache"),
	}

	log.Fatal(autotls.RunWithManager(r, &m))
}
