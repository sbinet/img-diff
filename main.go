// Copyright 2021 The img-diff Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gioui.org/app"
)

func main() {
	log.SetPrefix("img-diff: ")
	log.SetFlags(0)

	var (
		batch = flag.Bool("batch", false, "enable batch mode")
		diff  = flag.Float64("max", 0.1, "maximum allowed difference in batch mode")
	)
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		log.Fatalf("missing input image(s)")
	}

	img1, err := loadImage(flag.Arg(0))
	if err != nil {
		log.Fatalf("could not load image %q: %+v", flag.Arg(0), err)
	}
	img2, err := loadImage(flag.Arg(1))
	if err != nil {
		log.Fatalf("could not load image %q: %+v", flag.Arg(1), err)
	}

	gui := NewUI(img1, img2)
	if *batch {
		fmt.Printf("diff=[%g, %g]\n", gui.dmin, gui.dmax)
		switch {
		case gui.dmax > *diff:
			os.Exit(1)
		default:
			os.Exit(0)
		}
	}

	go gui.run()

	app.Main()
}
