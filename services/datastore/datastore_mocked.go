package datastore

// Mocked fake datastore for testing
type Mocked struct {
	responseErr  error
	responseData []byte
}

// NewMocked return a fake datastore for testing puupose
func NewMocked(data []byte, err error) Datastore {
	return &Mocked{
		responseData: data,
		responseErr:  err,
	}
}

// Put implements datastore.Put
func (d *Mocked) Put(key string, value []byte) error {
	return d.responseErr
}

// Get implements datastore.Get
func (d *Mocked) Get(key string) (data []byte, err error) {
	return d.responseData, d.responseErr
}

// Delete implements datastore.Delete
func (d *Mocked) Delete(key string) error {
	return d.responseErr
}
