package cache

type basic struct {
	store map[string][]byte
}

// InitBasicCache
func InitBasicCache() error {
	b := basic{
		store: make(map[string][]byte),
	}
	cache = b
	return nil
}

func (b basic) get(key string) ([]byte, error) {
	v, ok := b.store[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (b basic) set(key string, value []byte) error {
	b.store[key] = value
	return nil
}

func (b basic) del(key string) error {
	delete(b.store, key)
	return nil
}
