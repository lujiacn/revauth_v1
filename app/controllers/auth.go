package controllers

import (
	"path"
	"strings"

	"github.com/lujiacn/revauth/app/models"

	auth "github.com/lujiacn/revauth/auth"

	"github.com/revel/revel"
	"github.com/revel/revel/cache"
	mgodb "gopkg.in/lujiacn/mgodb.v1"
)

type Auth struct {
	*revel.Controller
	mgodb.MgoController
}

//Authenticate for LDAP authenticate
func (c Auth) Authenticate(account, password string) revel.Result {
	if account == "" || password == "" {
		c.Flash.Error("Please fill in account and password")
		return c.Redirect("/login")
	}
	authUser := auth.Authenticate(account, password)
	if !authUser.IsAuthenticated {
		c.Flash.Error("Authenticate failed: %v", authUser.Error)
		return c.Redirect("/login")
	}

	c.Session["Identity"] = strings.ToLower(account)
	c.Session["UserName"] = authUser.Name

	//save user information
	user := &models.User{}
	user.Identity = strings.ToLower(account)
	user.Mail = authUser.Email
	user.Avatar = authUser.Avatar
	user.Name = authUser.Name
	user.Depart = authUser.Depart

	err := user.SaveUser(c.MgoSession)
	if err != nil {
		c.Flash.Error("Error during save user", err)
		return c.Redirect("/login")
	}

	c.Flash.Success("Welcome, %v", user.Name)
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

//Login
func (c Auth) Login() revel.Result {
	AppName := strings.ToUpper(revel.AppName)
	loginTpl := path.Join(revel.AppPath, "app", "views", "login.html")
	c.ViewArgs["AppName"] = AppName
	return c.RenderTemplate(loginTpl)
}
