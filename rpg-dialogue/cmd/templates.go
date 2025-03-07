package main

import (
	"dialogue/internal/models"
	"html/template"
	"path/filepath"
	"time"
)

// Different data that could be passed into templates.
type data struct {
	//Data that gathered from user's forms.
	StoryForm     StoryForm
	UserForm      UserForm
	UserLoginForm UserLoginForm
	PasswordForm  accountPasswordUpdateForm

	//Data that gathered from the databases.
	DataDialogues models.DialoguesData
	UserData      *models.User

	//Data that could be extracted from the context via helper function "newTemplateData".
	CurrentYear     int
	Flash           string
	IsAuthenticated bool
	UserID          int
}

// newTemplateCache parses existing pages once and stores them.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	patterns := []string{
		"./ui/html/pages/firstBlock/*.html",
		"./ui/html/pages/block/*.html",
		"./ui/html/pages/users/*.html",
		"./ui/html/pages/*.html",
	}
	files, err := globbing(patterns)
	if err != nil {
		panic(err)
	}
	for _, page := range files {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}
	return cache, nil
}

func humanTime(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanTime": humanTime,
}
