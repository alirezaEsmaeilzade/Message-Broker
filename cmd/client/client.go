package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	pb "therealbroker/api/proto"
	"time"
)

const address = "localhost:5050"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did no connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewBrokerClient(conn)
	ct, _ := context.WithTimeout(context.Background(), 10*time.Second)
	sub, _ := c.Subscribe(ct, &pb.SubscribeRequest{Subject: "ali"})
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, _ = c.Publish(ctx, &pb.PublishRequest{
		Subject:           "ali",
		Body:              []byte("Hello World!"),
		ExpirationSeconds: 5,
	})
	msg, err := sub.Recv()
	fmt.Println(msg, err)
}