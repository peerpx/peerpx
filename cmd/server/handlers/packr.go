package handlers

import "github.com/gobuffalo/packr"

var tplBox packr.Box

func init() {
	tplBox = packr.NewBox("../templates")
}
