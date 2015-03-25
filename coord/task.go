package coord

import "github.com/coreos/etcd/raft/raftpb"

type Task struct {
	entries  []raftpb.Entry
	snapshot raftpb.Snapshot
	done     chan error
}

type TaskRunner struct {
	todo chan Task
	stop chan struct{}
}

func NewTaskRunner() *TaskRunner {
	tr := &TaskRunner{
		todo: make(chan Task),
		stop: make(chan struct{}),
	}
	go tr.run()
	return tr
}

func (tr *TaskRunner) run() {
	for {
		select {
		case task := <-tr.todo:
			//TODO: FOLLOW: https://sourcegraph.com/github.com/coreos/etcd@32105e6ed063ad0fba8077b3a446ef3cf476c17c/.tree/etcdserver/server.go#startline=353&endline=353
			task.done <- nil
		case <-tr.stop:
			return
		}
	}
}
