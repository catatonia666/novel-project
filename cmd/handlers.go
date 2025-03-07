package main

import (
	"dialogue/internal/models"
	"dialogue/internal/validator"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type StoryForm struct {
	Title   string `schema:"title"`
	Content string `schema:"content"`
	Options string `schema:"options"`
	Privacy bool   `schema:"privacy"`
	validator.Validator
}

type UserForm struct {
	Nickname string `schema:"nickname"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
	validator.Validator
}

type UserLoginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
	validator.Validator
}

type accountPasswordUpdateForm struct {
	CurrentPassword         string `schema:"currentPassword"`
	NewPassword             string `schema:"newPassword"`
	NewPasswordConfirmation string `schema:"newPasswordConfirmation"`
	validator.Validator
}

// redirectHome redirects default query to the home page.
func (app *application) redirectHomePage(c *gin.Context) {
	if c.Request.URL.Path != "/" {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Redirect(http.StatusFound, "/home")
}

// homePage renders home page with public and private stories related to a user.
func (app *application) homePage(c *gin.Context) {
	//Get user ID from context.
	userID := app.getID(c)

	//Pass related data and stories and render the page.
	data := app.newTemplateData(c)
	data.DataDialogues.DialoguesToDisplay = app.dialogues.Latest(userID)
	app.render(c, http.StatusOK, "home.html", data)
}

// emptyFBView renders an empty form where user can create a starting point of a new story.
func (app *application) emptyFBView(c *gin.Context) {
	data := app.newTemplateData(c)
	app.render(c, http.StatusOK, "createFB.html", data)
}

// createFB method parse form, get the values from it and create first block of the story.
// The first block has it's unique ID, and it is equal to the ID of the story itself.
func (app *application) createFB(c *gin.Context) {

	//Get values from the form and store them into form variable.
	var storyForm StoryForm
	app.parse(c, &storyForm)

	//Basic validations checks.
	storyForm.CheckField(validator.NotBlank(storyForm.Title), "title", "This field cannot be blank")
	storyForm.CheckField(validator.NotBlank(storyForm.Content), "content", "This field cannot be blank")
	if !storyForm.Valid() {
		data := app.newTemplateData(c)
		data.StoryForm = storyForm
		app.render(c, http.StatusUnprocessableEntity, "createFB.html", data)
		return
	}

	//Parse options from the form and store them into the slice of strings.
	optionsSlice := strings.Split(storyForm.Options, "\r\n")

	//Get user ID from context and put gathered data into DB, then get the ID of fresh created first block of the story.
	userID := app.getID(c)
	newStoryID := app.dialogues.CreateFB(userID, storyForm.Title, storyForm.Content, optionsSlice, storyForm.Privacy)

	app.setFlash(c, "First step is done, and the story have been created!")
	path := "firstblock?id=" + strconv.Itoa(int(newStoryID))
	c.Redirect(http.StatusFound, path)
}

// createdFBView renders view of fresh created story with nessessary data.
func (app *application) createdFBView(c *gin.Context) {

	//Get the ID of fresh story.
	storyID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		return
	}

	//Get the data related to the story with ID and pass it to the view.
	data := app.newTemplateData(c)
	data.DataDialogues = app.dialogues.CreatedFBView(storyID)
	app.render(c, http.StatusOK, "renderFB.html", data)
}

// editFBView allows to edit first block of the story.
func (app *application) editFBView(c *gin.Context) {

	//get the ID of the story and data of the first block.
	storyID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}

	// Render the form for editing with existing data.
	data := app.newTemplateData(c)
	data.DataDialogues = app.dialogues.CreatedFBView(storyID)
	app.render(c, http.StatusOK, "editFB.html", data)
}

// editFB passes edited data to the data base.
func (app *application) editFB(c *gin.Context) {

	//Parse edited data for the first block of a story.
	var storyForm StoryForm
	app.parse(c, &storyForm)
	optionsSlice := strings.Split(storyForm.Options, "\r\n")

	//Get ID of the story and update it's data with a new one.
	storyID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}

	userID := app.getID(c)
	app.dialogues.EditFB(storyID, userID, storyForm.Title, storyForm.Content, optionsSlice)
	path := "firstblock?id=" + strconv.Itoa(storyID)
	c.Redirect(http.StatusFound, path)
}

// deleteFB deletes the whole story and all blocks related to it.
func (app *application) deleteFB(c *gin.Context) {
	id, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}
	app.dialogues.DeleteFB(id)
	c.Redirect(http.StatusFound, "/home")
}

// createdBView renders existing block of a story.
func (app *application) createdBView(c *gin.Context) {

	//Get ID of a block.
	blockID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}

	//Retrieve data from database and render the block.
	data := app.newTemplateData(c)
	data.DataDialogues = app.dialogues.EditBView(blockID)
	app.render(c, http.StatusOK, "renderB.html", data)
}

// editBView allows to edit block of the story.
func (app *application) editBView(c *gin.Context) {

	//Get ID of a block.
	blockID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}

	// Render the form for editing with existing data.
	data := app.newTemplateData(c)
	data.DataDialogues = app.dialogues.EditBView(blockID)
	app.render(c, http.StatusOK, "editB.html", data)
}

// editFB passes edited data to the data base.
func (app *application) editB(c *gin.Context) {

	//Parse form and store it.
	var blockForm StoryForm
	app.parse(c, &blockForm)

	//Parse options from the form and store them into the slice of strings.
	optionsSlice := strings.Split(blockForm.Options, "\r\n")

	//Get ID of the editing block and update it's data with a new one.
	blockID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}

	userID := app.getID(c)
	app.dialogues.EditB(blockID, userID, blockForm.Title, blockForm.Content, optionsSlice)
	path := "block?id=" + strconv.Itoa(blockID)
	c.Redirect(http.StatusFound, path)
}

// redirectBlock redirects user to a block.
func (app *application) redirectBlock(c *gin.Context) {
	path := "block?id=" + strings.ReplaceAll(c.Request.URL.Path, "/", "")
	c.Redirect(http.StatusFound, path)
}

// deleteB deletes a block and other blocks if they are not related to other blocks.
func (app *application) deleteB(c *gin.Context) {
	blockID, err := strconv.Atoi(c.Request.URL.Query().Get("id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
	}
	app.dialogues.DeleteB(blockID)
	c.Redirect(http.StatusFound, "/home")
}

// userSignupView renders the page for signing user up.
func (app *application) userSignupView(c *gin.Context) {
	data := app.newTemplateData(c)
	app.render(c, http.StatusOK, "signup.html", data)
}

// userSignup signs up a new user with provided data.
func (app *application) userSignup(c *gin.Context) {

	//Parse provided form.
	var userForm UserForm
	app.parse(c, &userForm)

	//Basic validation checks.
	userForm.CheckField(validator.NotBlank(userForm.Nickname), "nickname", "This field cannot be blank")
	userForm.CheckField(validator.NotBlank(userForm.Email), "email", "This field cannot be blank")
	userForm.CheckField(validator.Matches(userForm.Email, validator.EmailRX), "email", "This field must be a valid email address")
	userForm.CheckField(validator.NotBlank(userForm.Password), "password", "This field cannot be blank")
	userForm.CheckField(validator.MinChars(userForm.Password, 8), "password", "This field must be at least 8 characters long")
	if !userForm.Valid() {
		data := app.newTemplateData(c)
		data.UserForm = userForm
		app.render(c, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	//Save new user into the data base.
	err := app.users.Insert(userForm.Nickname, userForm.Email, userForm.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			userForm.AddFieldError("email", "Email address is already in use")
			data := app.newTemplateData(c)
			data.UserForm = userForm
			app.render(c, http.StatusUnprocessableEntity, "signup.html", data)
		} else {
			app.serverError(c, err)
		}
		return
	}

	app.setFlash(c, "You successfully signed up. Please, log in for more content.")
	c.Redirect(http.StatusFound, "/user/login")
}

// userLoginView renders the page for logging user in.
func (app *application) userLoginView(c *gin.Context) {
	data := app.newTemplateData(c)
	app.render(c, http.StatusOK, "login.html", data)
}

// userLogin logs user in.
func (app *application) userLogin(c *gin.Context) {
	app.setFlash(c, "You logged in with a geat success.")

	//Parse provided form.
	var userForm UserLoginForm
	app.parse(c, &userForm)

	//Basic validations check.
	userForm.CheckField(validator.NotBlank(userForm.Email), "email", "This field cannot be blank")
	userForm.CheckField(validator.Matches(userForm.Email, validator.EmailRX), "email", "This field must be a valid email address")
	userForm.CheckField(validator.NotBlank(userForm.Password), "password", "This field cannot be blank")
	if !userForm.Valid() {
		data := app.newTemplateData(c)
		data.UserLoginForm = userForm
		app.render(c, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	//Authenticate user and log him in if no errors.
	userID, err := app.users.Authenticate(userForm.Email, userForm.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			userForm.AddNonFieldError("Email or password is incorrect")
			data := app.newTemplateData(c)
			data.UserLoginForm = userForm
			app.render(c, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(c, err)
		}
		return
	}

	//Generate new session ID for the user and save related data via cookies.
	sessionID := generateSessionID()
	c.SetCookie("session_id", sessionID, 3600, "/", "", false, true)
	sessionData := map[string]any{
		"userID":       userID,
		"createdAt":    time.Now().Unix(),
		"lastActiveAt": time.Now().Unix(),
	}
	app.redisClient.Set(c, sessionID, serialize(sessionData), 12*time.Hour)
	app.redisClient.Set(c, sessionID+":userID", userID, 12*time.Hour)

	c.Redirect(http.StatusFound, "/home")
}

// userLogoutPost logouts the user and destroy current session.
func (app *application) userLogout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil {
		app.redisClient.Del(c, sessionID)
	}
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.SetCookie("flash_message", "You logged out with a great success", 5, "/", "", false, true)

	c.Redirect(http.StatusFound, "/home")
}

// accountView renders a page with data related to the user (nickname and other).
func (app *application) accountView(c *gin.Context) {

	//Get user ID and then other data related to the user.
	userID := app.getID(c)
	user, err := app.users.GetUser(userID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			c.Redirect(http.StatusFound, "/user/login")
		} else {
			app.serverError(c, err)
		}
		return
	}

	//Renders the page with all related data.
	data := app.newTemplateData(c)
	data.UserData = user
	app.render(c, http.StatusOK, "account.html", data)
}

// passwordUpdateView renders a page where user can change the password for the account.
func (app *application) passwordUpdateView(c *gin.Context) {
	data := app.newTemplateData(c)
	data.PasswordForm = accountPasswordUpdateForm{}
	app.render(c, http.StatusOK, "password.html", data)
}

// passwordUpdate updates user's password with new provided information.
func (app *application) passwordUpdate(c *gin.Context) {

	//Parse the form provided by the user.

	var passwordForm accountPasswordUpdateForm
	app.parse(c, &passwordForm)

	//Basic validations check.
	passwordForm.CheckField(validator.NotBlank(passwordForm.CurrentPassword), "currentPassword", "This field cannot be blank")
	passwordForm.CheckField(validator.NotBlank(passwordForm.NewPassword), "newPassword", "This field cannot be blank")
	passwordForm.CheckField(validator.MinChars(passwordForm.NewPassword, 8), "newPassword", "This field must be at least 8 characters long")
	passwordForm.CheckField(validator.NotBlank(passwordForm.NewPasswordConfirmation), "newPasswordConfirmation", "This field cannot be blank")
	passwordForm.CheckField(passwordForm.NewPassword == passwordForm.NewPasswordConfirmation, "newPasswordConfirmation", "Passwords do not match")
	if !passwordForm.Valid() {
		data := app.newTemplateData(c)
		data.PasswordForm = passwordForm
		app.render(c, http.StatusUnprocessableEntity, "password.html", data)
		return
	}

	//Get ID of a user.
	userID := app.getID(c)

	//Update password with new information.
	err := app.users.PasswordUpdate(userID, passwordForm.CurrentPassword, passwordForm.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			passwordForm.AddFieldError("currentPassword", "Current password is incorrect")
			data := app.newTemplateData(c)
			data.PasswordForm = passwordForm
			app.render(c, http.StatusUnprocessableEntity, "password.tmpl", data)
		} else {
			app.serverError(c, err)
		}
		return
	}
	app.setFlash(c, "Password successfully updated.")
	c.Redirect(http.StatusFound, "/account/view")
}

// about contains basic idea of the site.
func (app *application) about(c *gin.Context) {
	data := app.newTemplateData(c)
	app.render(c, http.StatusOK, "about.html", data)
}
