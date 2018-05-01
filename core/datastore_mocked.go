package core

// DatastoreMocked fake datastore for testing
type DataStoreMocked struct{}

// NewDatastoreMocked return a fake datastore for testing puupose
func NewDatastoreMocked() Datastore {
	return &DataStoreMocked{}
}

// Put implements datastore.Put
func (d *DataStoreMocked) Put(key string, value []byte) error {
	return nil
}

// Get implements datastore.Get
func (d *DataStoreMocked) Get(key string) (data []byte, err error) {
	return
}

// Delete implements datastore.Delete
func (d *DataStoreMocked) Delete(key string) error {
	return nil
}
