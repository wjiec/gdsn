package gdsn

import (
	"errors"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// DSN represents a parsed dsn
type DSN struct {
	*url.URL
}

// Address returns address for dsn
func (d *DSN) Address() string {
	switch d.Scheme {
	case "unix", "unixgram", "unixpacket":
		return d.Path
	}
	return d.Host
}

// bindError represents an invalid bind argument passed
type bindError struct {
	Type reflect.Type
}

func (e *bindError) Error() string {
	if e.Type == nil {
		return "bind: nil"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "bind: non-pointer"
	}

	return "bind: invalid"
}

// Bind bind dsn into structure by tag
func (d *DSN) Bind(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &bindError{Type: reflect.TypeOf(v)}
	}

	rv = reflect.Indirect(rv)
	if rv.Kind() != reflect.Struct {
		return &bindError{Type: rv.Type()}
	}

	port, _ := strconv.ParseInt(d.Port(), 10, 64)
	password, _ := d.User.Password()
	binders := map[string]binder{
		"scheme":   stringBinder(d.Scheme),
		"address":  stringBinder(d.Address()),
		"username": stringBinder(d.User.Username()),
		"password": stringBinder(password),
		"host":     stringBinder(d.Hostname()),
		"port":     integerBinder(port),
	}
	return d.bind(rv, binders)

}

// binder represents a value binder for field
type binder func(reflect.Value) error

// bind bind a DSN components into a struct
func (d *DSN) bind(rv reflect.Value, binders map[string]binder) error {
	return visitFields(rv, func(tag string, field reflect.Value) error {
		if binder, ok := binders[tag]; ok {
			if field.Type().Kind() == reflect.String {
				return binder(field)
			}
		}

		if strings.HasPrefix(tag, queryTag) {
			if err := d.bindQuery(field, tag); err != nil {
				return err
			}
		}

		return nil
	})
}

const (
	queryTag    = "query"
	queryPrefix = "query."
)

// bindQuery bind query parameters into a struct or slice
func (d *DSN) bindQuery(field reflect.Value, tag string) error {
	if !field.CanSet() || !field.CanAddr() {
		return nil
	}

	query := d.Query()
	rt := field.Type()
	if tag == queryTag && rt.Kind() == reflect.Struct {
		if field.CanAddr() {
			return visitFields(field, func(tag string, field reflect.Value) error {
				return anyBinder(query, tag)(field)
			})
		}
		return errors.New("bind: unaddrable query struct")
	}

	name := strings.TrimPrefix(tag, queryPrefix)
	return anyBinder(query, name)(field)
}

// stringBinder bind a string into a field
func stringBinder(value string) binder {
	return func(field reflect.Value) error {
		if field.CanSet() && field.CanAddr() {
			field.SetString(value)
		}
		return nil
	}
}

// integerBinder bind a integer into a field
func integerBinder(value int64) binder {
	return func(field reflect.Value) error {
		if field.CanSet() && field.CanAddr() {
			field.SetInt(value)
		}
		return nil
	}
}

// anyBinder bind any type into field
func anyBinder(query url.Values, tag string) binder {
	return func(field reflect.Value) error {
		switch field.Type().Kind() {
		case reflect.String:
			return stringBinder(query.Get(tag))(field)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n, err := strconv.ParseInt(query.Get(tag), 10, 64)
			if err != nil {
				return err
			}
			return integerBinder(n)(field)
		case reflect.Slice:
			if vals, ok := query[tag]; ok {
				slice := reflect.MakeSlice(field.Type(), len(vals), len(vals))
				for i, v := range vals {
					slice.Index(i).Set(reflect.ValueOf(v))
				}

				field.Set(slice)
			}
		}

		return nil
	}
}

type visitHandler func(tag string, field reflect.Value) error

func visitFields(rv reflect.Value, handler visitHandler) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		if tag, ok := rt.Field(i).Tag.Lookup("dsn"); ok {
			if err := handler(tag, rv.Field(i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Parse parse raw into a DSN value
func Parse(raw string) (*DSN, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	return &DSN{u}, nil
}
