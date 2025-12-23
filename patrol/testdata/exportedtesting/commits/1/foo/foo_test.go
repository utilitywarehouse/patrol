package foo_test

import (
	"fmt"
	"testing"

	"github.com/shangardezi/exportedtesting/foo"
)

func TestFoo(t *testing.T) {
	fmt.Println(&foo.A{})
}
