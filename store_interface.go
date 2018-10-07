package main

// ReadOnlyKVStore is a simple  interface to query data.
type ReadOnlyKVStore interface {
	Get(key []byte) []byte
	Has(key []byte) bool
	Iterator(start, end []byte) Iterator
	ReverseIterator(start, end []byte) Iterator
}

// interface for writing
type SetDeleter interface {
	Set(key, value []byte)
	Delete(key []byte)
}

// interface to get/set data
type KVStore interface {
	ReadOnlyKVStore
	SetDeleter
	// NewBatch return s a batch that can write multiple ops atomically
	NewBatch() Batch
}

// Batch can write multiple ops atomically to an underlying KVStore
type Batch interface {
	SetDeleter
	Write()
}

type Iterator interface {
	Valid() bool
	Next()
	Key() (key []byte)
	Value() (Value []byte)

	// Close releases the Iterator
	Close()
}

type CacheableKVStore interface {
	KVStore
	CacheWrap() KVCacheWrap
}

type KVCacheWrap interface {
	CacheableKVStore
	Write()
	Discard()
}

type CommitKVStore interface {
	Get(key []byte) []byte
	CacheWrap() KVCacheWrap
	Commit() CommitID
	LoadLatestVersion() error
	LastestVersion() CommitID
}

type CommitID struct {
	Version int64
	Hash    []byte
}
