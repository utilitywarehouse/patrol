package foo_test

import (
	"fmt"
	"testing"

	"github.com/utilitywarehouse/exportedtesting/foo"
)

func TestFoo(t *testing.T) {
	fmt.Println(&foo.A{})
}
