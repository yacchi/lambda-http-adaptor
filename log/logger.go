package log

import "log"

type Writer func(...interface{})

var Info Writer = log.Println

var Warning Writer = log.Println
