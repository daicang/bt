package config

import "os"

var (
	PageSize = os.Getpagesize()
)
