package auth

import (
	"context"
	"fmt"

	gAuth "grpcLdap/auth"

	"github.com/revel/revel"
	"google.golang.org/grpc"
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
