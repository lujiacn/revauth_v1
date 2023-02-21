package controllers

import (
	"strings"

	mgodo "github.com/lujiacn/mgodo"
	revauth "github.com/lujiacn/revauth_v1"
	"github.com/lujiacn/revauth_v1/app/models"
	"github.com/revel/revel"
	"github.com/revel/revel/cache"
)

type Auth struct {
	*revel.Controller
	mgodo.MgoController
}

// Authenticate for LDAP authenticate
func (c *Auth) Authenticate(account, password string) revel.Result {
	//get nextUrl
	nextUrl := c.Params.Get("nextUrl")
	if nextUrl == "" {
		nextUrl = "/"
	}

	if account == "" || password == "" {
		c.Flash.Error("Please fill in account and password")
		return c.Redirect(c.Request.Referer())
	}
	// account can be ID or mail
	authUser := revauth.Authenticate(account, password)
	authIdentity := strings.ToLower(authUser.Account) // ID returned from grpcldap service
	if !authUser.IsAuthenticated {
		//Save LoginLog
		loginLog := new(models.LoginLog)
		loginLog.Account = authIdentity
		loginLog.Status = "FAILURE"
		loginLog.IPAddress = c.Request.RemoteAddr
		mgodo.New(c.MgoSession, loginLog).Create()

		c.Flash.Error("Authenticate failed: %v", authUser.Error)
		return c.Redirect(c.Request.Referer())
	}

	// save login log
	loginLog := new(models.LoginLog)
	loginLog.Account = authIdentity
	loginLog.Status = "SUCCESS"
	loginLog.IPAddress = c.Request.RemoteAddr
	mgodo.New(c.MgoSession, loginLog).Create()

	c.Session["Identity"] = authIdentity

	//save current user information
	currentUser := new(models.User)
	currentUser.Identity = authIdentity
	currentUser.Mail = authUser.Email
	currentUser.Avatar = authUser.Avatar
	currentUser.Name = authUser.Name
	currentUser.Depart = authUser.Depart
	currentUser.First = authUser.First
	currentUser.Last = authUser.Last

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
	return c.Redirect(nextUrl)
}

// Logout
func (c *Auth) Logout() revel.Result {
	//delete cache which is logged in user info
	cache.Delete(c.Session.ID())

	c.Session = make(map[string]interface{})
	c.Flash.Success("You have logged out.")
	return c.Redirect("/")
}
