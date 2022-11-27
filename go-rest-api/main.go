package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// adapted from tutorial https://go.dev/doc/tutorial/web-service-gin

type password struct {
	Shahash  string `json:"shaHash"`
	Password string `json:"password"`
}

// handful of passwords to mock
var passwords = []password{
	{Shahash: "86f7e437faa5a7fce15d1ddcb9eaeaea377667b8", Password: "a"},
	{Shahash: "e9d71f5ee7c92d6dc9e92ffdad17b8bd49418f98", Password: "b"},
	{Shahash: "84a516841ba77a5b4648de2cd0dfcb30ea46dbb4", Password: "c"},
}

// getPassword
func getPasswords(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, passwords)
}

func postPasswords(c *gin.Context) {
	var newPassword password

	if err := c.BindJSON(&newPassword); err != nil {
		return
	}

	passwords = append(passwords, newPassword)
	c.IndentedJSON(http.StatusCreated, newPassword)
}

func getPasswordByHash(c *gin.Context) {
	hash := c.Param("shahash")

	for _, p := range passwords {
		if p.Shahash == hash {
			c.IndentedJSON(http.StatusOK, p)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "password not found"})

}

func main() {
	router := gin.Default()
	router.GET("/passwords", getPasswords)
	router.GET("/passwords/:shahash", getPasswordByHash)
	router.POST("/passwords", postPasswords)

	router.Run("localhost:8080")
}
