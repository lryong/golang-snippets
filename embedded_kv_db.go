package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

type MirrorStatus struct {
	Name       string    `json:"name"`
	Worker     string    `json:"worker"`
	IsMaster   bool      `json:"is_master"`
	Status     string    `json:"status"`
	LastUpdate time.Time `json:"last_updated"`
	LastEnded  time.Time `json:"last_ended"`
	Upstream   string    `json:"upstream"`
	Size       string    `json:"size"`
	ErrorMsg   string    `json:"error_msg"`
}

type WorkerStatus struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`   // worker url
	Token      string    `json:"token"` // session token
	LastOnline time.Time `json:"last_online"`
}

type dbAdapter interface {
	Init() error
	ListWorkers() ([]WorkerStatus, error)
	DeleteWorker(workerID string) error
	CreateWorker(w WorkerStatus) (WorkerStatus, error)
	UpdateMirrorStatus(workerID, mirrorID string, status MirrorStatus) (MirrorStatus, error)
	GetMirrorStatus(workerID, mirrorID string) (MirrorStatus, error)
	ListMirrorStatus(workerID string) ([]MirrorStatus, error)
	ListAllMirrorStatus() ([]MirrorStatus, error)
	FlushDisabledJobs() error
	Close() error
}

func makeDBAdapter(dbType, dbFile string) (dbAdapter, error) {
	if dbType == "bolt" {
		innerDB, err := bolt.Open(dbFile, 06000, nil)
		if err != nil {
			return nil, err
		}
		db := boltAdapter{
			db:     innerDB,
			dbFile: dbFile,
		}
		err = db.Init()
		return &db, err
	}
	// unsupported db-type
	return nil, fmt.Errorf("unsupported db-type: %s", dbType)
}

const (
	_workerBucketKey = "workers"
	_statusBucketKey = "mirror_status"
	Disabled         = "1"
)

type boltAdapter struct {
	db     *bolt.DB
	dbFile string
}

func (b *boltAdapter) Init() (err error) {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(_workerBucketKey))
		if err != nil {
			return fmt.Errorf("create bucket %s error: %s", _statusBucketKey, err.Error())
		}

		_, err := tx.CreateBucketIfNotExists([]byte(_statusBucketKey))
		if err != nil {
			return fmt.Errorf("create bucket %s error: %s", _statusBucketKey, err.Error())
		}

		return nil
	})
}

func (b *boltAdapter) ListWorkers() (ws []WorkerStatus, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_workerBucketKey))
		c := bucket.Cursor()
		var w WorkerStatus
		for k, v := c.First(); k != nil; k, v = c.Next() {
			jsonErr := json.Unmarshal(v, &w)
			if jsonErr != nil {
				err = fmt.Errorf("%s; %s", err.Error(), jsonErr)
				continue
			}
			ws = append(ws, w)
		}
		return err
	})
	return
}

func (b *boltAdapter) GetWorker(workerID string) (w WorkerStatus, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_workerBucketKey))
		v := bucket.Get([]byte(workerID))
		if v == nil {
			return fmt.Errorf("invalid workerID %s", workerID)
		}
		err := json.Unmarshal(v, &w)
		return err
	})
	return
}

func (b *boltAdapter) DeleteWorker(workerID string) (err error) {
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_workerBucketKey))
		v := bucket.Get([]byte(workerID))
		if v == nil {
			return fmt.Errorf("invalid workerID %s", workerID)
		}
		err := bucket.Delete([]byte(workerID))
		return err
	})
	return
}

func (b *boltAdapter) CreateWorker(w WorkerStatus) (WorkerStatus, error) {
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_workerBucketKey))
		v, err := json.Marshal(w)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(w.ID), v)
		return err
	})
	return w, err
}

func (b *boltAdapter) UpdateMirrorStatus(workerID, mirrorID string, status MirrorStatus) (MirrorStatus, error) {
	id := mirrorID + "/" + workerID
	err := b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_statusBucketKey))
		v, err := json.Marshal(status)
		err = bucket.Put([]byte(id), v)
		return err
	})
	return status, err
}

func (b *boltAdapter) GetMirrorStatus(workerID, mirrorID string) (m MirrorStatus, err error) {
	id := workerID + "/" + mirrorID
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_statusBucketKey))
		v := bucket.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("no mirror '%s' exists in worker '%s'", mirrorID, workerID)
		}
		err := json.Unmarshal(v, &m)
		return err
	})
	return
}

func (b *boltAdapter) ListMirrorStatus(workerID string) (ms []MirrorStatus, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_statusBucketKey))
		c := bucket.Cursor()
		var m MirrorStatus
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if wID := strings.Split(string(k), "/")[1]; wID == workerID {
				jsonErr := json.Unmarshal(v, &m)
				if jsonErr != nil {
					err = fmt.Errorf("%s; %s", err.Error(), jsonErr)
					continue
				}
				ms = append(ms, m)
			}
		}
		return err
	})
	return
}

func (b *boltAdapter) ListAllMirrorStatus() (ms []MirrorStatus, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_statusBucketKey))
		c := bucket.Cursor()
		var m MirrorStatus
		for k, v := c.First(); k != nil; k, v = c.Next() {
			jsonErr := json.Unmarshal(v, &m)
			if jsonErr != nil {
				err = fmt.Errorf("%s; %s", err.Error(), jsonErr)
				continue
			}
			ms = append(ms, m)
		}
		return err
	})
	return
}

func (b *boltAdapter) FlushDisabledJobs() (err error) {
	err = b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(_statusBucketKey))
		c := bucket.Cursor()
		var m MirrorStatus
		for k, v := c.First(); k != nil; k, v = c.Next() {
			jsonErr := json.Unmarshal(v, &m)
			if jsonErr != nil {
				err = fmt.Errorf("%s; %s", err.Error(), jsonErr)
				continue
			}
			if m.Status == Disabled || len(m.Name) == 0 {
				err = c.Delete()
			}
		}
		return err
	})
	return
}

func (b *boltAdapter) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
