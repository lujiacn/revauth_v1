package revauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/lujiacn/revauth/app/models"
	gAuth "github.com/lujiacn/revauth/auth"
	"google.golang.org/grpc"
	"gopkg.in/lujiacn/mgodo.v0"

	"github.com/revel/revel"
)

var (
	grpcDial string
)

//Init reading LDAP configuration
func Init() {
	grpcAuthServer, ok := revel.Config.String("grpcauth.server")
	if !ok {
		panic("Authenticate server not defined")

	}
	grpcAuthPort := revel.Config.StringDefault("grpcauth.port", "50051")
	grpcDial = grpcAuthServer + ":" + grpcAuthPort
}

//Authenticate do auth and return Auth object including user information and lognin success or not
func Authenticate(account, password string) *gAuth.AuthReply {
	conn, err := grpc.Dial(grpcDial, grpc.WithInsecure())
	if err != nil {
		return &gAuth.AuthReply{Error: fmt.Sprintf("Connect auth server failed, %v", err)}
	}
	defer conn.Close()
	c := gAuth.NewAuthClient(conn)
	r, err := c.Authenticate(context.Background(), &gAuth.AuthRequest{Account: account, Password: password})
	if err != nil {
		return &gAuth.AuthReply{Error: fmt.Sprintf("Authenticate failed due to %v ", err)}
	}
	return r
}

func Query(account string) *gAuth.QueryReply {
	conn, err := grpc.Dial(grpcDial, grpc.WithInsecure())
	if err != nil {
		fmt.Println("connect error", err)
		return &gAuth.QueryReply{Error: fmt.Sprintf("Connect auth server failed, %v", err)}
	}
	defer conn.Close()
	c := gAuth.NewAuthClient(conn)
	r, err := c.Query(context.Background(), &gAuth.QueryRequest{Account: account})
	if err != nil {
		fmt.Println("Query user error", err)
		return &gAuth.QueryReply{Error: fmt.Sprintf("User not found: %v ", err)}
	}
	fmt.Println("r", r)
	return r

}

func QueryAndSave(account string) (*models.User, error) {
	authUser := Query(account)
	if authUser.Error != "" {
		return nil, fmt.Errorf(authUser.Error)
	}
	if authUser.NotExist {
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
