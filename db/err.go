package db

import (
	"errors"

	ast "github.com/aerospike/aerospike-client-go/types"
)

var (
	ERR_NO_EXIST           = errors.New("Key doesn't exist")
	ERR_NO_PK              = errors.New("No field is marked as primary key")
	ERR_INVALID_PK         = errors.New("The field maked as pk is not a []byte or a string or is nil")
	ERR_MULTIPLE_PK        = errors.New("Either one field can be marked as primary key or the GetPrimaryKey function returns one")
	ERR_NO_POINTER         = errors.New("A pointer is required to unmarshal data into it")
	ERR_DATA_TYPE_MISMATCH = errors.New("Data stored and expected value in struct do not match")
)

func IsErrDuplicateKey(err error) bool {
	ae, ok := err.(ast.AerospikeError)
	return ok && ae.ResultCode() == ast.KEY_EXISTS_ERROR
}

func IsErrIndexExists(err error) bool {
	ae, ok := err.(ast.AerospikeError)
	return ok && ae.ResultCode() == ast.INDEX_FOUND
}
