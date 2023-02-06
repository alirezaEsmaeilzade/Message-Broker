package main

import (
	"context"
	"net/http"
	metrics "therealbroker/cmd"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net"
	//"net/http"
	_ "net/http/pprof"
	pb "therealbroker/api/proto"
	//metrics "therealbroker/cmd"
	module "therealbroker/internal/broker"
	br "therealbroker/pkg/broker"
	"time"
)

const (
	port       = ":5050"
	metricPort = ":8000"
)

type Server struct {
	pb.UnimplementedBrokerServer
}

func convertErrorToGrpcError(err error) error{
	switch err {
	case br.ErrUnavailable:
		return status.Error(codes.Unavailable, br.ErrUnavailable.Error())
	case br.ErrInvalidID:
		return status.Error(codes.InvalidArgument, br.ErrInvalidID.Error())
	case br.ErrExpiredID:
		return status.Error(codes.InvalidArgument, br.ErrExpiredID.Error())
	case nil:
		return nil
	default:
		return status.Error(codes.NotFound, "")
	}
}
var broker = module.NewModule()


func (s *Server) Publish(ctx context.Context, in *pb.PublishRequest) (*pb.PublishResponse, error) {
	start := time.Now()
	defer metrics.Latency.WithLabelValues("publish").Observe(float64(time.Since(start).Nanoseconds()))
	msg := br.Message{
		Body:       string(in.Body),
		Expiration: time.Duration(in.ExpirationSeconds) * time.Second,
	}
	ID, err := broker.Publish(ctx, in.Subject, msg)
	if err != nil {
		metrics.FailedCalls.WithLabelValues("publish").Inc()
	} else {
		metrics.SucceedCalls.WithLabelValues("publish").Inc()
	}
	return &pb.PublishResponse{Id: int32(ID)}, convertErrorToGrpcError(err)
}

func (s *Server) Subscribe(in *pb.SubscribeRequest, outStream pb.Broker_SubscribeServer) error {
	start := time.Now()
	defer metrics.Latency.WithLabelValues("publish").Observe(float64(time.Since(start).Nanoseconds()))
	ch, err := broker.Subscribe(outStream.Context(), in.Subject)
	if err != nil {
		metrics.FailedCalls.WithLabelValues("subscribe").Inc()
		return convertErrorToGrpcError(err)
	}
	metrics.ActiveSubscriptions.Inc()
	defer metrics.ActiveSubscriptions.Dec()
	for message := range ch{
		messageResponse := &pb.MessageResponse{Body: []byte(message.Body)}
		err := outStream.Send(messageResponse)
		if err != nil {
			return convertErrorToGrpcError(err)
		}else{
			log.Println("Message send to channel")
		}
	}
	metrics.SucceedCalls.WithLabelValues("subscribe").Inc()
	return nil

}

func (s *Server) Fetch(ctx context.Context, in *pb.FetchRequest) (*pb.MessageResponse, error) {
	msg, err := broker.Fetch(ctx, in.Subject, int(in.Id))
	if err != nil {
		return nil, convertErrorToGrpcError(err)
	}
	return &pb.MessageResponse{Body: []byte(msg.Body)}, nil
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(metricPort, nil))
	}()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterBrokerServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}