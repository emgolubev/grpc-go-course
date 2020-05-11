package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc/credentials"

	"google.golang.org/grpc/codes"

	"github.com/emgolubev/grpc-go-course/greet/greetpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet method was invoked with %v", req)

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	response := &greetpb.GreetResponse{Result: result}

	return response, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	firstName := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number " + strconv.Itoa(i+1)
		response := &greetpb.GreetResponse{Result: result}

		stream.Send(response)

		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	result := ""
	for {
		msg, err := stream.Recv()

		fmt.Printf("Recieve client stream: %v", msg)

		if err == io.EOF {
			return stream.SendAndClose(&greetpb.GreetResponse{
				Result: result,
			})
		}

		if err != nil {
			log.Fatalf("Failed client stream request: %v", err)
		}

		result += msg.GetGreeting().GetFirstName() + "! "
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			log.Fatalf("Greet Everyone failed request: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName

		sErr := stream.Send(&greetpb.GreetResponse{Result: result})

		if sErr != nil {
			log.Fatalf("Error while sending to client: %v", err)
			return sErr
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet method was invoked with %v", req)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("client canceled the request")
			return nil, status.Error(codes.Canceled, "Client canceled the request")
		}
		time.Sleep(time.Second)
	}

	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	response := &greetpb.GreetResponse{Result: result}

	return response, nil
}

func main() {
	fmt.Println("Hello World!")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)

	}

	opts := []grpc.ServerOption{}
	tls := true

	if tls {
		certFile := "../../ssl/server.crt"
		keyFile := "../../ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)

		if sslErr != nil {
			log.Fatalf("failed loading certificates: %v", sslErr)
			return
		}

		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)

	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve, %v", err)
	}
}
