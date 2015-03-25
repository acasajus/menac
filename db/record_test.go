package db

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	as "github.com/aerospike/aerospike-client-go"
)

type RecordTest struct {
	Record
	When time.Time
}

func (r RecordTest) GetPrimaryKey() []byte {
	return []byte("This is pk")
}

func TestStructGetPK(t *testing.T) {
	rt := &struct {
		Record

		Id string `db:"pk"`
	}{Record{}, "pkvalue"}
	if pk, err := structGetPK(rt); err != nil {
		t.Errorf("Could not get PK: %s", err)
	} else if !bytes.Equal(pk, []byte("pkvalue")) {
		t.Error("Unexpected pk %s vs pkvalue", string(pk))
	}
	if pk, err := structGetPK(&RecordTest{Record{}, time.Now()}); err != nil {
		t.Errorf("Could not get PK: %s", err)
	} else if !bytes.Equal(pk, []byte("This is pk")) {
		t.Error("Unexpected pk")
	}

}

func TestStructToData(t *testing.T) {
	rt := &struct {
		Record

		Id           string `db:"pk"`
		SomeData     []string
		OtherData    int `db:"-"`
		privateStuff string
	}{Record{}, "pkvalue", nil, 4, "sad"}
	rt.createdAt = time.Now()
	rt.updatedAt = time.Now()
	pk, bins, err := structToData(rt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pk, []byte(rt.Id)) {
		t.Errorf("PK is different")
	}
	expected := []string{
		"Id:pkvalue",
		"SomeData:[]",
		"_CreatedAt:" + rt.GetCreatedAt().Format(time.RFC3339),
		"_UpdatedAt:" + rt.GetUpdatedAt().Format(time.RFC3339),
	}
	if len(bins) != len(expected) {
		t.Errorf("There aren't the expected bins")
	}
	for i, b := range bins {
		if b.String() != expected[i] {
			t.Errorf("Unexpected string for bin %d: %s (expected %s)", i, b.String(), expected[i])
		}
	}

	rt2 := &RecordTest{When: time.Now()}
	pk, bins, err = structToData(rt2)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pk, rt2.GetPrimaryKey()) {
		t.Errorf("PK is different %s vs expected %s", string(pk), string(rt2.GetPrimaryKey()))
	}
	expected = []string{
		"When:" + rt2.When.Format(time.RFC3339),
		"_CreatedAt:" + rt2.GetCreatedAt().Format(time.RFC3339),
		"_UpdatedAt:" + rt2.GetUpdatedAt().Format(time.RFC3339),
	}
	if len(bins) != len(expected) {
		t.Errorf("There aren't the expected bins")
	}
	for i, b := range bins {
		if b.String() != expected[i] {
			t.Errorf("Unexpected string for bin %d: %s (expected %s)", i, b.String(), expected[i])
		}
	}

	rt3 := &struct {
		Record
	}{}
	pk, bins, err = structToData(rt3)
	if err == nil {
		t.Fatal("OOps. Struct without pk was converted")
	}
	if err != ERR_NO_PK {
		t.Errorf("Unexpected error: %s", err)
	}

	rt4 := &struct {
		Record

		Id int `db:"pk"`
	}{}
	pk, bins, err = structToData(rt4)
	if err == nil {
		t.Fatal("OOps. Struct with invalid pk type was converted")
	}
	if err != ERR_INVALID_PK {
		t.Errorf("Unexpected error: %s", err)
	}

	rt5 := &struct {
		Record

		Id []string `db:"pk"`
	}{Id: []string{}}
	pk, bins, err = structToData(rt5)
	if err == nil {
		t.Fatal("OOps. Struct with invalid pk type was converted")
	}
	if err != ERR_INVALID_PK {
		t.Errorf("Unexpected error: %s", err)
	}

	rt6 := &struct {
		Record

		Id []byte `db:"pk"`
	}{Id: []byte("pkdata")}
	pk, bins, err = structToData(rt6)
	if err != nil {
		t.Errorf("[]byte pk didn't manage to convert: %s", err)
	}
	if !bytes.Equal(pk, rt6.Id) {
		t.Errorf("PK is different %s vs expected %s", string(pk), string(rt6.Id))
	}

	rt7 := &struct {
		Record

		Id  string `db:"pk"`
		Id2 string `db:"pk"`
	}{}
	pk, bins, err = structToData(rt7)
	if err == nil {
		t.Errorf("Multiple pk struct is converted")
	}
	if err != ERR_MULTIPLE_PK {
		t.Errorf("Unexpected error: %s", err)
	}
}

