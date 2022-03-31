package tarus_store

import bolt "go.etcd.io/bbolt"

var (
	bucketKeyVersion       = []byte("v0")
	bucketKeyObjectSession = []byte("session")

	bucketKeyStatus      = []byte("status")
	bucketKeyWorkerId    = []byte("worker_id")
	bucketKeyContainerId = []byte("container_id")
	bucketKeyBinTarget   = []byte("bin_target")
	bucketKeyWorkdir     = []byte("workdir")
	bucketKeyCreatedAt   = []byte("created_at")
	bucketKeyUpdatedAt   = []byte("updated_at")
)

func createBucketIfNotExists(tx *bolt.Tx, keys ...[]byte) (*bolt.Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(keys[0])
	if err != nil {
		return nil, err
	}

	for _, key := range keys[1:] {
		bkt, err = bkt.CreateBucketIfNotExists(key)
		if err != nil {
			return nil, err
		}
	}

	return bkt, nil
}

func getSessionBucket(tx *bolt.Tx, taskKey []byte) *bolt.Bucket {
	return getBucket(tx, bucketKeyVersion, bucketKeyObjectSession, taskKey)
}

func getOrCreateSessionBucket(tx *bolt.Tx, taskKey []byte) (*bolt.Bucket, error) {
	return createBucketIfNotExists(tx, bucketKeyVersion, bucketKeyObjectSession, taskKey)
}
