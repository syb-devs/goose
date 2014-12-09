package client

import (
	"errors"
	"io"

	"bitbucket.org/syb-devs/goose"
)

var (
	ErrInvalidBucketID   = errors.New("invalid bucket id")
	ErrInvalidObjectID   = errors.New("invalid object id")
	ErrInvalidObjectName = errors.New("invalid object name")
)

type ObjectsService struct {
	s *Service
}

func NewObjectsService(s *Service) *ObjectsService {
	rs := &ObjectsService{s: s}
	return rs
}

func (sv *ObjectsService) Upload(bucketID, name, contentType string, data io.Reader) (*goose.Object, error) {
	defer panicHandler()

	if !goose.ValidObjectID(bucketID) {
		return nil, ErrInvalidBucketID
	}
	if name == "" {
		return nil, ErrInvalidObjectName
	}
	ps := &URLParams{
		Path:  dict{"bucket": bucketID},
		Query: dict{"name": name},
	}
	url, err := sv.s.url("/buckets/{bucket}/objects", ps)
	if err != nil {
		return nil, newCtxErr("building URL", err)
	}
	req, err := sv.s.newRequest("POST", url, data)
	if err != nil {
		return nil, newCtxErr("creating a new request", err)
	}
	req.Header.Set("Content-Type", contentType)

	res, err := sv.s.do(req)
	if err != nil {
		return nil, newCtxErr("processing request", err)
	}

	ret := &goose.Object{}
	if err = decodeJSON(res.Body, &ret); err != nil {
		return nil, newCtxErr("decoding JSON", err)
	}
	return ret, nil
}

func (sv *ObjectsService) Delete(bucketID, objectID string) error {
	if !goose.ValidObjectID(bucketID) {
		return ErrInvalidBucketID
	}
	if !goose.ValidObjectID(objectID) {
		return ErrInvalidObjectID
	}
	url, err := sv.s.url("/buckets/"+bucketID+"/objects/"+objectID, nil)
	if err != nil {
		return err
	}
	_, err = sv.s.delete(url)
	return err
}

func (sv *ObjectsService) Retrieve(bucketID, objectID string) (*goose.Object, error) {
	url, err := sv.s.url("/buckets/"+bucketID+"/objects/"+objectID, nil)
	if err != nil {
		return nil, err
	}
	object := &goose.Object{}
	if err = sv.s.getInto(url, object); err != nil {
		return nil, err
	}
	return object, nil
}

func (sv *ObjectsService) List(bucketID string) ([]goose.Object, error) {
	url, err := sv.s.url("/buckets/"+bucketID+"/objects", nil)
	if err != nil {
		return nil, err
	}
	oList := []goose.Object{}
	if err = sv.s.getInto(url, &oList); err != nil {
		return nil, err
	}
	return oList, nil
}
