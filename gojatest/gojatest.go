package gojatest

import (
	"bytes"
	"io"
	"os"

	"github.com/dop251/goja"
)

func ReadAll(file, test string) (*goja.Runtime, error) {
	src := &bytes.Buffer{}

	for _, i := range []string{file, test} {
		f, err := os.Open(i)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(src, f)
		if err != nil {
			return nil, err
		}
		f.Close()
		src.Write([]byte("\n\n"))
	}
	src.Write([]byte(`
let assert = {
    throw: (cb) => {

    },
    that: (a) => {

    }
};
	
`))

	vm := goja.New()
	_, err := vm.RunString(src.String())
	if err != nil {
		return nil, err
	}
	return vm, nil
}
