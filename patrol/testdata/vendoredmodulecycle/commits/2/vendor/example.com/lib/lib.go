// Package lib is the root package of the example.com/lib module, and it
// imports one of the module's own subpackages.
package lib

import "example.com/lib/sub"

func Hello() string {
	return "hello " + sub.Name()
}
