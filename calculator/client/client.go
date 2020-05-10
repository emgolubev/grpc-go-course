package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/emgolubev/grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
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

	doComputeAverage(c)

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
