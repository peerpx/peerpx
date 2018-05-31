package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Basic struct {
	kv map[string]string
}

var (
	ErrNotFound        = errors.New("key not found")
	ErrIncomptibleType = errors.New("incompatible type")
)

// InitBasicConfig initialize a config with a basic scheme
func InitBasicConfig(r io.Reader) error {
	/*fd, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("unable to open conf file %s: %v", path, err)
	}
	defer fd.Close()
	*/
	b := Basic{
		kv: make(map[string]string),
	}

	// scan file
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// comments
		if line[0] == 35 {
			continue
		}
		keyValue := strings.SplitAfterN(line, ":", 2)
		if len(keyValue) != 2 {
			return fmt.Errorf("bad syntax found in config file for line: %s", line)
		}
		b.kv[strings.ToLower(keyValue[0][:len(keyValue[0])-1])] = strings.TrimSpace(keyValue[1])
	}
	//log.Printf("KEYVALUE %v", b.kv)
	conf = b
	return nil
}

func (c Basic) set(key string, value interface{}) error {
	// cast to string
	c.kv[strings.ToLower(key)] = value.(string)
	return nil
}

func (c Basic) get(key string) interface{} {
	v := c.kv[strings.ToLower(key)]
	return v
}

func (c Basic) getOrPanic(key string) interface{} {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		panic(ErrNotFound)
	}
	return v
}

func (c Basic) getInt(key string) int {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		return 0
	}
	vInt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		vInt = 0
	}
	return int(vInt)
}

func (c Basic) getIntOrPanic(key string) int {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		panic(ErrNotFound)
	}
	vInt, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(err)
	}
	return int(vInt)
}

func (c Basic) getFloat64(key string) float64 {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		return 0
	}
	vFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		vFloat = 0
	}
	return vFloat
}

func (c Basic) getFloat64OrPanic(key string) float64 {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		panic(ErrNotFound)
	}
	vFloat, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(err)
	}
	return vFloat
}

func (c Basic) getString(key string) string {
	i := c.get(key)
	return i.(string)
}

func (c Basic) getStringOrPanic(key string) string {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		panic(ErrNotFound)
	}
	return v
}
