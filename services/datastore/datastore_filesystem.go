package datastore

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

// Fs is a file system datastore
type Fs struct {
	basePath string
}

// InitFilesystemDatastore initialize datastore as file system datastore
func InitFilesystemDatastore(basePath string) error {
	finfo, err := os.Stat(basePath)
	if err != nil {
		return err
	}
	if !finfo.IsDir() {
		return fmt.Errorf("%s is not a directory", basePath)
	}
	ds = &Fs{basePath: basePath}
	return nil
}

// Put implements datastore.put
func (d *Fs) put(key string, value []byte) error {
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

// get implements datastore.Get
func (d *Fs) get(key string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath.Join(d.getPath(key), key))
	if os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return data, err
}

// delete implements datastore.Delete
func (d *Fs) delete(key string) error {
	err := os.Remove(filepath.Join(d.getPath(key), key))
	if os.IsNotExist(err) {
		return ErrNotFound
	}
	return err
}

// getPath returns storage path
func (d *Fs) getPath(key string) (fPath string) {
	fPath = d.basePath
	runes := []rune(key)
	if len(key) > 4 {
		fPath = filepath.Join(fPath, string(runes[0:2]), string(runes[2:4]))
	}
	return
}
