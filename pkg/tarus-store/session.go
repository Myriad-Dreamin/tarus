package tarus_store

import (
	"context"
	"github.com/Myriad-Dreamin/tarus/api/tarus"
	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sync"
	"time"
)

type judgeSessionStore struct {
	db *DB
	l  sync.RWMutex
}

func NewJudgeSessionStore(db *DB) JudgeSessionStore {
	return &judgeSessionStore{
		db: db,
	}
}

const MinTimestamp = -62135596800

func (j *judgeSessionStore) SetJudgeSession(ctx context.Context, key []byte, meta *tarus.OCIJudgeSession) error {
	j.l.RLock()
	defer j.l.RUnlock()
	if err := update(ctx, j.db, func(tx *bolt.Tx) error {
		bkt, err := getOrCreateSessionBucket(tx, key)
		if err != nil {
			return errors.Wrapf(errdefs.ErrNotFound, "session key %v", key)
		}
		err = readSessionTimestamp(meta, bkt)
		if err != nil {
			return err
		}

		if meta.CreatedAt.GetSeconds() == MinTimestamp {
			meta.CreatedAt = timestamppb.New(time.Now().UTC())
		}
		meta.UpdatedAt = timestamppb.New(time.Now().UTC())
		return writeSession(meta, bkt)
	}); err != nil {
		return err
	}
	return nil
}

func (j *judgeSessionStore) GetJudgeSession(ctx context.Context, key []byte) (*tarus.OCIJudgeSession, error) {
	var info = new(tarus.OCIJudgeSession)
	if err := view(ctx, j.db, func(tx *bolt.Tx) error {
		bkt := getSessionBucket(tx, key)
		if bkt == nil {
			return errors.Wrapf(errdefs.ErrNotFound, "session key %v", key)
		}

		return readSession(info, bkt)
	}); err != nil {
		return nil, err
	}

	return info, nil
}

func readSessionTimestamp(meta *tarus.OCIJudgeSession, bkt *bolt.Bucket) error {
	var c, u time.Time
	if err := readTimestamps(bkt, &c, &u); err != nil {
		return err
	}

	meta.CreatedAt = timestamppb.New(c)
	meta.UpdatedAt = timestamppb.New(u)
	return nil
}

func readSession(meta *tarus.OCIJudgeSession, bkt *bolt.Bucket) error {
	if err := readSessionTimestamp(meta, bkt); err != nil {
		return err
	}

	if b := bkt.Get(bucketKeyContainerId); b != nil {
		meta.ContainerId = string(b)
	}
	if b := bkt.Get(bucketKeyBinTarget); b != nil {
		meta.BinTarget = string(b)
	}
	if b := bkt.Get(bucketKeyWorkdir); b != nil {
		meta.HostWorkdir = string(b)
	}
	if b := bkt.Get(bucketKeyStatus); b != nil {
		statusDecoded, err := decodeInt32(b)
		if err != nil {
			return err
		}
		meta.CommitStatus = statusDecoded
	}

	return nil
}

func writeSession(meta *tarus.OCIJudgeSession, bkt *bolt.Bucket) error {
	if err := writeTimestamps(bkt, meta.CreatedAt.AsTime(), meta.UpdatedAt.AsTime()); err != nil {
		return err
	}
	if err := bkt.Put(bucketKeyContainerId, []byte(meta.ContainerId)); err != nil {
		return err
	}
	if err := bkt.Put(bucketKeyBinTarget, []byte(meta.BinTarget)); err != nil {
		return err
	}
	if err := bkt.Put(bucketKeyWorkdir, []byte(meta.HostWorkdir)); err != nil {
		return err
	}

	// Write size
	statusEncoded, err := encodeInt32(meta.CommitStatus)
	if err != nil {
		return err
	}

	return bkt.Put(bucketKeyStatus, statusEncoded)
}
