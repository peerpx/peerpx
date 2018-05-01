package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
	Important note about keys

	key must only contains char that can be used for file path
	normally for our usage key are photo hash (base58(sha256(photoAsByte)))
	or variants as:
		- (base58(sha256(photoAsByte)))
		- (base58(sha256(photoAsByte)))_small
		_ (base58(sha256(photoAsByte)))_medium

	To avoid incompatibility may be we should re-encode base58(key) ?
	Which will be useless for a large majority of cases...

*/

// DatastoreFs is a file system datastore
type DatastoreFs struct {
	basePath string
}

// NewDatastoreFs return a file system datastore
func NewDatastoreFs(basePath string) (datastore Datastore, err error) {
	finfo, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}
	if !finfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", basePath)
	}
	return &DatastoreFs{basePath: basePath}, nil
}

// Put implements datastore.Put
func (d *DatastoreFs) Put(key string, value []byte) error {
	basePath := d.getPath(key)
	// path exists ? no -> create it
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(basePath, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	// write file
	return ioutil.WriteFile(filepath.Join(basePath, key), value, 0644)
}

// Get implements datastore.Get
func (d *DatastoreFs) Get(key string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(d.getPath(key), key))
	if os.IsNotExist(err) {
		return nil, ErrNotFoundInDatastore
	}
	return data, err
}

// Delete implements datastore.Delete
func (d *DatastoreFs) Delete(key string) error {
	err := os.Remove(filepath.Join(d.getPath(key), key))
	if os.IsNotExist(err) {
		return ErrNotFoundInDatastore
	}
	return err
}

// getPath returns storage path
func (d *DatastoreFs) getPath(key string) (fPath string) {
	fPath = d.basePath
	runes := []rune(key)
	if len(key) > 4 {
		fPath = filepath.Join(fPath, string(runes[0:2]), string(runes[2:4]))
	}
	return
}
