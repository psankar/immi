package idb

import "immi/pkg/dao"

type pg struct {
}

func NewPGDB() (*pg, error) {
	return &pg{}, nil
}

func (pg *pg) AppendImmis(immis []dao.Immi) error {
	return nil
}
