package core

// DatastoreMocked fake datastore for testing
type DatastoreMocked struct {
	responseErr  error
	responseData []byte
}

// NewDatastoreMocked return a fake datastore for testing puupose
func NewDatastoreMocked(data []byte, err error) Datastore {
	return &DatastoreMocked{
		responseData: data,
		responseErr:  err,
	}
}

// Put implements datastore.Put
func (d *DatastoreMocked) Put(key string, value []byte) error {
	return d.responseErr
}

// Get implements datastore.Get
func (d *DatastoreMocked) Get(key string) (data []byte, err error) {
	return d.responseData, d.responseErr
}

// Delete implements datastore.Delete
func (d *DatastoreMocked) Delete(key string) error {
	return d.responseErr
}
