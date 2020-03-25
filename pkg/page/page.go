package page

import "unsafe"

const (
	InternalPageFlag = 0x01
	LeafPageFlag     = 0x02
	MaxAllocSize     = 0xFFFFFFF
	maxPageIndex     = 0x7FFFFFF
)

type Pgid uint64

type Page struct {
	ID    Pgid
	flags uint16
	count uint16
	data  uintptr
}

type leafPageElement struct {
	offset  int
	keysz   int
	valuesz int
}

type innerPageElement struct {
	offset  int
	childID Pgid
	keysz   int
	valuesz int
}

func (p *Page) leafPageElement(index int) *leafPageElement {
	return &(*[maxPageIndex]leafPageElement)(unsafe.Pointer(&p.data))
}

func (p *Page) innerPageElement(index int) *innerPageElement {
	return &(*[maxPageIndex]innerPageElement)(unsafe.Pointer(&p.data))
}

func (e *leafPageElement) key() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.offset : e.offset+e.keysz]
}

func (e *leafPageElement) value() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.offset+e.keysz : e.offset+e.keysz+e.valuesz]
}

func (e *innerPageElement) key() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.offset : e.offset+e.keysz]
}

func (e *innerPageElement) value() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.offset+e.keysz : e.offset+e.keysz+e.valuesz]
}
