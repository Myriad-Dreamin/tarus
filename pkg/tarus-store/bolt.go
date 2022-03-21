package tarus_store

/*
   Copyright The containerd Authors.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

type transactionKey struct{}

// WithTransactionContext returns a new context holding the provided
// bolt transaction. Functions which require a bolt transaction will
// first check to see if a transaction is already created on the
// context before creating their own.
func WithTransactionContext(ctx context.Context, tx *bolt.Tx) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

type transactor interface {
	View(fn func(*bolt.Tx) error) error
	Update(fn func(*bolt.Tx) error) error
}

// view gets a bolt db transaction either from the context
// or starts a new one with the provided bolt database.
func view(ctx context.Context, db transactor, fn func(*bolt.Tx) error) error {
	tx, ok := ctx.Value(transactionKey{}).(*bolt.Tx)
	if !ok {
		return db.View(fn)
	}
	return fn(tx)
}

// update gets a writable bolt db transaction either from the context
// or starts a new one with the provided bolt database.
func update(ctx context.Context, db transactor, fn func(*bolt.Tx) error) error {
	tx, ok := ctx.Value(transactionKey{}).(*bolt.Tx)
	if !ok {
		return db.Update(fn)
	} else if !tx.Writable() {
		return fmt.Errorf("unable to use transaction from context: %w", bolt.ErrTxNotWritable)
	}
	return fn(tx)
}

func getBucket(tx *bolt.Tx, keys ...[]byte) *bolt.Bucket {
	bkt := tx.Bucket(keys[0])

	for _, key := range keys[1:] {
		if bkt == nil {
			break
		}
		bkt = bkt.Bucket(key)
	}

	return bkt
}

func readTimestamps(bkt *bolt.Bucket, created, updated *time.Time) error {
	for _, f := range []struct {
		b []byte
		t *time.Time
	}{
		{bucketKeyCreatedAt, created},
		{bucketKeyUpdatedAt, updated},
	} {
		v := bkt.Get(f.b)
		if v != nil {
			if err := f.t.UnmarshalBinary(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeTimestamps(bkt *bolt.Bucket, created, updated time.Time) error {
	createdAt, err := created.MarshalBinary()
	if err != nil {
		return err
	}
	updatedAt, err := updated.MarshalBinary()
	if err != nil {
		return err
	}
	for _, v := range [][2][]byte{
		{bucketKeyCreatedAt, createdAt},
		{bucketKeyUpdatedAt, updatedAt},
	} {
		if err := bkt.Put(v[0], v[1]); err != nil {
			return err
		}
	}

	return nil
}

func encodeInt32(i int32) ([]byte, error) {
	return encodeInt(int64(i))
}

func encodeInt(i int64) ([]byte, error) {
	var (
		buf      [binary.MaxVarintLen64]byte
		iEncoded = buf[:]
	)
	iEncoded = iEncoded[:binary.PutVarint(iEncoded, i)]

	if len(iEncoded) == 0 {
		return nil, fmt.Errorf("failed encoding integer = %v", i)
	}
	return iEncoded, nil
}

func decodeInt32(b []byte) (int32, error) {
	i, err := decodeInt(b)
	if int64(int32(i)) != i {
		return 0, fmt.Errorf("failed decoding integer = !%v", i)
	}
	return int32(i), err
}

func decodeInt(b []byte) (int64, error) {
	r, i := binary.Varint(b)

	if i <= 0 {
		return 0, fmt.Errorf("failed decoding integer = %v", i)
	}
	return r, nil
}
