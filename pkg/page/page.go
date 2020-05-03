package page

import "unsafe"

const (
	InternalPageFlag     = 0x01
	LeafPageFlag         = 0x02
	MaxAllocSize         = 0xFFFFFFF
	maxPageIndex         = 0x7FFFFFF
	InnerPageElementSize = int(unsafe.Sizeof(InnerPageElement{}))
	LeafPageElementSize  = int(unsafe.Sizeof(LeafPageElement{}))
)

type Pgid uint64

type Page struct {
	ID    Pgid
	Flags uint16
	Count uint16
	Data  uintptr
}

type LeafPageElement struct {
	Offset  uint32
	Keysz   uint32
	Valuesz uint32
}

type InnerPageElement struct {
	Offset  uint32
	Keysz   uint32
	Valuesz uint32
	ChildID Pgid
}

func (p *Page) LeafPageElement(index uint16) *LeafPageElement {
	return &((*[maxPageIndex]LeafPageElement)(unsafe.Pointer(&p.Data)))[index]
}

func (p *Page) InnerPageElement(index uint16) *InnerPageElement {
	return &((*[maxPageIndex]InnerPageElement)(unsafe.Pointer(&p.Data))[index])
}

func (e *LeafPageElement) Key() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.Offset : e.Offset+e.Keysz]
}

func (e *LeafPageElement) Value() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.Offset+e.Keysz : e.Offset+e.Keysz+e.Valuesz]
}

func (e *InnerPageElement) Key() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.Offset : e.Offset+e.Keysz]
}

func (e *InnerPageElement) Value() []byte {
	buf := (*[MaxAllocSize]byte)(unsafe.Pointer(&e))
	return buf[e.Offset+e.Keysz : e.Offset+e.Keysz+e.Valuesz]
}
