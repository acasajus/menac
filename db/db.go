package db

import (
	"errors"
	"time"

	as "github.com/aerospike/aerospike-client-go"
)

type DB interface {
	LinkRecordToDB(RecordObject) RecordObject
	CreateNewRecord(RecordObject) error
	DeleteRecord(RecordObject) (bool, error)
	GetRecord([]byte, RecordObject) error
	TouchRecord(RecordObject) error
	ExistsRecord(r RecordObject) (bool, error)
	ReplaceRecord(r RecordObject) error
	ScanRecords(r RecordObject) chan ChanRecord
	RegisterIndexes(r RecordObject) error
	DeleteIndexes(r RecordObject) error
	Search(RecordObject, string, string) chan ChanRecord
}

type RecordObject interface {
	recordData
	Validate() error
}

func NewDB(namespace string, host string, port int) (DB, error) {
	c, err := as.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	if !c.IsConnected() {
		return nil, errors.New("Client says it's not connected to the aerospike cluster")
	}
	return &db{namespace, c}, nil
}

func NewTestDB(host string, port int) (DB, error) {
	c, err := as.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	c.DefaultWritePolicy = as.NewWritePolicy(0, 10)
	if !c.IsConnected() {
		return nil, errors.New("Client says it's not connected to the aerospike cluster")
	}
	return &db{"test", c}, nil
}

type db struct {
	namespace string
	client    *as.Client
}

func (d *db) LinkRecordToDB(r RecordObject) RecordObject {
	r.setDB(d)
	return r
}

func (d *db) CreateNewRecord(r RecordObject) error {
	if err := r.Validate(); err != nil {
		return err
	}
	pk, bins, err := structToData(r)
	if err != nil {
		return err
	}
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return err
	}
	r.setCreatedAt(time.Now())
	r.setUpdatedAt(r.GetCreatedAt())
	wPolicy := as.NewWritePolicy(r.GetGeneration(), r.GetExpiration())
	wPolicy.RecordExistsAction = as.CREATE_ONLY
	if err := d.client.PutBins(wPolicy, key, bins...); err != nil {
		return err
	}
	r.setStored()
	return nil
}

func (d *db) GetRecord(pk []byte, r RecordObject) error {
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return err
	}
	record, err := d.client.Get(nil, key)
	if err != nil {
		return err
	}
	if record == nil {
		return ERR_NO_EXIST
	}
	if err = recordToStruct(record, r); err != nil {
		return err
	}
	r.setDB(d)
	return nil
}

func (d *db) ReplaceRecord(r RecordObject) error {
	if err := r.Validate(); err != nil {
		return err
	}
	r.setUpdatedAt(time.Now())
	pk, bins, err := structToData(r)
	if err != nil {
		return err
	}
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return err
	}
	wPolicy := as.NewWritePolicy(r.GetGeneration(), r.GetExpiration())
	wPolicy.RecordExistsAction = as.REPLACE_ONLY
	wPolicy.GenerationPolicy = as.EXPECT_GEN_EQUAL

	return d.client.PutBins(wPolicy, key, bins...)
}

func (d *db) DeleteRecord(r RecordObject) (bool, error) {
	//TODO: Only get pk instead of all the bins
	pk, _, err := structToData(r)
	if err != nil {
		return false, err
	}
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return false, err
	}
	wPolicy := as.NewWritePolicy(r.GetGeneration(), r.GetExpiration())
	wPolicy.GenerationPolicy = as.EXPECT_GEN_EQUAL

	return d.client.Delete(wPolicy, key)
}

func (d *db) TouchRecord(r RecordObject) error {
	//TODO: Only get pk instead of all the bins
	pk, _, err := structToData(r)
	if err != nil {
		return err
	}
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return err
	}

	return d.client.Touch(nil, key)
}

func (d *db) ExistsRecord(r RecordObject) (bool, error) {
	//TODO: Only get pk instead of all the bins
	pk, _, err := structToData(r)
	if err != nil {
		return false, err
	}
	key, err := as.NewKey(d.namespace, structName(r), pk)
	if err != nil {
		return false, err
	}

	return d.client.Exists(nil, key)
}

type ChanRecord struct {
	Record interface{}
	Error  error
}

func (d *db) ScanRecords(r RecordObject) chan ChanRecord {
	sp := as.NewScanPolicy()
	sp.IncludeBinData = true

	recordSet, err := d.client.ScanAll(nil, d.namespace, structName(r))
	ifc := make(chan ChanRecord)
	go func() {
		defer close(ifc)
		if err != nil {
			ifc <- ChanRecord{Error: err}
			return
		}
		for {
			select {
			case record, open := <-recordSet.Records:
				if !open {
					return
				}
				if err := recordToStruct(record, r); err != nil {
					ifc <- ChanRecord{Error: err}
					return
				}
				r.setDB(d)
				ifc <- ChanRecord{Record: r}
			// do something
			case err := <-recordSet.Errors:
				if err != nil {
					ifc <- ChanRecord{Error: err}
					return
				}
			}
		}
	}()
	return ifc
}

func (d *db) RegisterIndexes(r RecordObject) error {
	idxs := structIndexes(r)
	if len(idxs) == 0 {
		return nil
	}
	wPolicy := as.NewWritePolicy(0, 0)
	wPolicy.RecordExistsAction = as.CREATE_ONLY

	for _, idx := range idxs {
		idxTask, err := d.client.CreateIndex(wPolicy, d.namespace, structName(r), idx+"_index", idx, as.STRING)
		if err != nil {
			if !IsErrIndexExists(err) {
				continue
			}
			return err
		}
		for err := range idxTask.OnComplete() {
			if err != nil && IsErrIndexExists(err) {
				return err
			}
		}
	}
	return nil
}

func (d *db) DeleteIndexes(r RecordObject) error {
	idxs := structIndexes(r)
	if len(idxs) == 0 {
		return nil
	}
	for _, idx := range idxs {
		err := d.client.DropIndex(nil, d.namespace, structName(r), idx+"_index")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *db) Search(r RecordObject, indexName string, value string) chan ChanRecord {
	stm := as.NewStatement(d.namespace, structName(r))
	stm.Addfilter(as.NewEqualFilter(indexName, value))
	recordSet, err := d.client.Query(nil, stm)

	ifc := make(chan ChanRecord)
	go func() {
		defer close(ifc)
		if err != nil {
			ifc <- ChanRecord{Error: err}
			return
		}
		for {
			select {
			case record, open := <-recordSet.Records:
				if !open {
					return
				}
				if err := recordToStruct(record, r); err != nil {
					ifc <- ChanRecord{Error: err}
					return
				}
				r.setDB(d)
				ifc <- ChanRecord{Record: r}
			// do something
			case err := <-recordSet.Errors:
				if err != nil {
					ifc <- ChanRecord{Error: err}
					return
				}
			}
		}
	}()
	return ifc

}
