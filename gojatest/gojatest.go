package gojatest

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
)

type Test struct {
	program *goja.Program
}

func New(file, test string) (*Test, error) {
	ffile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer ffile.Close()
	ftest, err := os.Open(test)
	if err != nil {
		return nil, err
	}
	defer ftest.Close()
	ast, err := parser.ParseFile(nil, "testing.js",
		io.MultiReader(ffile, ftest, strings.NewReader(`
const assert = {
    throw: (cb) => {
		let catched = false;
		let ee = null;
		try {
			cb()
		} catch (e) {
			catched = true;
			ee = e;
		}
		if (!catched) {
			throw("This test should throw error");
		}
		return ee;
    },
    that: (a, txt) => {
		if (!a) {
			throw("Failed : " + txt);
		}
    }
};
`)), 0)
	if err != nil {
		return nil, err
	}
	prg, err := goja.CompileAST(ast, true)
	if err != nil {
		return nil, err
	}
	return &Test{
		program: prg,
	}, nil
}

func (t *Test) Tests() ([]string, error) {
	vm := goja.New()
	_, err := vm.RunProgram(t.program)
	if err != nil {
		return nil, err
	}
	tt := make([]string, 0)
	for _, k := range vm.GlobalObject().Keys() {
		if strings.HasPrefix(k, "test") {
			tt = append(tt, k)
		}
	}
	if r := recover(); r != nil {
		fmt.Println("Recovering from panic:", r)
	}
	sort.Strings(tt)

	return tt, nil
}

func (t *Test) Run(vm *goja.Runtime, test string) error {
	fun, ok := goja.AssertFunction(vm.Get(test))
	if !ok {
		return fmt.Errorf("test not found : %s", test)
	}
	_, err := fun(goja.Null())
	return err
}

func (t *Test) RunAll() error {
	vm := goja.New()
	_, err := vm.RunProgram(t.program)
	if err != nil {
		return err
	}
	tests, err := t.Tests()
	if err != nil {
		return err
	}
	for _, tt := range tests {
		err := t.Run(vm, tt)
		if err != nil {
			return err
		}
	}
	return nil
}
