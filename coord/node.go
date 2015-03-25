package coord

import (
	"log"
	"math/rand"
	"time"

	pb "github.com/acasajus/menac/coord/proto"
	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"golang.org/x/net/context"
)

type Node struct {
	storage  *BoltStorage
	clients  []pb.CoordinateClient
	raftNode raft.Node

	tasker      *TaskRunner
	done        chan struct{}
	messageChan chan *raftpb.Message
}

func (n *Node) generateId() uint64 {
	return uint64(time.Now().Unix() + rand.Int63())
}

func NewNode(s *BoltStorage, peers []raft.Peer) *Node {
	return &Node{
		clients:     make([]pb.CoordinateClient, 0),
		tasker:      NewTaskRunner(),
		done:        make(chan struct{}),
		messageChan: make(chan *raftpb.Message),
	}
}

func (n *Node) GetMessageChan() chan *raftpb.Message {
	return n.messageChan
}

func (n *Node) run() {
	//FOLLOW: https://sourcegraph.com/github.com/coreos/etcd@32105e6ed063ad0fba8077b3a446ef3cf476c17c/.tree/etcdserver/raft.go#selected=72
	ticker := time.Tick(time.Duration(100) * time.Millisecond)
	for {
		select {
		case <-ticker:
			n.raftNode.Tick()
		case step := <-n.messageChan:
			n.raftNode.Step(context.Background(), *step)
		case rd := <-n.raftNode.Ready():
			if rd.SoftState != nil {
				if rd.RaftState == raft.StateLeader {
					log.Println("I'm now the leader of the cluster")
				}
			}
			//TOOD: Apply snapshot if there's any + committed entries
			t := Task{
				entries:  rd.CommittedEntries,
				snapshot: rd.Snapshot,
				done:     make(chan error),
			}
			//Execute task or exit if we've go the we're finished msg
			select {
			case n.tasker.todo <- t:
			case <-n.done:
				return
			}
			//Store stuff in the DB
			if !raft.IsEmptySnap(rd.Snapshot) {
				if err := n.storage.ApplySnapshot(rd.Snapshot); err != nil {
					log.Fatalln("coord: Cannot apply snapshot: %s", err)
				}
				log.Printf("coord: applied incoming snapshot at index %d", rd.Snapshot.Metadata.Index)
			}
			if !raft.IsEmptyHardState(rd.HardState) {
				if err := n.storage.SetHardState(rd.HardState); err != nil {
					log.Fatalln("coord: Cannot save hard state: %s", err)
				}
			}
			if err := n.storage.Append(rd.Entries); err != nil {
				log.Fatalln("coord: Cannot save entries: %s", err)
			}

			//Send messages to known peers
			for _, m := range rd.Messages {
				m = m
				//TOOD: Send the message to everyone registered in the hub
				//n.hub.SendMessage(m)
			}

			//Wait until tasker has finished processing
			<-t.done
			n.raftNode.Advance()
		case <-n.done:
			return
		}
	}
}
