package db

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

var gdb DB

func getDB() DB {
	if gdb == nil {
		d, err := NewDB("test", "127.0.0.1", 3000)
		if err != nil {
			panic(err)
		}
		gdb = d
	}
	return gdb
}

type SomeTestStruct struct {
	Record

	Id   string `db:"pk"`
	Data string `db:"indexed"`
}

func (sts *SomeTestStruct) Validate() error {
	if len(sts.Data) == 0 {
		return errors.New("Data is empty")
	}
	return nil
}

func (sts *SomeTestStruct) ChangeData() string {
	sts.Data = time.Now().Format(time.RFC3339Nano)
	return sts.Data
}

func NewSTS() *SomeTestStruct {
	r := &SomeTestStruct{Record{}, "deleteme:" + time.Now().Format(time.RFC3339Nano), ""}
	r.SetExpiration(int32(5))
	return r
}

func TestCreateGetExistsTouchDeleteRecord(t *testing.T) {
	d := getDB()
	r := NewSTS()
	expiration := r.GetExpiration()
	if err := d.CreateNewRecord(r); err == nil {
		t.Fatalf("Could store invalid record")
	} else if err.Error() != "Data is empty" {
		t.Fatalf("Got unexpected error: %s", err)
	}
	r.ChangeData()
	if err := d.CreateNewRecord(r); err != nil {
		t.Fatalf("Could not create new record: %s", err)
	}
	if err := d.CreateNewRecord(r); err == nil {
		t.Fatalf("Could create duplicate record")
	}
	//Exists
	if ok, err := d.ExistsRecord(r); err != nil {
		t.Fatalf("Could check if record exists: %s", err)
	} else if !ok {
		t.Fatalf("Record does not exist")
	}

	time.Sleep(1100 * time.Millisecond)
	if err := d.GetRecord([]byte(r.Id), r); err != nil {
		t.Fatalf("Could not retrieve record: %s", err)
	}
	if r.GetExpiration() >= expiration {
		t.Errorf("Expiration didn't decrease %d vs %d", r.GetExpiration(), expiration)
	}
	if err := d.TouchRecord(r); err != nil {
		t.Fatalf("Could not touch record: %s", err)
	}
	if err := d.GetRecord([]byte(r.Id), r); err != nil {
		t.Fatalf("Could not retrieve record: %s", err)
	}
	if r.GetExpiration() < expiration {
		t.Error("Expiration didn't reset")
	}

	if ok, err := d.DeleteRecord(r); err != nil {
		t.Fatalf("Could not delete record: %s", err)
	} else if !ok {
		t.Fatal("Did not delete record")
	}
	if ok, err := d.DeleteRecord(r); err == nil && ok {
		t.Fatal("Could delete record twice ")
	} else if err != nil {
		t.Fatalf("Error while deleting record: %s", err)
	}
	if ok, err := d.ExistsRecord(r); err != nil {
		t.Fatalf("Could check if record exists: %s", err)
	} else if ok {
		t.Fatalf("Record still exists")
	}
}

func TestReplaceRecord(t *testing.T) {
	d := getDB()
	r := NewSTS()
	pre := r.ChangeData()
	if err := d.CreateNewRecord(r); err != nil {
		t.Fatalf("Could not create new record: %s", err)
	}
	if err := d.GetRecord([]byte(r.Id), r); err != nil {
		t.Fatalf("Could not retrieve record: %s", err)
	}
	if pre != r.Data {
		t.Fatalf("Data is diferent %s vs %s", pre, r.Data)
	}
	changed := r.ChangeData()
	if err := d.ReplaceRecord(r); err != nil {
		t.Fatalf("Could not replace record: %s", err)
	}
	if err := d.GetRecord([]byte(r.Id), r); err != nil {
		t.Fatalf("Could not retrieve record: %s", err)
	}
	if changed != r.Data {
		t.Fatalf("Changed data is diferent %s vs %s", pre, r.Data)
	}
}

func TestScan(t *testing.T) {
	d := getDB()
	r := NewSTS()
	r.ChangeData()
	pk := r.Id
	if err := d.CreateNewRecord(r); err != nil {
		t.Fatalf("Could not create new record: %s", err)
	}
	found := false
	for sr := range d.ScanRecords(r) {
		if sr.Error != nil {
			t.Fatalf("Error while scanning records: %s", sr.Error)
			continue
		}
		b, ok := sr.Record.(*SomeTestStruct)
		if !ok {
			t.Fatalf("Record was not of expected type %#v vs %#v", b, sr.Record)
			continue
		}
		if b.Id == pk {
			found = true
		}
	}
	if !found {
		t.Error("Record was not found during scan")
	}
}

func TestCreateDropIndex(t *testing.T) {
	d := getDB()
	r := NewSTS()
	if err := d.DeleteIndexes(r); err != nil {
		t.Error(err)
	}
	if err := d.RegisterIndexes(r); err != nil {
		t.Errorf("Could not create index: %s", err)
	}
	if err := d.RegisterIndexes(r); !IsErrIndexExists(err) {
		t.Errorf("Unexpected error: %s", err)
	}
	if err := d.DeleteIndexes(r); err != nil {
		t.Error(err)
	}
}

func TestSearch(t *testing.T) {
	d := getDB()
	r := NewSTS()
	//Make sure indexes are there
	if err := d.RegisterIndexes(r); err != nil && !IsErrIndexExists(err) {
		t.Fatalf("Could not create index: %s", err)
	}

	values := []string{"data1", "data2", "data3" + time.Now().Format(time.RFC3339Nano)}
	for c, v := range values {
		r.Data = v
		for i := 0; i <= c; i++ {
			r.Id = fmt.Sprintf("%s:%d:%d", time.Now().Format(time.RFC3339Nano), i, c)
			if err := d.CreateNewRecord(r); err != nil {
				t.Fatalf("Could not create new record: %s", err)
			}
		}
	}
	found := 0
	for sr := range d.Search(r, "Data", values[len(values)-1]) {
		if sr.Error != nil {
			t.Fatalf("Error while scanning records: %s", sr.Error)
			continue
		}
		b, ok := sr.Record.(*SomeTestStruct)
		if !ok {
			t.Fatalf("Record was not of expected type %#v vs %#v", b, sr.Record)
			continue
		}
		found += 1
	}

	if found != len(values) {
		t.Fatalf("Found %d records instead of 3", found)
	}
}
