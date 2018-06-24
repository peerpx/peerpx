package datastore

// Mocked is a mocked datastore for testing
type Mocked struct {
	responseErr  error
	responseData []byte
}

// InitMokedDatastore return a fake datastore for testing purpose
// note that the returned error is useless, but we add it to keep
// constant signature og datastore initialiser
func InitMokedDatastore(data []byte, err error) error {
	ds = &Mocked{
		responseData: data,
		responseErr:  err,
	}
	return nil
}

// Put implements datastore.put
func (d *Mocked) put(key string, value []byte) error {
	return d.responseErr
}

// exists implements datastore.exists
func (d *Mocked) exists(key string) (bool, error) {
	return d.responseData[0] == 1, d.responseErr
}

// Get implements datastore.get
func (d *Mocked) get(key string) (data []byte, err error) {
	return d.responseData, d.responseErr
}

// Delete implements datastore.delete
func (d *Mocked) delete(key string) error {
	return d.responseErr
}
