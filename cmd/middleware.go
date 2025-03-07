package main

import (
	"github.com/gin-gonic/gin"
)

// authenticateMiddleware checks if the session presists for the user and sets key for authentication.
func (app *application) authenticateMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		//Get the session ID from cookie.
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Set(isAuthenticatedContextKey, false)
			c.Next()
			return
		}

		//Get user ID related to the session from Redis.
		var sessionData map[string]any
		jsonData, _ := app.redisClient.Get(c, sessionID).Result()
		deserialize(jsonData, &sessionData)
		idFloat, _ := sessionData["userID"].(float64)
		id := int(idFloat)

		if id == 0 {
			c.Set(isAuthenticatedContextKey, false)
			c.Next()
			return
		}

		//Check if the user with gathered ID persists in database.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(c, err)
			return
		}

		//Set the value for the authenticated key.
		if exists {
			c.Set(isAuthenticatedContextKey, true)
		} else {
			c.Set(isAuthenticatedContextKey, false)
		}
		c.Next()
	}
}
