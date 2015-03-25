package coord

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	pb "github.com/coreos/etcd/raft/raftpb"
)

type BoltStorage struct {
	db *bolt.DB
}

const (
	STORAGE_RAFT_BUCKET = "conf-raft"
)

var ErrCompacted = errors.New("requested index is unavailable due to compaction")
var ErrSnapOutOfDate = errors.New("requested index is older than the existing snapshot")
var ErrUnavailable = errors.New("requested entry at index is unavailable")

func CreateBoltStorage(fileName string) (*BoltStorage, error) {
	db, err := bolt.Open(fileName, 0600, nil)
	if err != nil {
		return nil, err
	}
	return &BoltStorage{db}, nil
}

func (b *BoltStorage) Term(i uint64) (uint64, error) {
	e, err := b.getEntry(i)
	if err != nil {
		return 0, err
	}
	return e.Term, nil
}

func (b *BoltStorage) bucketWrite(do func(*bolt.Bucket) error) error {
	var err error
	bname := []byte(STORAGE_RAFT_BUCKET)
	err = b.db.Update(func(tx *bolt.Tx) error {
		u := tx.Bucket(bname)
		if u == nil {
			if u, err = tx.CreateBucket(bname); err != nil {
				return err
			}
		}
		return do(u)
	})
	return err
}

func (b *BoltStorage) bucketRead(do func(*bolt.Bucket) error) error {
	var err error
	bname := []byte(STORAGE_RAFT_BUCKET)
	err = b.db.View(func(tx *bolt.Tx) error {
		u := tx.Bucket(bname)
		if u == nil {
			return fmt.Errorf("Inexistant bucket: %s", STORAGE_RAFT_BUCKET)
		}
		return do(u)
	})
	return err
}

func (b *BoltStorage) getIndexes() (uint64, uint64) {
	var first, last uint64
	b.bucketRead(func(u *bolt.Bucket) error {
		d := u.Get([]byte("first-index"))
		if d != nil {
			a, c := binary.Uvarint(d)
			if c <= 0 {
				panic("first index has an invalid value on the store")
			}
			first = a
		}
		d = u.Get([]byte("last-index"))
		if d != nil {
			a, c := binary.Uvarint(d)
			if c <= 0 {
				panic("last index has an invalid value on the store")
			}
			last = a
		}
		return nil
	})
	return first, last
}

func (b *BoltStorage) getUInt64(name string) uint64 {
	index := uint64(0)
	b.bucketRead(func(u *bolt.Bucket) error {
		d := u.Get([]byte(name + "-uint64"))
		if d == nil {
			return errors.New("")
		}
		a, c := binary.Uvarint(d)
		if c <= 0 {
			panic(name + "index has an invalid value on the store")
		}
		index = a
		return nil
	})
	return index
}

func (b *BoltStorage) setUInt64(name string, index uint64, u ...*bolt.Bucket) error {
	d := make([]byte, 8)
	c := binary.PutUvarint(d, index)
	if len(u) > 0 {
		return u[0].Put([]byte(name+"-uint64"), d[:c])
	}
	err := b.bucketWrite(func(u *bolt.Bucket) error {
		return u.Put([]byte(name+"-uint64"), d)
	})
	return err
}

func (b *BoltStorage) FirstIndex() (uint64, error) {
	return b.getUInt64("first"), nil
}

func (b *BoltStorage) LastIndex() (uint64, error) {
	return b.getUInt64("last"), nil
}

func (b *BoltStorage) InitialState() (pb.HardState, pb.ConfState, error) {
	hs := b.getHardState()
	snap, err := b.Snapshot()
	return *hs, snap.Metadata.ConfState, err
}

func (b *BoltStorage) getHardState() *pb.HardState {
	var hs *pb.HardState
	b.bucketRead(func(u *bolt.Bucket) error {
		if data := u.Get([]byte("hardstate")); data == nil {
			return errors.New("")
		} else {
			return hs.Unmarshal(data)
		}
	})
	return hs
}

func (b *BoltStorage) SetHardState(st pb.HardState) error {
	err := b.bucketWrite(func(u *bolt.Bucket) error {
		m, err := st.Marshal()
		if err != nil {
			return nil
		}
		return u.Put([]byte("hardstate"), m)
	})
	return err
}

func (b *BoltStorage) SetNodeId(id uint64) error {
	return b.setUInt64("nodeid", id)
}

func (b *BoltStorage) GetNodeId() uint64 {
	return b.getUInt64("nodeid")
}

func (b *BoltStorage) CreateSnapshot(i uint64, cs *pb.ConfState, data []byte) (pb.Snapshot, error) {
	snap, err := b.Snapshot()
	if err != nil {
		return snap, err
	}
	if snap.Metadata.Index >= i {
		return snap, ErrSnapOutOfDate
	}
	last := b.getUInt64("last")
	if i > last {
		log.Panicf("snapshot %d is out of bound lastindex(%d)", i, last)
	}
	entry, err := b.getEntry(i)
	if err != nil {
		return snap, err
	}
	snap.Metadata.Index = i
	snap.Metadata.Term = entry.Term
	if cs != nil {
		snap.Metadata.ConfState = *cs
	}
	snap.Data = data
	return snap, nil
}

