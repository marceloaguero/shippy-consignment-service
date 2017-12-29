package main

import (
	"log"
	"net"
	"os"

	pb "github.com/marceloaguero/shippy-consignment-service/proto"
	vesselProto "github.com/marceloaguero/shippy-vessel-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultHost          = "db:27017"
	port                 = ":50051"
	addressVesselService = "vessel-service:50052"
)

func main() {

	// Database host from environment
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = defaultHost
	}

	session, err := CreateSession(host)
	defer session.Close()
	if err != nil {
		log.Panicf("Could not connect to datastore with host %s - %v", host, err)
	}

	conn, err := grpc.Dial(addressVesselService, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect to vessel service: %v", err)
	}
	defer conn.Close()

	vesselClient := vesselProto.NewVesselServiceClient(conn)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	// pb.RegisterShippingServiceServer(s, &service{session, vesselClient})
	pb.RegisterShippingServiceServer(s, &service{session, vesselClient})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
