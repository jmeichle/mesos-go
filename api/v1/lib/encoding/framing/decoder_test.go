package framing

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	var (
		byteCopy = UnmarshalFunc(func(b []byte, m interface{}) error {
			if m == nil {
				return errors.New("unmarshal target may not be nil")
			}
			v, ok := m.(*[]byte)
			if !ok {
				return fmt.Errorf("expected *[]byte instead of %T", m)
			}
			if v == nil {
				return errors.New("target *[]byte may not be nil")
			}
			*v = append((*v)[:0], b...)
			return nil
		})
		fakeError        = errors.New("fake unmarshal error")
		errorUnmarshaler = UnmarshalFunc(func(_ []byte, _ interface{}) error {
			return fakeError
		})
		singletonReader = func(b []byte) ReaderFunc {
			eof := false
			return func() ([]byte, error) {
				if eof {
					panic("reader should only be called once")
				}
				eof = true
				return b, io.EOF
			}
		}
		errorReader = func(err error) ReaderFunc {
			return func() ([]byte, error) { return nil, err }
		}
	)
	for ti, tc := range []struct {
		r        Reader
		uf       UnmarshalFunc
		wants    []byte
		wantsErr error
	}{
		{errorReader(ErrorBadSize), byteCopy, nil, ErrorBadSize},
		{singletonReader(([]byte)("james")), byteCopy, ([]byte)("james"), io.EOF},
		{singletonReader(([]byte)("james")), errorUnmarshaler, nil, fakeError},
	} {
		var (
			buf []byte
			d   = NewDecoder(tc.r, tc.uf)
			err = d.Decode(&buf)
		)
		if err != tc.wantsErr {
			t.Errorf("test case %d failed: expected error %q instead of %q", ti, tc.wantsErr, err)
		}
		if !reflect.DeepEqual(buf, tc.wants) {
			t.Errorf("test case %d failed: expected %#v instead of %#v", ti, tc.wants, buf)
		}
	}
}
