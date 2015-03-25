package db

import (
	"reflect"
	"time"
	"unicode"
	"unicode/utf8"

	as "github.com/aerospike/aerospike-client-go"
)

func structName(s recordData) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = reflect.ValueOf(s).Elem().Type()
	}
	return t.Name()
}

func structGetPK(s recordData) ([]byte, error) {
	pk := s.GetPrimaryKey()
	if pk != nil {
		return pk, nil
	}
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		sv = sv.Elem()
		st = sv.Type()
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if field.Anonymous {
			continue
		}
		if rune, _ := utf8.DecodeRuneInString(field.Name); unicode.IsLower(rune) {
			continue
		}
		tag := field.Tag.Get("db")
		if tag != "pk" {
			continue
		}
		if pk != nil {
			return nil, ERR_MULTIPLE_PK
		}
		v := sv.Field(i)
		switch v.Kind() {
		case reflect.String:
			pk = []byte(v.String())
		case reflect.Slice:
			var ok bool
			pk, ok = v.Interface().([]byte)
			if !ok {
				return nil, ERR_INVALID_PK
			}
			if pk == nil {
				return nil, ERR_INVALID_PK
			}
		default:
			return nil, ERR_INVALID_PK
		}
	}
	if pk == nil {
		return nil, ERR_INVALID_PK
	}
	return pk, nil
}

func structToData(s recordData) ([]byte, []*as.Bin, error) {
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		sv = sv.Elem()
		st = sv.Type()
	}
	bins := make([]*as.Bin, 0)
	pk := s.GetPrimaryKey()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if field.Anonymous {
			continue
		}
		if rune, _ := utf8.DecodeRuneInString(field.Name); unicode.IsLower(rune) {
			continue
		}
		tag := field.Tag.Get("db")
		switch tag {
		case "-":
			continue
		case "pk":
			if pk != nil {
				return nil, nil, ERR_MULTIPLE_PK
			}
			v := sv.Field(i)
			switch v.Kind() {
			case reflect.String:
				pk = []byte(v.String())
			case reflect.Slice:
				var ok bool
				pk, ok = v.Interface().([]byte)
				if !ok {
					return nil, nil, ERR_INVALID_PK
				}
				if pk == nil {
					return nil, nil, ERR_INVALID_PK
				}
			default:
				return nil, nil, ERR_INVALID_PK
			}
		}
		tag = field.Name
		iv := sv.Field(i).Interface()
		switch iv.(type) {
		case time.Time:
			bins = append(bins, as.NewBin(tag, iv.(time.Time).Format(time.RFC3339)))
		default:
			bins = append(bins, as.NewBin(tag, iv))
		}
	}
	if pk == nil {
		return nil, nil, ERR_NO_PK
	}
	bins = append(bins, as.NewBin("_CreatedAt", s.GetCreatedAt().Format(time.RFC3339)))
	bins = append(bins, as.NewBin("_UpdatedAt", s.GetUpdatedAt().Format(time.RFC3339)))
	return pk, bins, nil
}

func recordToStruct(r *as.Record, s recordData) error {
	st := reflect.TypeOf(s)
	if st.Kind() != reflect.Ptr {
		return ERR_NO_POINTER
	}
	sv := reflect.ValueOf(s).Elem()
	st = sv.Type()

	s.setStored()
	s.setGeneration(int32(r.Generation))
	s.SetExpiration(int32(r.Expiration))
	if ca, ok := r.Bins["_CreatedAt"]; ok {
		t, err := time.Parse(time.RFC3339, ca.(string))
		if err != nil {
			return err
		}
		s.setCreatedAt(t)
	}
	if ca, ok := r.Bins["_UpdatedAt"]; ok {
		t, err := time.Parse(time.RFC3339, ca.(string))
		if err != nil {
			return err
		}
		s.setCreatedAt(t)
	}
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if field.Anonymous {
			continue
		}
		if rune, _ := utf8.DecodeRuneInString(field.Name); unicode.IsLower(rune) {
			continue
		}
		tag := field.Tag.Get("db")
		switch tag {
		case "-":
			continue
		default:
			v := sv.Field(i)
			b, ok := r.Bins[field.Name]
			if !ok {
				continue
			}
			switch v.Kind() {
			case reflect.Array, reflect.Slice:
				switch v.Interface().(type) {
				case []byte:
					v.SetBytes(b.([]byte))
				default:
					bSlice := b.([]interface{})
					slice := reflect.MakeSlice(field.Type, 0, len(bSlice))
					for _, bp := range bSlice {
						slice = reflect.Append(slice, reflect.ValueOf(bp))
					}
					v.Set(slice)
				}
			default:
				switch v.Interface().(type) {
				case time.Time:
					t, err := time.Parse(time.RFC3339, b.(string))
					if err != nil {
						return ERR_DATA_TYPE_MISMATCH
					}
					v.Set(reflect.ValueOf(t))
				default:
					v.Set(reflect.ValueOf(b))
				}
			}
		}
	}
	return nil
}

func structIndexes(s recordData) []string {
	st := reflect.TypeOf(s)
	sv := reflect.ValueOf(s)
	if st.Kind() == reflect.Ptr {
		sv = sv.Elem()
		st = sv.Type()
	}
	indexes := make([]string, 0)
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		if field.Anonymous {
			continue
		}
		if rune, _ := utf8.DecodeRuneInString(field.Name); unicode.IsLower(rune) {
			continue
		}
		tag := field.Tag.Get("db")
		if tag != "indexed" {
			continue
		}
		v := sv.Field(i)
		if v.Kind() != reflect.String {
			panic("Indexed field" + field.Name + "can only be of string type")
		}
		indexes = append(indexes, field.Name)
	}
	return indexes
}

type recordData interface {
	SetExpiration(int32)
	GetGeneration() int32
	GetExpiration() int32
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	Stored() bool
	GetPrimaryKey() []byte
	GetDB() DB

	setGeneration(int32)
	setCreatedAt(time.Time)
	setUpdatedAt(time.Time)
	setStored()
	setDB(DB)
}

type Record struct {
	generation int32
	expiration int32
	updatedAt  time.Time
	createdAt  time.Time
	stored     bool
	db         DB
}

func (r *Record) setDB(d DB) {
	r.db = d
}

func (r *Record) GetDB() DB {
	return r.db
}

func (r *Record) GetPrimaryKey() []byte {
	return nil
}

func (r Record) GetGeneration() int32 {
	return r.generation
}
func (r Record) GetExpiration() int32 {
	return r.expiration
}
func (r Record) GetUpdatedAt() time.Time {
	return r.updatedAt
}
func (r Record) GetCreatedAt() time.Time {
	return r.createdAt
}
func (r Record) Stored() bool {
	return r.stored
}
func (r *Record) setGeneration(g int32) {
	r.generation = g
}
func (r *Record) SetExpiration(e int32) {
	r.expiration = e
}
func (r *Record) setUpdatedAt(t time.Time) {
	r.updatedAt = t
}
func (r *Record) setCreatedAt(t time.Time) {
	r.createdAt = t
}
func (r *Record) setStored() {
	r.stored = true
}
