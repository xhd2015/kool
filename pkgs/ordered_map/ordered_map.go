package ordered_map

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Option func(cfg *mapConfig)

// mapConfig holds configuration for parsing
type mapConfig struct {
	capacity int
}

// OrderedMap is a map that maintains the order of the keys
// it is useful for encoding and decoding JSON
type OrderedMap struct {
	keys []string
	m    map[string]interface{}
}

type Map interface {
	Set(key string, val interface{})
	Get(key string) interface{}
	Has(key string) bool
	GetOK(key string) (val interface{}, ok bool)
	Range(fn func(key string, val interface{}) bool)
}

func NewOrderedMap(opts ...Option) Map {
	return newOrderedMap(opts...)
}

func newOrderedMap(opts ...Option) *OrderedMap {
	cfg := &mapConfig{
		capacity: 0,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	n := cfg.capacity
	return &OrderedMap{
		keys: make([]string, 0, n),
		m:    make(map[string]interface{}, n),
	}
}

func WithCapacity(capacity int) Option {
	return func(cfg *mapConfig) {
		cfg.capacity = capacity
	}
}

func (c *OrderedMap) Set(key string, val interface{}) {
	if _, ok := c.m[key]; ok {
		idx := -1
		for i, k := range c.keys {
			if k == key {
				idx = i
				break
			}
		}
		if idx < 0 {
			panic(fmt.Errorf("inconsistent key not found:%v", key))
		}
		keys := c.keys
		c.keys = make([]string, 0, len(c.keys))
		c.keys = append(c.keys, keys[0:idx]...)
		if idx+1 < len(keys) {
			c.keys = append(c.keys, keys[idx+1:]...)
		}
		c.m[key] = val
		return
	}
	c.keys = append(c.keys, key)
	c.m[key] = val
}

func (c *OrderedMap) Add(key string, val interface{}) {
	if _, ok := c.m[key]; ok {
		panic(fmt.Errorf("key already exists:%s", key))
	}
	c.keys = append(c.keys, key)
	c.m[key] = val
}

func (c *OrderedMap) Get(key string) interface{} {
	return c.m[key]
}

func (c *OrderedMap) Has(key string) bool {
	_, ok := c.m[key]
	return ok
}

func (c *OrderedMap) GetOK(key string) (val interface{}, ok bool) {
	val, ok = c.m[key]
	return
}

func (c *OrderedMap) Range(fn func(key string, val interface{}) bool) {
	for _, k := range c.keys {
		if !fn(k, c.m[k]) {
			return
		}
	}
}

func (c *OrderedMap) UnmarshalJSON(data []byte) error {
	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)

	// expecting {
	tok, err := decoder.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		// nil
		return nil
	}
	if t, ok := tok.(json.Delim); !ok || t != json.Delim('{') {
		return fmt.Errorf("expecting '{'")
	}

	*c = *newOrderedMap()
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if _, ok := token.(json.Delim); ok {
			return fmt.Errorf("expect key,found:%v", token)
		}

		key := fmt.Sprint(token)

		// fmt.Printf("found key:%v, len(data) = %v, reader.Len()=%v, buf len=%v\n", key, len(data), reader.Len(), decoder.Buffered().(*bytes.Reader).Len())

		upIdx := len(data) - reader.Len() - decoder.Buffered().(*bytes.Reader).Len()

		// subData :=
		subData := data[upIdx:]
		idx, ch := nextNonSpace(subData)
		if idx < 0 || ch != ':' {
			return fmt.Errorf("expecting ':',found:'%v'", string(ch))
		}
		idx, ch = nextNonSpace(subData[idx+1:])
		var val interface{}
		// if {}, use a sorted map
		if ch == '{' {
			smap := NewOrderedMap()
			err = decoder.Decode(smap)
			val = smap
		} else {
			err = decoder.Decode(&val)
		}

		if err != nil {
			return err
		}
		c.Set(key, val)
	}
	tok, err = decoder.Token()
	if err != nil {
		return err
	}
	if t, ok := tok.(json.Delim); !ok || t != json.Delim('}') {
		return fmt.Errorf("expecting '{'")
	}
	if decoder.More() {
		return fmt.Errorf("invalid JSON")
	}
	return nil
}
func nextNonSpace(data []byte) (int, byte) {
	for i := 0; i < len(data); i++ {
		if data[i] == ' ' || data[i] == '\t' || data[i] == '\n' {
			continue
		}
		return i, data[i]
	}
	return -1, 0
}

func (c *OrderedMap) MarshalJSON() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	buf.WriteByte('{')
	n := len(c.keys)
	for i, k := range c.keys {
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}

		valBytes, err := json.Marshal(c.m[k])
		if err != nil {
			return nil, err
		}

		// val := c.m[k]
		// rval := reflect.ValueOf(val)

		// ok := true
		// switch rval.Kind() {
		// case reflect.Func, reflect.Chan:
		// 	ok = false
		// case reflect.Map:
		// 	if rval.Type().Elem().Kind() == reflect.Interface && rval.Type().Elem().NumMethod() == 0 {
		// 		ok = false
		// 	}
		// }
		// var valBytes []byte
		// if ok {
		// 	var err error
		// 	valBytes, err = json.Marshal(val)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// } else {
		// 	valBytes = []byte("null")
		// }

		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(valBytes)
		if i < n-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
