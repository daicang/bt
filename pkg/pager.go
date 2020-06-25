package page

import (
	db "github.com/daicang/bt/pkg/bt"
)

type Pager struct {
	db    *db.DB
	pager map[Pgid]*Page
}

func (p *Pager) page(id Pgid) *Page {
	if p, ok := p.pager[id]; ok {
		return p
	}
	return p.db.GetPage(id)
}
