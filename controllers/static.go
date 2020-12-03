package controllers

import (
	"revision/lenslocked.com/views"
)

// Static returns a static instance
type Static struct {
	Home    *views.View
	Contact *views.View
}

// NewStatic returns a new instance of Static
func NewStatic() *Static {
	return &Static{
		Home:    views.NewView("bootstrap", "static/index"),
		Contact: views.NewView("bootstrap", "static/contact"),
	}
}
