build:
	protoc -I. --go_out=plugins=grpc:. proto/consignment.proto
	docker build -t consignment-service .