package goose

//TODO: basic ACL
//TODO: retrieve the user from JWT auth in the request
type User struct {
}

func (u *User) CanReadBucket(bucket *Bucket) bool {
	return true
}

func (u *User) CanWriteBucket(bucket *Bucket) bool {
	return true
}
