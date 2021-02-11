// +build gen

package main

import (
	"github.com/lu4p/binclude"
	"github.com/lu4p/binclude/bincludegen"
)

func main() {
	bincludegen.Generate(binclude.None, ".")
}
