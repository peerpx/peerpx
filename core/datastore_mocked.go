package core

// DatastoreMocked fake datastore for testing
type DatastoreMocked struct{}

// NewDatastoreMocked return a fake datastore for testing puupose
func NewDatastoreMocked() Datastore {
	return &DatastoreMocked{}
}

// Put implements datastore.Put
func (d *DatastoreMocked) Put(key string, value []byte) error {
	return nil
}

// Get implements datastore.Get
func (d *DatastoreMocked) Get(key string) (data []byte, err error) {
	return
}

// Delete implements datastore.Delete
func (d *DatastoreMocked) Delete(key string) error {
	return nil
}
