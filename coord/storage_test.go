package coord

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	pb "github.com/coreos/etcd/raft/raftpb"
)

func createStorage() *BoltStorage {
	f, err := ioutil.TempFile("", "BoltStorageTest")
	if err != nil {
		panic(err)
	}
	name := f.Name()
	f.Close()
	bs, err := CreateBoltStorage(name)
	if err != nil {
		panic(err)
	}
	return bs
}

func deleteStorage(b *BoltStorage) {
	path := b.db.Path()
	if err := os.Remove(path); err != nil {
		panic(err)
	}
}

func TestStorageIndexes(t *testing.T) {
	b := createStorage()
	defer deleteStorage(b)

	for _, name := range []string{"first", "last"} {
		for exp, iter := 1, 1; iter < 10; exp, iter = exp*(exp+1), iter+1 {
			exp := uint64(1)
			if err := b.setUInt64(name, exp); err != nil {
				t.Fatal(err)
			}
			if i := b.getUInt64(name); i != exp {
				t.Errorf("%d: Unexpected index value %d vs expected %d", iter, i, exp)
			}
			f, l := b.getIndexes()
			switch name {
			case "first":
				if f != exp {
					t.Errorf("%d: Unexpected index value %d vs expected %d when retrieving both", iter, f, exp)
				}
				if f, _ := b.FirstIndex(); f != exp {
					t.Errorf("%d: Unexpected index value %d vs expected %d when comparing with func", iter, f, exp)
				}
			case "last":
				if l != exp {
					t.Errorf("%d: Unexpected index value %d vs expected %d when retrieving both", iter, l, exp)
				}
				if l, _ := b.LastIndex(); l != exp {
					t.Errorf("%d: Unexpected index value %d vs expected %d when comparing with func", iter, l, exp)
				}
			}
		}
	}
}

func TestStorageAppend(t *testing.T) {
	b := createStorage()
	defer deleteStorage(b)

	ents := []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}}
	tests := []struct {
		entries []pb.Entry

		werr     error
		wentries []pb.Entry
	}{
		{
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}},
		},
		{
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 6}, {Index: 5, Term: 6}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 6}, {Index: 5, Term: 6}},
		},
		{
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}, {Index: 6, Term: 5}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}, {Index: 6, Term: 5}},
		},
		// truncate incoming entries, truncate the existing entries and append
		{
			[]pb.Entry{{Index: 2, Term: 3}, {Index: 3, Term: 3}, {Index: 4, Term: 5}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 5}},
		},
		// tunncate the existing entries and append
		{
			[]pb.Entry{{Index: 4, Term: 5}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 5}},
		},
		// direct append
		{
			[]pb.Entry{{Index: 6, Term: 5}},
			nil,
			[]pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}, {Index: 6, Term: 5}},
		},
	}

	for i, tt := range tests {
		if err := b.forceEntries(ents); err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		err := b.Append(tt.entries)
		if err != tt.werr {
			t.Errorf("#%d: err = %v, want %v", i, err, tt.werr)
		}
		l, _ := b.LastIndex()
		out, err := b.Entries(ents[0].Index, l+1)
		if err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		if !reflect.DeepEqual(out, tt.wentries) {
			t.Errorf("#%d: entries = %v, want %v", i, out, tt.wentries)
		}
	}
}

func TestStorageCreateSnapshot(t *testing.T) {
	b := createStorage()
	defer deleteStorage(b)
	ents := []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}}
	cs := &pb.ConfState{Nodes: []uint64{1, 2, 3}}
	data := []byte("data")

	tests := []struct {
		i uint64

		werr  error
		wsnap pb.Snapshot
	}{
		{4, nil, pb.Snapshot{Data: data, Metadata: pb.SnapshotMetadata{Index: 4, Term: 4, ConfState: *cs}}},
		{5, nil, pb.Snapshot{Data: data, Metadata: pb.SnapshotMetadata{Index: 5, Term: 5, ConfState: *cs}}},
	}

	for i, tt := range tests {
		err := b.forceEntries(ents)
		if err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		snap, err := b.CreateSnapshot(tt.i, cs, data)
		if err != tt.werr {
			t.Errorf("#%d: err = %v, want %v", i, err, tt.werr)
		}
		if !reflect.DeepEqual(snap, tt.wsnap) {
			t.Errorf("#%d: snap = %+v, want %+v", i, snap, tt.wsnap)
		}
	}
}

func TestStorageCompact(t *testing.T) {
	b := createStorage()
	defer deleteStorage(b)
	ents := []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}}
	tests := []struct {
		i uint64

		werr   error
		windex uint64
		wterm  uint64
		wlen   int
	}{
		{2, ErrCompacted, 3, 3, 3},
		{3, ErrCompacted, 3, 3, 3},
		{4, nil, 4, 4, 2},
		{5, nil, 5, 5, 1},
	}

	for i, tt := range tests {
		err := b.forceEntries(ents)
		if err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		err = b.Compact(tt.i)
		if err != tt.werr {
			t.Errorf("#%d: err = %v, want %v", i, err, tt.werr)
		}
		f, _ := b.FirstIndex()
		l, _ := b.LastIndex()
		out, err := b.Entries(f, l+1)
		if err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		if out[0].Index != tt.windex {
			t.Errorf("#%d: index = %d, want %d", i, out[0].Index, tt.windex)
		}
		if out[0].Term != tt.wterm {
			t.Errorf("#%d: term = %d, want %d", i, out[0].Term, tt.wterm)
		}
		if len(out) != tt.wlen {
			t.Errorf("#%d: len = %d, want %d", i, len(out), tt.wlen)
		}
	}
}

func TestStorageTerm(t *testing.T) {
	b := createStorage()
	defer deleteStorage(b)
	ents := []pb.Entry{{Index: 3, Term: 3}, {Index: 4, Term: 4}, {Index: 5, Term: 5}}
	tests := []struct {
		i uint64

		werr  error
		wterm uint64
	}{
		{2, ErrUnavailable, 0},
		{3, nil, 3},
		{4, nil, 4},
		{5, nil, 5},
	}

	for i, tt := range tests {
		err := b.forceEntries(ents)
		if err != nil {
			t.Fatalf("#%d: err = %s", i, err)
		}
		term, err := b.Term(tt.i)
		if err != tt.werr {
			t.Errorf("#%d: err = %v, want %v", i, err, tt.werr)
		}
		if term != tt.wterm {
			t.Errorf("#%d: term = %d, want %d", i, term, tt.wterm)
		}
	}
}
