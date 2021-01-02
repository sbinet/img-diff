// Copyright 2021 The img-diff Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"

	"gioui.org/app"
)

func main() {
	log.SetPrefix("img-diff: ")
	log.SetFlags(0)

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
	go gui.run()

	app.Main()
}
