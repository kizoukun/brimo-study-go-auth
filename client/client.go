package main

import (
	"context"
	"go-auth/server/pb"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Connect to the server
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewUserServiceClient(conn)

	// Call GetUser
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.RegisterUser(ctx, &pb.RegisterUserRequest{Name: "name5", Password: "password5"})
	// res, err := client.LoginUser(ctx, &pb.LoginUserRequest{Name: "name4", Password: "password4"})
	if err != nil {
		log.Fatalf("could not get user: %v", err)
	}
	log.Printf("User: %v", res)
}