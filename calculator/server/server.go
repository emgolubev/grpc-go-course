package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sort"

	"github.com/emgolubev/grpc-go-course/calculator/calculatorpb"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	var result int32 = 0

	for _, num := range req.GetNumbers().GetList() {
		result += num
	}

	response := &calculatorpb.SumResponse{Result: result}

	return response, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PNDRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {

	n := req.GetNumber()
	k := int32(2)

	for n > 1 {
		if n%k == 0 {
			res := &calculatorpb.PNDResponse{Result: k}
			stream.Send(res)
			n = n / k
		} else {
			k = k + 1
		}

	}

	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	sum := int32(0)
	cnt := 0

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Result: float32(sum) / float32(cnt),
			})
		}

		// TODO: check errors

		sum += msg.GetNumber()
		cnt++

		fmt.Println(sum)
	}
}

func (*server) FindMax(stream calculatorpb.CalculatorService_FindMaxServer) error {
	numbers := []int{}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		numbers = append(numbers, int(msg.GetNumber()))

		sort.Ints(numbers)

		stream.Send(&calculatorpb.OneIntResponse{
			Number: int32(numbers[len(numbers)-1]),
		})

	}
}

func main() {
	fmt.Println("Hello Calculator!")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)

	}

	s := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve, %v", err)
	}
}
