package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/acasajus/menac/coord"
	"google.golang.org/grpc"
)

func main() {
	svcAddr := flag.String("connect", "", "address to connect to")
	port := flag.Int("port", 0, "Port to listen to")
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	coord.RegisterServer(grpcServer)
	node := NewNode()

	if *svcAddr != "" {
		conn, err := grpc.Dial(*svcAddr)
		if err != nil {
			log.Fatalln(err)
		}
		defer conn.Close()
		client := coord.NewClient(conn)
	}

	log.Println("Listening at", lis.Addr())
	grpcServer.Serve(lis)
}
