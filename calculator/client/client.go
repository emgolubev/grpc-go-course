package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/emgolubev/grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello, I am a client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("Couldn't connect: %v", err)
	}

	defer cc.Close()
	c := calculatorpb.NewCalculatorServiceClient(cc)

	// doPND(c)

	// doFindMax(c)

	doErrorUnary(c)

}

func doComputeAverage(c calculatorpb.CalculatorServiceClient) {
	stream, _ := c.ComputeAverage(context.Background())

	for _, num := range os.Args[1:] {
		i, _ := strconv.Atoi(num)
		req := &calculatorpb.AverageRequest{
			Number: int32(i),
		}
		stream.Send(req)
	}

	res, _ := stream.CloseAndRecv()

	fmt.Printf("Success compute average request: %v", res.GetResult())
}

func doFindMax(c calculatorpb.CalculatorServiceClient) {
	stream, _ := c.FindMax(context.Background())

	waitc := make(chan struct{})

	go func() {
		for _, num := range os.Args[1:] {
			i, _ := strconv.Atoi(num)
			req := &calculatorpb.OneIntRequest{
				Number: int32(i),
			}
			stream.Send(req)
			fmt.Printf("Sent to server: %v\n", req)
			time.Sleep(1000 * time.Millisecond)
		}

		stream.CloseSend()
	}()

	go func() {
		for {
			req, err := stream.Recv()

			if err == io.EOF {
				break
			}

			if err != nil {
				break
			}

			fmt.Printf("The current max: %d\n", req.GetNumber())
		}

		close(waitc)
	}()

	<-waitc
}

func doPND(c calculatorpb.CalculatorServiceClient) {
	number, _ := strconv.Atoi(os.Args[1])
	req := &calculatorpb.PNDRequest{
		Number: int32(number),
	}

	stream, err := c.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("Failed PND request: %v", err)
	}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Failed PND msg request: %v", err)
		}

		fmt.Printf("Success PND request: %v\n", msg.Result)
	}
}

func doSum(c calculatorpb.CalculatorServiceClient) {
	numbers := []int32{}

	for _, num := range os.Args[1:] {
		i, _ := strconv.Atoi(num)
		numbers = append(numbers, int32(i))
	}

	req := &calculatorpb.SumRequest{
		Numbers: &calculatorpb.NumbersList{
			List: numbers,
		},
	}

	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("Failed request: %v", err)
	}

	fmt.Printf("Sum of %v is %d", numbers, res.Result)
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	// correct call
	doErrorCall(c, 10)

	// error call
	doErrorCall(c, -32)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) error {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{
		Number: n,
	})

	if err != nil {
		s, ok := status.FromError(err)

		if ok {
			// actual error from gRPC
			fmt.Println(s.Message())
			fmt.Println(s.Code())
			if s.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
			}
		} else {
			log.Fatalf("Big error while SquareRoot: %v", s)
		}

		return err
	}

	fmt.Printf("square root from 10 = %f\n", res.GetNumberRoot())

	return nil
}
