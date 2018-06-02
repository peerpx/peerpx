package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type Basic struct {
	kv map[string]string
}

var (
	ErrNotFound = errors.New("key not found")
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
	conf = b
	return nil
}

// todo InitBasicFile

func (c Basic) set(key string, value interface{}) error {
	// cast to string
	c.kv[strings.ToLower(key)] = value.(string)
	return nil
}

func (c Basic) isSet(key string) (bool, error) {
	_, found := c.kv[strings.ToLower(key)]
	return found, nil
}

func (c Basic) getE(key string) (interface{}, error) {
	v, found := c.kv[strings.ToLower(key)]
	if !found {
		return nil, ErrNotFound
	}
	return v, nil
}

func (c Basic) getIntE(key string) (int, error) {
	v, err := c.getE(key)
	if err != nil {
		return 0, err
	}
	vInt, err := strconv.ParseInt(v.(string), 10, 64)
	if err != nil {
		vInt = 0
	}
	return int(vInt), err
}

func (c Basic) getFloat64E(key string) (float64, error) {
	v, err := c.getE(key)
	if err != nil {
		return 0, err
	}
	vF, err := strconv.ParseFloat(v.(string), 64)
	if err != nil {
		vF = 0
	}
	return vF, err
}

func (c Basic) getBoolE(key string) (bool, error) {
	v, err := c.getE(key)
	if err != nil {
		return false, err
	}
	vB, err := strconv.ParseBool(v.(string))
	if err != nil {
		vB = false
	}
	return vB, err
}

func (c Basic) getStringE(key string) (string, error) {
	v, err := c.getE(key)
	if err != nil {
		return "", err
	}
	return v.(string), err
}

func (c Basic) getStringSliceE(key string) ([]string, error) {
	v, err := c.getE(key)
	if err != nil {
		return nil, err
	}
	// string separator = ,
	parts := strings.Split(v.(string), ",")
	if len(parts) == 0 {
		return []string{}, nil
	}
	sl := make([]string, len(parts))
	for i, s := range parts {
		sl[i] = strings.TrimSpace(s)
	}
	return sl, nil
}

func (c Basic) getTime(key string) (time.Time, error) {
	v, err := c.getE(key)
	if err != nil {
		return time.Time{}, err
	}
	// in config time is represented by a int timestamp
	ts, err := strconv.ParseInt(v.(string), 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(ts, 0), nil
}

func (c Basic) getDuration(key string) (time.Duration, error) {
	v, err := c.getE(key)
	if err != nil {
		return time.Duration(0), err
	}
	// in config duration is represented by string
	return time.ParseDuration(v.(string))
}
