package coord

import (
	"io"
	"log"

	pb "github.com/acasajus/menac/coord/proto"
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type peerEmitFunc struct {
	id     uint64
	stream pb.Coordinate_EmitRaftStepServer
}

func RegisterServer(s *grpc.Server) {
	c := &coordSvc{}
	c.Initialize()
	pb.RegisterCoordinateServer(s, c)
}

type coordSvc struct {
	messageChan chan *raftpb.Message
	hub         *Hub
}

func (si *coordSvc) Initialize() {

}

func (si *coordSvc) EmitRaftStep(stream pb.Coordinate_EmitRaftStepServer) error {
	header, _ := metadata.FromContext(stream.Context())
	header = header
	id := uint64(0)
	//TODO: Add peer to raft
	si.hub.AddServerStream(id, stream)
	for {
		msg, err := stream.Recv()
		switch err {
		case io.EOF:
			return nil
		case nil:
			break
		default:
			return err
		}
		step := &raftpb.Message{}
		if err := step.Unmarshal(msg.Data); err != nil {
			log.Printf("Cannot unmarshal step %s", err)
			continue
		}
		si.messageChan <- step
	}
}

func (ci *coordSvc) Register(c context.Context, p *pb.PeerInfo) (*pb.PeerInfoList, error) {
	return &pb.PeerInfoList{}, nil
}
