package controllers

import (
	"strings"

	"github.com/lujiacn/revauth/app/revauth"

	"github.com/lujiacn/revauth/app/models"

	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	mgodo "gopkg.in/lujiacn/mgodo.v0"
)

type Auth struct {
	*revel.Controller
	mgodo.MgoController
}

//Authenticate for LDAP authenticate
func (c Auth) Authenticate(account, password string) revel.Result {
	if account == "" || password == "" {
		c.Flash.Error("Please fill in account and password")
		return c.Redirect("/login")
	}
	authUser := revauth.Authenticate(account, password)
	if !authUser.IsAuthenticated {
		c.Flash.Error("Authenticate failed: %v", authUser.Error)
		return c.Redirect("/login")
	}

	c.Session["Identity"] = strings.ToLower(account)

	//save current user information
	currentUser := new(models.User)
	currentUser.Identity = strings.ToLower(account)
	currentUser.Mail = authUser.Email
	currentUser.Avatar = authUser.Avatar
	currentUser.Name = authUser.Name
	currentUser.Depart = authUser.Depart

	// cache user info
	go cache.Set(c.Session.ID(), currentUser, cache.DefaultExpiryTime)

	go func(user *models.User) {
		// save to local user
		s := mgodo.NewMgoSession()
		defer s.Close()
		err := user.SaveUser(s)
		if err != nil {
			revel.AppLog.Errorf("Save user error: %v", err)
		}

	}(currentUser)

	c.Flash.Success("Welcome, %v", currentUser.Name)
	return c.Redirect("/")
}

//Logout
func (c Auth) Logout() revel.Result {
	//delete cache which is logged in user info
	cache.Delete(c.Session.ID())

	c.Session = make(map[string]string)
	c.Flash.Success("You have logged out.")
	return c.Redirect("/")
}
