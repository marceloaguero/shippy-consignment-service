package main

import (
	"context"
	"log"
	"net"

	pb "github.com/marceloaguero/shippy-consignment-service/proto"
	vesselProto "github.com/marceloaguero/shippy-vessel-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port          = ":50051"
	vesselService = "vessel-service:50052"
)

type Repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

type ConsignmentRepository struct {
	consignments []*pb.Consignment
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *ConsignmentRepository) GetAll() []*pb.Consignment {
	return repo.consignments
}

/* service struct implements the defined protobuf interface.
// Server API for ShippingService service

	type ShippingServiceServer interface {
		CreateConsignment(context.Context, *Consignment) (*Response, error)
		GetConsignments(context.Context, *GetRequest) (*Response, error)
	}
*/
type service struct {
	repo         Repository
	vesselClient vesselProto.VesselServiceClient
}

func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {
	// Call a client instance of our vessel service
	// with our consignment weight and the amount of containers as the capacity value
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)

	// Set the VesselId as the vessel we got back
	// from our vessel service
	req.VesselId = vesselResponse.Vessel.Id

	// Save our consignment
	consignment, err := s.repo.Create(req)
	if err != nil {
		return nil, err
	}

	return &pb.Response{Created: true, Consignment: consignment}, nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {
	consignments := s.repo.GetAll()
	return &pb.Response{Consignments: consignments}, nil
}

func main() {
	repo := &ConsignmentRepository{}

	conn, err := grpc.Dial(vesselService, grpc.WithInsecure())
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

	pb.RegisterShippingServiceServer(s, &service{repo, vesselClient})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
