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

type Pgid uint64

type Page struct {
	ID    Pgid
	flags uint16
	count uint16
	ptr   uintptr
}
