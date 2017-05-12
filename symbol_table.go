package main

type SymbolTable struct {
	table map[string]float64
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		table: make(map[string]float64),
	}
}

func (s *SymbolTable) Set(name string, value float64) {
	s.table[name] = value
}

func (s *SymbolTable) Get(name string) (float64, bool) {
	value, found := s.table[name]
	return value, found
}
