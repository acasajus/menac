package ostore

import (
	"io"
	"strings"
	"time"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type S3Store struct {
	conn   *s3.S3
	bucket *s3.Bucket
}

const (
	S3_ACCES_KEY    = "access_key"
	S3_SECRET_KEY   = "secret_key"
	S3_ENDPOINT_URL = "endpoint"
	S3_BUCKET       = "bucket_name"
)

func (s *S3Store) Initialize(c map[string]string) error {
	auth, err := aws.GetAuth(c[S3_ACCES_KEY], c[S3_SECRET_KEY])
	if err != nil {
		return err
	}
	region := aws.Region{S3Endpoint: c[S3_ENDPOINT_URL]}
	s.conn = s3.New(auth, region)
	buckets, err := s.conn.ListBuckets()
	if err != nil {
		return err
	}
	for _, bucket := range buckets.Buckets {
		if bucket.Name == c[S3_BUCKET] {
			s.bucket = &bucket
			return nil
		}
	}
	s.bucket = &s3.Bucket{s.conn, c[S3_BUCKET]}
	return s.bucket.PutBucket(s3.Private)
}

func (s *S3Store) Put(so *StoredObject, data io.Reader, length int64) error {
	path := so.getPath()
	_, err := s.bucket.Head(path)
	if err != nil && !strings.Contains(err.Error(), "404") {
		return err
	}
	hr := NewHashReader(data)
	headers := map[string][]string{}
	if so.Expiration.After(time.Now()) {
		headers["X-Expiration"] = []string{so.Expiration.Format(time.RFC3339)}
	}
	for k, v := range so.Metadata {
		headers["X-Object-"+k] = []string{v}
	}
	if err := s.bucket.PutReaderHeader(path, hr, length, headers, s3.Private); err != nil {
		return err
	}
	if hr.HexDigest() != so.Hash {
		//TODO: Do somethign wit del's error
		s.bucket.Del(path)
		return ErrHashMismatch
	}
	return nil
}

func (s *S3Store) Get(so *StoredObject, data io.Writer) error {
	r, err := s.bucket.GetReader(so.getPath())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return ErrNotExists
		}
		return err
	}
	_, err = io.Copy(data, r)
	return err
}

func (s *S3Store) Delete(so *StoredObject) error {
	err := s.bucket.Del(so.getPath())
	if err != nil && strings.Contains(err.Error(), "404") {
		return ErrNotExists
	}
	return err
}