func (b *BoltStorage) ApplySnapshot(snap pb.Snapshot) error {
	return b.bucketWrite(func(u *bolt.Bucket) error {
		d, err := snap.Marshal()
		if err != nil {
			return err
		}
		u.Put([]byte("snapshot"), d)
		c := u.Cursor()
		prefix := []byte("entry-")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err = u.Delete(k); err != nil {
				return err
			}
		}
		entry := pb.Entry{Term: snap.Metadata.Term, Index: snap.Metadata.Index}
		d, err = entry.Marshal()
		if err != nil {
			log.Panicf("Marshal failed %s", err)
		}
		if err = u.Put([]byte(fmt.Sprintf("entry-%d", snap.Metadata.Index)), d); err != nil {
			return err
		}
		if err = b.setUInt64("first", snap.Metadata.Index, u); err != nil {
			return err
		}
		if err = b.setUInt64("last", snap.Metadata.Index, u); err != nil {
			return err
		}
		return nil
	})
}

func (b *BoltStorage) Snapshot() (pb.Snapshot, error) {
	var snap pb.Snapshot
	b.bucketRead(func(u *bolt.Bucket) error {
		d := u.Get([]byte("snapshot"))
		if d == nil {
			return nil
		}
		snap.Unmarshal(d)
		return nil
	})
	return snap, nil
}

func (b *BoltStorage) Compact(cIndex uint64) error {
	first, last := b.getIndexes()
	if cIndex <= first {
		return ErrCompacted
	}
	if cIndex > b.getUInt64("last") {
		log.Panicf("compact %d is out of bound lastindex(%d)", cIndex, last)
	}
	return b.bucketWrite(func(u *bolt.Bucket) error {
		c := u.Cursor()
		prefix := []byte("entry-")
		entry := pb.Entry{}
		for k, v := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, v = c.Next() {
			err := entry.Unmarshal(v)
			if err != nil {
				return errors.New("Corrupt entry in store")
			}
			if entry.Index < cIndex {
				if err = u.Delete(k); err != nil {
					return err
				}
			}
		}
		return b.setUInt64("first", cIndex, u)
	})
}

func (b *BoltStorage) forceEntries(entries []pb.Entry) error {
	return b.bucketWrite(func(u *bolt.Bucket) error {
		c := u.Cursor()
		prefix := []byte("entry-")
		for k, _ := c.Seek(prefix); bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := u.Delete(k); err != nil {
				return err
			}
		}
		for _, e := range entries {
			d, err := e.Marshal()
			if err != nil {
				return nil
			}
			if err := u.Put([]byte(fmt.Sprintf("entry-%d", e.Index)), d); err != nil {
				return err
			}
		}
		if err := b.setUInt64("first", entries[0].Index, u); err != nil {
			return nil
		}
		return b.setUInt64("last", entries[len(entries)-1].Index, u)
	})
}

func (b *BoltStorage) Append(entries []pb.Entry) error {
	switch len(entries) {
	case 0:
		return nil
	case 1:
		break
	default:
		cC := entries[0].Index
		for p, e := range entries[1:] {
			if cC+uint64(p)+1 != e.Index {
				return fmt.Errorf("Entry %d is not continuous (expected %d vs %d)", p, cC+uint64(p), e.Index)
			}
		}
	}

	first, last := b.getIndexes()
	if first >= entries[0].Index+uint64(len(entries))-1 {
		return nil
	}
	if last+1 < entries[0].Index {
		return errors.New("First entry index is not contiguous to last entry stored")
	}

	// truncate compacted entries
	if first > entries[0].Index {
		entries = entries[first-entries[0].Index:]
	}
	return b.bucketWrite(func(u *bolt.Bucket) error {
		for _, e := range entries {
			d, err := e.Marshal()
			if err != nil {
				return nil
			}
			if err := u.Put([]byte(fmt.Sprintf("entry-%d", e.Index)), d); err != nil {
				return err
			}
		}
		return b.setUInt64("last", entries[len(entries)-1].Index, u)
	})
}

func (b *BoltStorage) getEntry(index uint64) (pb.Entry, error) {
	var entry pb.Entry
	return entry, b.bucketRead(func(u *bolt.Bucket) error {
		k := []byte(fmt.Sprintf("entry-%d", index))
		d := u.Get(k)
		if d == nil {
			return ErrUnavailable
		}
		return entry.Unmarshal(d)
	})
}

func (b *BoltStorage) Entries(lo, hi uint64) ([]pb.Entry, error) {
	first, last := b.getIndexes()
	if lo < first {
		return nil, ErrCompacted
	}
	//limits are [lo,hi)
	if hi > last+1 {
		return nil, fmt.Errorf("entries's hi(%d) is out of bound lastindex(%d)", hi, last)
	}
	entries := make([]pb.Entry, 0, hi-1-lo)
	for i := lo; i <= last && i < hi; i++ {
		e, err := b.getEntry(i)
		if err != nil {
			return entries, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
