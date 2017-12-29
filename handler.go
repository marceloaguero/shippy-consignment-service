package main

import (
	"context"
	"log"

	pb "github.com/marceloaguero/shippy-consignment-service/proto"
	vesselProto "github.com/marceloaguero/shippy-vessel-service/proto"
	mgo "gopkg.in/mgo.v2"
)

/* service struct implements the defined protobuf interface.
// Server API for ShippingService service

	type ShippingServiceServer interface {
		CreateConsignment(context.Context, *Consignment) (*Response, error)
		GetConsignments(context.Context, *GetRequest) (*Response, error)
	}
*/
type service struct {
	session      *mgo.Session
	vesselClient vesselProto.VesselServiceClient
}

func (s *service) GetRepo() Repository {
	return &ConsignmentRepository{s.session.Clone()}
}

func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment) (*pb.Response, error) {
	repo := s.GetRepo()
	defer repo.Close()

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
	err = repo.Create(req)
	if err != nil {
		return nil, err
	}

	return &pb.Response{Created: true, Consignment: req}, nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest) (*pb.Response, error) {
	repo := s.GetRepo()
	defer repo.Close()

	consignments, err := repo.GetAll()
	if err != nil {
		return nil, err
	}

	return &pb.Response{Consignments: consignments}, nil
}
