package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/schema"
	"github.com/redis/go-redis/v9"
)

// serverError handles some unexpected errors.
func (app *application) serverError(c *gin.Context, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.errorLog.Output(2, trace)
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error":   http.StatusText(http.StatusInternalServerError),
		"message": "An unexpected error occurred.",
	})
}

// newTemplateData gathers context data and passes it to every request by default.
func (app *application) newTemplateData(c *gin.Context) *data {
	return &data{
		CurrentYear:     time.Now().Year(),
		Flash:           app.getFlash(c),
		IsAuthenticated: app.isAuthenticated(c),
		UserID:          app.getID(c),
	}
}

// isAuthenticated checks if user is authenticaded or not.
func (app *application) isAuthenticated(c *gin.Context) bool {
	getValue, ok := c.Get(isAuthenticatedContextKey)
	if !ok {
		return false
	}
	isAuthenticated := getValue.(bool)
	return isAuthenticated
}

// setFlash sets a flash message by putting it into the cookie.
func (app *application) setFlash(c *gin.Context, flashText string) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.SetCookie("flash_message", flashText, 5, "/", "", false, true)
		return
	}
	app.redisClient.Set(c, "flash:"+sessionID, flashText, 5*time.Minute)
}

// getFlash extracts flash from the context. It checks if the flash is in temporary flash_message cookie or inside created session.
// It checks every request and passes data.
func (app *application) getFlash(c *gin.Context) string {
	var flashTextTmp string

	//Checking if flash is inside temporary cookie, if so extract it.
	flashTextTmp, err := c.Cookie("flash_message")
	if err == nil {
		c.SetCookie("flash_message", "", -1, "/", "", false, true)
		return flashTextTmp
	}
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		flashTextTmp, _ = c.Cookie("flash_message")
		c.SetCookie("flash_message", "", -1, "/", "", false, true)
		return flashTextTmp
	}

	//Trying to gather flash from session.
	flashSession, err := app.redisClient.Get(c, "flash:"+sessionID).Result()
	if err == redis.Nil {
		flashSession = ""
	} else if err != nil {
		log.Println("Redis error:", err)
		flashSession = ""
	} else {
		app.redisClient.Del(c, "flash:"+sessionID)
	}
	return flashSession
}

// render renders a page from created HTML pages cash.
func (app *application) render(c *gin.Context, status int, page string, data interface{}) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.errorLog.Println(err)
		return
	}
	c.Status(status)
	err := ts.ExecuteTemplate(c.Writer, "base", data)
	if err != nil {
		app.errorLog.Println(err)
		return
	}
}

// globbing is a wrap around glob() function to check for different patterns.
func globbing(patterns []string) ([]string, error) {
	var result []string
	for _, v := range patterns {
		pattern, err := filepath.Glob(v)
		if err != nil {
			return nil, err
		}
		result = append(result, pattern...)
	}
	return result, nil
}

// generateSessionID generated a string wich becomes session ID.
func generateSessionID() string {
	return uuid.New().String()
}

// serialize is a helper to marshal data and return it as a string.
func serialize(data map[string]any) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// deserialize is a function that unmarshals json data into string.
func deserialize(jsonStr string, target *map[string]any) {
	json.Unmarshal([]byte(jsonStr), target)
}

// getID gets user ID from current session.
func (app *application) getID(c *gin.Context) int {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return 0 //If session does not exists returns nil.
	}

	// Get ID of a user directly from Redis.
	userIDStr, err := app.redisClient.Get(c, sessionID+":userID").Result()
	if err == redis.Nil {
		return 0
	} else if err != nil {
		return 0
	}

	// Convert string, wich is ID, to int.
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0
	}
	return userID
}

// parse is a helper function to parse forms from the user.
func (app *application) parse(c *gin.Context, form any) {
	if err := c.Request.ParseForm(); err != nil {
		app.errorLog.Print(err.Error())
		app.serverError(c, err)
	}
	dec := schema.NewDecoder()
	if err := dec.Decode(form, c.Request.PostForm); err != nil {
		app.errorLog.Print(err.Error())
		app.serverError(c, err)
	}
}
