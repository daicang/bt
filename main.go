package main

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/daicang/bt/pkg/page"
)

const (
	maxMapSize = 0xFFFFFFFFFFFF
)

var (
	defaultPageSize = os.Getpagesize()
)

type DB struct {
	path    string
	file    *os.File
	dataref []byte
	data    *[maxMapSize]byte
	datasz  int
	pagesz  int
}

func mmap(db *DB, sz int) error {
	b, err := syscall.Mmap(int(db.file.Fd()), 0, sz, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	db.dataref = b
	db.data = (*[maxMapSize]byte)(unsafe.Pointer(&b[0]))
	db.datasz = sz
	return nil
}

type meta struct {
}

func (db *DB) init() error {
	buf := make([]byte, defaultPageSize*4)
	for i := 0; i < 2; i++ {
		p := (*page.Page)(unsafe.Pointer(&buf[i*defaultPageSize]))
		p.ID = page.Pgid(i)
	}
	db.file.WriteAt()

}

func main() {
	var err error
	db := &DB{
		pagesz: defaultPageSize,
	}
	db.file, err = os.OpenFile(db.path, os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

}
