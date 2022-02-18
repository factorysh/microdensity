package gojatest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dop251/goja"
)

type Test struct {
	vm *goja.Runtime
}

func New(file, test string) (*Test, error) {
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
		if !a {
			throw("False");
		}
    }
};
`))

	vm := goja.New()
	_, err := vm.RunString(src.String())
	if err != nil {
		return nil, err
	}
	return &Test{
		vm: vm,
	}, nil
}

func (t *Test) Tests() []string {
	tt := make([]string, 0)
	for _, k := range t.vm.GlobalObject().Keys() {
		if strings.HasPrefix(k, "test") {
			tt = append(tt, k)
		}
	}
	return tt
}

func (t *Test) Run(test string) error {
	fun, ok := goja.AssertFunction(t.vm.Get(test))
	if !ok {
		return fmt.Errorf("test not found : %s", test)
	}
	_, err := fun(goja.Null())
	return err
}

func (t *Test) RunAll() {
	for _, tt := range t.Tests() {
		err := t.Run(tt)
		fmt.Println(err)
	}
}
