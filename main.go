package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	h := Handler{}

	var flags flag.FlagSet

	opts := protogen.Options{ParamFunc: flags.Set}
	opts.Run(h.Run)
}
