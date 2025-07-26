package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Shuba-Buba/usdt-rate-service/proto/rate/v1"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewRateServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := client.GetRates(ctx, &pb.GetRatesRequest{})
	if err != nil {
		log.Fatalf("Failed to get rates: %v", err)
	}

	timestamp := time.Unix(response.Timestamp, 0)

	fmt.Printf("USDT Exchange Rates:\n")
	fmt.Println("Ask: ", response.Asks)
	fmt.Println("Bid: ", response.Bids)
	fmt.Printf("Timestamp: %s\n", timestamp.Format(time.RFC3339))

	responseHelathCheck, err := client.HealthCheck(ctx, &pb.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("Failed to get health check: %v", err)
	}

	fmt.Printf("Health check: %s\n", responseHelathCheck.Status)
}