type RecordWithAll struct {
	Record

	Id           string `db:"pk"`
	SomeData     []string
	SomeBytes    []byte
	SomeOther    []int64
	OtherData    int    `db:"-"`
	IndexedData  string `db:"indexed"`
	privateStuff string
}

type SimpleRecord struct {
	*Record
}

func TestStructName(t *testing.T) {
	r1 := RecordWithAll{}
	n := structName(&r1)
	expected := "RecordWithAll"
	if n != expected {
		t.Errorf("Unexpected struct name: %s vs %s", n, expected)
	}
	r2 := SimpleRecord{&Record{}}
	n = structName(r2)
	expected = "SimpleRecord"
	if n != expected {
		t.Errorf("Unexpected struct name: %s vs %s", n, expected)
	}
}

func TestRecordToStruct(t *testing.T) {
	checkSlice := func(iter int, a interface{}, b interface{}) {
		if (a == nil || reflect.ValueOf(a).Len() == 0) && (b == nil || reflect.ValueOf(b).Len() == 0) {
			return
		}
		if a == nil {
		}

		if !reflect.DeepEqual(a, b) {
			t.Errorf("%d: Slice mismatch %#v vs %#v", iter, a, b)
		}
	}

	checkRecordToStructIsOk := func(iter int, r *RecordWithAll) {
		pk, bins, err := structToData(r)
		if err != nil {
			t.Fatalf("Could not convert struct: %s", err)
		}
		bMap := as.BinMap{}
		for _, b := range bins {
			bMap[b.Name] = b.Value.GetObject()
		}
		k, _ := as.NewKey("test", structName(r), pk)
		rec := &as.Record{
			Key:        k,
			Node:       nil,
			Bins:       bMap,
			Generation: 5,
			Expiration: 50,
		}
		rb := &RecordWithAll{}
		if err := recordToStruct(rec, rb); err != nil {
			t.Fatal(err)
		}
		if !r.GetCreatedAt().Equal(rb.GetCreatedAt()) {
			t.Errorf("%d: Created at mismatch %s vs %s", iter, r.GetCreatedAt(), rb.GetCreatedAt())
		}
		if !r.GetUpdatedAt().Equal(rb.GetUpdatedAt()) {
			t.Errorf("%d: Updated at mismatch %s vs %s", iter, r.GetUpdatedAt(), rb.GetUpdatedAt())
		}
		if int32(rec.Generation) != rb.GetGeneration() {
			t.Errorf("%d: Generation mismatch %s vs %s", iter, rec.Generation, rb.GetGeneration())
		}
		if int32(rec.Expiration) != rb.GetExpiration() {
			t.Errorf("%d: Expiration mismatch %s vs %s", iter, rec.Expiration, rb.GetExpiration())
		}
		if r.Id != rb.Id {
			t.Errorf("%d: Id (PK) mismatch %s vs %s", iter, r.Id, rb.Id)
		}
		checkSlice(iter, r.SomeData, rb.SomeData)
		checkSlice(iter, r.SomeOther, rb.SomeOther)
		checkSlice(iter, r.SomeBytes, rb.SomeBytes)
	}

	checkRecordToStructIsOk(0, &RecordWithAll{Record{}, "pkvalue", nil, nil, nil, 4, "ASD", "sad"})
	checkRecordToStructIsOk(1, &RecordWithAll{
		Record{},
		"some other crappy stuff",
		[]string{"A2", "ASD"},
		[]byte("asdas"),
		[]int64{4, 5, 5, 43, 3},
		4, "sad", "ASDAS",
	})
}

func TestStructIndexes(t *testing.T) {
	idx := structIndexes(&RecordWithAll{})
	if len(idx) != 1 || idx[0] != "IndexedData" {
		t.Errorf("Index list differs from expected %#v", idx)
	}
}
