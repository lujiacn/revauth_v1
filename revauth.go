package revauth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/lujiacn/mgodo"
	"github.com/lujiacn/revauth/app/models"
	gAuth "github.com/lujiacn/revauth/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/revel/revel"
)

var (
	grpcDial        string
	grpcAuthConnect string
	conn            *grpc.ClientConn // keep connection
)

//Init reading LDAP configuration
func Init() {
	// update grpcauth server and port to grpc://connection_string
	var found bool

	grpcAuthConnect, found = revel.Config.String("grpcauth.connect")
	if !found {
		panic("Authenticate connection not defined")
	}
	connect()
}

func connect() {
	// parse connection scheme
	h, err := url.Parse(grpcAuthConnect)
	if err != nil {
		panic("Invalid connection format. eg: grpc://host:port/path")
	}

	if h.Scheme == "grpc" || h.Scheme == "" {
		conn, err = grpc.Dial(path.Join(h.Host, h.Path), grpc.WithInsecure())
		if err != nil {
			revel.AppLog.Critf("%v", err)
		}
	}

	if h.Scheme == "grpcs" {
		config := &tls.Config{
			InsecureSkipVerify: true,
		}
		conn, err = grpc.Dial(path.Join(h.Host, h.Path), grpc.WithTransportCredentials(credentials.NewTLS(config)))
		if err != nil {
			revel.AppLog.Critf("%v", err)
		}
	}
}

//Authenticate do auth and return Auth object including user information and lognin success or not
func Authenticate(account, password string) *gAuth.AuthReply {
	if conn == nil {
		connect()
	}

	c := gAuth.NewAuthClient(conn)
	r, err := c.Authenticate(context.Background(), &gAuth.AuthRequest{Account: account, Password: password})
	if err != nil {
		return &gAuth.AuthReply{Error: fmt.Sprintf("Authenticate failed due to %v ", err)}
	}
	return r
}

func Query(account string) *gAuth.QueryReply {
	if conn == nil {
		connect()
	}

	c := gAuth.NewAuthClient(conn)
	r, err := c.Query(context.Background(), &gAuth.QueryRequest{Account: account})
	if err != nil {
		return &gAuth.QueryReply{Error: fmt.Sprintf("User not found: %v ", err)}
	}
	return r

}

func QueryMail(email string) *gAuth.QueryReply {

	if conn == nil {
		connect()
	}

	c := gAuth.NewAuthClient(conn)
	r, err := c.Query(context.Background(), &gAuth.QueryRequest{Email: email})
	if err != nil {
		return &gAuth.QueryReply{Error: fmt.Sprintf("User not found: %v ", err)}
	}
	return r

}

func QueryMailAndSave(email string) (*models.User, error) {
	authUser := QueryMail(email)

	if authUser.Error != "" && authUser.Error != "<nil>" {
		fmt.Println("Errors", authUser.Error)
		return nil, fmt.Errorf(authUser.Error)
	}
	if authUser.NotExist {
		fmt.Println("Not exist", authUser.Error)
		return nil, fmt.Errorf("User not exist")
	}

	user := new(models.User)
	user.Identity = strings.ToLower(authUser.Account)
	user.Mail = authUser.Email
	user.Avatar = authUser.Avatar
	user.Name = authUser.Name
	user.Depart = authUser.Depart
	s := mgodo.NewMgoSession()
	defer s.Close()
	user.SaveUser(s)
	return user, nil
}

func QueryAndSave(account string) (*models.User, error) {
	authUser := Query(account)

	if authUser.Error != "" && authUser.Error != "<nil>" {
		fmt.Println("Errors", authUser.Error)
		return nil, fmt.Errorf(authUser.Error)
	}
	if authUser.NotExist {
		fmt.Println("Not exist", authUser.Error)
		return nil, fmt.Errorf("User not exist")
	}

	user := new(models.User)
	user.Identity = strings.ToLower(account)
	user.Mail = authUser.Email
	user.Avatar = authUser.Avatar
	user.Name = authUser.Name
	user.Depart = authUser.Depart
	s := mgodo.NewMgoSession()
	defer s.Close()
	user.SaveUser(s)
	return user, nil
}
