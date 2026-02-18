package main

import "time"

// timeNow is a seam for testing. Production code calls time.Now().
var timeNow = time.Now
