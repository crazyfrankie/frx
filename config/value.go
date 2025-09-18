package config

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Value is config value interface.
type Value interface {
	Bool() (bool, error)
	Int() (int64, error)
	Float() (float64, error)
	String() (string, error)
	Duration() (time.Duration, error)
	Slice() ([]Value, error)
	Map() (map[string]Value, error)
	Scan(any) error
	Load() any
	Store(any)
}

type value struct {
	atomic.Value
}

func (v *value) typeAssertError() error {
	return fmt.Errorf("type asset %v failed", reflect.TypeOf(v.Load()))
}

func (v *value) Bool() (bool, error) {
	switch val := v.Load().(type) {
	case bool:
		return val, nil
	case int, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return strconv.ParseBool(fmt.Sprint(val))
	case string:
		return strconv.ParseBool(val)
	}

	return false, v.typeAssertError()
}

func (v *value) Int() (int64, error) {
	switch val := v.Load().(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case int64:
		return val, nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	}
	return 0, v.typeAssertError()
}

func (v *value) Float() (float64, error) {
	switch val := v.Load().(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	}
	return 0, v.typeAssertError()
}

func (v *value) String() (string, error) {
	switch val := v.Load().(type) {
	case string:
		return val, nil
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprint(val), nil
	case []byte:
		return string(val), nil
	case fmt.Stringer:
		return val.String(), nil
	}
	return "", v.typeAssertError()
}

func (v *value) Duration() (time.Duration, error) {
	val, err := v.Int()
	if err != nil {
		return 0, err
	}
	return time.Duration(val), nil
}

func (v *value) Slice() ([]Value, error) {
	vals, ok := v.Load().([]any)
	if !ok {
		return nil, v.typeAssertError()
	}
	slices := make([]Value, 0, len(vals))
	for _, val := range vals {
		a := new(value)
		a.Store(val)
		slices = append(slices, a)
	}

	return slices, nil
}

func (v *value) Map() (map[string]Value, error) {
	vals, ok := v.Load().(map[string]any)
	if !ok {
		return nil, v.typeAssertError()
	}
	maps := make(map[string]Value, len(vals))
	for key, val := range vals {
		a := new(value)
		a.Store(val)
		maps[key] = a
	}

	return maps, nil
}

func (v *value) Scan(a any) error {
	data, err := json.Marshal(v.Load())
	if err != nil {
		return err
	}
	if val, ok := a.(proto.Message); ok {
		return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, val)
	}

	return json.Unmarshal(data, a)
}

type errValue struct {
	err error
}

func (v errValue) Bool() (bool, error)              { return false, v.err }
func (v errValue) Int() (int64, error)              { return 0, v.err }
func (v errValue) Float() (float64, error)          { return 0.0, v.err }
func (v errValue) Duration() (time.Duration, error) { return 0, v.err }
func (v errValue) String() (string, error)          { return "", v.err }
func (v errValue) Scan(any) error                   { return v.err }
func (v errValue) Load() any                        { return nil }
func (v errValue) Store(any)                        {}
func (v errValue) Slice() ([]Value, error)          { return nil, v.err }
func (v errValue) Map() (map[string]Value, error)   { return nil, v.err }
