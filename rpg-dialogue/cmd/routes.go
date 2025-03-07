package main

import (
	"github.com/gin-gonic/gin"
)

func (app *application) routes() *gin.Engine {
	router := gin.Default()

	router.Use(app.authenticateMiddleware())

	router.Static("/static", "./ui/static")

	router.GET("/", app.redirectHomePage)
	router.GET("/home", app.homePage)
	router.GET("/about", app.about)

	router.GET("/newfirstblock", app.emptyFBView)
	router.POST("/newfirstblock", app.createFB)
	router.GET("/firstblock", app.createdFBView)
	router.POST("/firstblock", app.deleteFB)
	router.GET("/editfirstblock", app.editFBView)
	router.POST("/editfirstblock", app.editFB)

	router.GET("/block", app.createdBView)
	router.POST("/block", app.deleteB)
	router.GET("/editblock", app.editBView)
	router.POST("/editblock", app.editB)
	router.GET("/{digits:[0-9]+}", app.redirectBlock)

	router.GET("/user/signup", app.userSignupView)
	router.POST("/user/signup", app.userSignup)

	router.GET("/user/login", app.userLoginView)
	router.POST("/user/login", app.userLogin)
	router.POST("/user/logout", app.userLogout)

	router.GET("/account/view", app.accountView)

	router.GET("/account/password/update", app.passwordUpdateView)
	router.POST("/account/password/update", app.passwordUpdate)

	return router
}
