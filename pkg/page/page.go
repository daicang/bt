package page

import (
	"os"
)

const (
	InternalPageFlag = 0x01
	LeafPageFlag     = 0x02
	MaxAllocSize     = 0xFFFFFFF
)

var defaultPageSize = os.Getpagesize()

type pgid uint64

type Page struct {
	id    pgid
	flags uint16
	count uint16
	ptr   uintptr
}

type


func
