package client

import (
	"github.com/syb-devs/goose"
)

type BucketsService struct {
	s *Service
}

func NewBucketsService(s *Service) *BucketsService {
	rs := &BucketsService{s: s}
	return rs
}

func (sv *BucketsService) Create(name string) (*goose.Bucket, error) {
	url, err := sv.s.url("/buckets", nil)
	if err != nil {
		return nil, err
	}
	res, err := sv.s.sendJSON("POST", url, dict{"Name": name})
	if err != nil {
		return nil, err
	}
	bucket := &goose.Bucket{}
	if err = decodeJSON(res.Body, bucket); err != nil {
		return nil, err
	}
	return bucket, nil
}

func (sv *BucketsService) Retrieve(bucketID string) (*goose.Bucket, error) {
	url, err := sv.s.url("/buckets/"+bucketID, nil)
	if err != nil {
		return nil, err
	}
	bucket := &goose.Bucket{}
	if err = sv.s.getInto(url, bucket); err != nil {
		return nil, err
	}
	return bucket, nil
}

func (sv *BucketsService) RetrieveByName(name string) (*goose.Bucket, error) {
	url, err := sv.s.url("/buckets/name/"+name, nil)
	if err != nil {
		return nil, err
	}
	bucket := &goose.Bucket{}
	if err = sv.s.getInto(url, bucket); err != nil {
		return nil, err
	}
	return bucket, nil
}
