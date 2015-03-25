package coord

import (
	"log"

	pb "github.com/acasajus/menac/coord/proto"
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type RaftSender interface {
	Send(*pb.RaftStep) error
}

type hubAction struct {
	add          bool
	id           uint64
	client       pb.CoordinateClient
	serverStream pb.Coordinate_EmitRaftStepServer
}

type Hub struct {
	clientMap    map[uint64]pb.CoordinateClient
	sendMap      map[uint64]RaftSender
	sendChan     chan *raftpb.Message
	actionChan   chan *hubAction
	doneChan     chan struct{}
	receivedChan chan *raftpb.Message
}

func NewHub(receivedChan chan *raftpb.Message) *Hub {
	h := &Hub{
		receivedChan: receivedChan,
		clientMap:    make(map[uint64]pb.CoordinateClient),
		sendMap:      make(map[uint64]RaftSender),
		sendChan:     make(chan *raftpb.Message),
		actionChan:   make(chan *hubAction),
		doneChan:     make(chan struct{}),
	}
	go h.run()
	return h
}

func (h *Hub) SendMessage(msg *raftpb.Message) {
	h.sendChan <- msg
}

func (h *Hub) AddClient(id uint64, c *grpc.ClientConn) {
	h.actionChan <- &hubAction{
		add:    true,
		client: pb.NewCoordinateClient(c),
		id:     id,
	}
}

func (h *Hub) Remove(id uint64) {
	h.actionChan <- &hubAction{
		add: false,
		id:  id,
	}
}

func (h *Hub) AddServerStream(id uint64, ss pb.Coordinate_EmitRaftStepServer) {
	h.actionChan <- &hubAction{
		add:          true,
		serverStream: ss,
	}
}

func (h *Hub) clientReceive(id uint64, stream pb.Coordinate_EmitRaftStepClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			h.Remove(id)
		}
		step := &raftpb.Message{}
		if err := step.Unmarshal(msg.Data); err != nil {
			log.Printf("Cannot unmarshal raft message: %s", err)
			continue
		}
		h.receivedChan <- step
	}
}

func (h *Hub) run() {
	defer close(h.actionChan)
	defer close(h.sendChan)
	for {
		select {
		case <-h.doneChan:
			return
		case msg := <-h.sendChan:
			data, err := msg.Marshal()
			if err != nil {
				//TODO: Find a better solution than panic
				panic(err)
			}
			step := &pb.RaftStep{data}
			for id, stream := range h.sendMap {
				if err = stream.Send(step); err != nil {
					delete(h.sendMap, id)
					delete(h.clientMap, id)
				}
			}
		case action := <-h.actionChan:
			if action.add {
				if action.client != nil {
					stream, err := action.client.EmitRaftStep(context.Background())
					if err == nil {
						log.Printf("coord: Cannot create client stream for id %s: %s", action.id, err)
						continue
					}
					h.clientMap[action.id] = action.client
					h.sendMap[action.id] = stream
					go h.clientReceive(action.id, stream)
				}
				if action.serverStream != nil {
					h.sendMap[action.id] = action.serverStream
				}
			} else {
				delete(h.sendMap, action.id)
				delete(h.clientMap, action.id)
			}
		}
	}
}
