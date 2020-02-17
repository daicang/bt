package bdb

import "os"

type db struct {
}

func Open(path string, mode os.FileMode) (*db, error) {
	fd := os.OpenFile(path, os.O_CREATE, mode)
}
