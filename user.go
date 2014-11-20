package goose

type User struct {
}

func (u *User) CanReadBucket(bucket *Bucket) bool {
	return true
}

func (u *User) CanWriteBucket(bucket *Bucket) bool {
	return true
}
