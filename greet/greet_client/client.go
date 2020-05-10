package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emgolubev/grpc-go-course/greet/greetpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello, I am a client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Couldn't connect: %v", err)
	}

	defer cc.Close()
	c := greetpb.NewGreetServiceClient(cc)

	// doUnary(c)

	// doServerStreaming(c)

	// doClientStreaming(c)

	doBiDiStreaming(c)

}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Eugene",
			LastName:  "Golubev",
		},
	}

	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("Failed greet request: %v", err)
	}

	fmt.Printf("Success request: %v\n", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Eugene",
			LastName:  "Golubev",
		},
	}

	stream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("Failed greet many times request: %v", err)
	}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Failed greet many timed msg request: %v", err)
		}

		fmt.Printf("Success greet many times request: %v\n", msg.Result)
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	stream, _ := c.LongGreet(context.Background())

	requests := []*greetpb.GreetRequest{
		&greetpb.GreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Eugene",
				LastName:  "Golubev",
			},
		},
		&greetpb.GreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Oxana",
				LastName:  "Golubeva",
			},
		},
	}

	for _, r := range requests {
		stream.Send(r)
	}

	res, _ := stream.CloseAndRecv()

	fmt.Printf("Success client stream request: %v", res.GetResult())

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	stream, _ := c.GreetEveryone(context.Background())

	waitc := make(chan struct{})

	requests := []*greetpb.GreetRequest{
		&greetpb.GreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Eugene",
				LastName:  "Golubev",
			},
		},
		&greetpb.GreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Oxana",
				LastName:  "Golubeva",
			},
		},
	}

	go func() {
		for _, req := range requests {
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}

		stream.CloseSend()
	}()

	go func() {
		for {
			req, err := stream.Recv()

			if err == io.EOF {
				fmt.Println("Close client stream")
				break
			}

			if err != nil {
				break
			}

			fmt.Printf("Success greet everyone: %v\n", req.GetResult())
		}

		close(waitc)
	}()

	<-waitc
}
