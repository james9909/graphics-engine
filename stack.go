package main

import (
	"bytes"
)

type Stack struct {
	stack []*Matrix
}

func NewStack() *Stack {
	return &Stack{
		stack: make([]*Matrix, 0, 100),
	}
}

func (s *Stack) Pop() *Matrix {
	if s.IsEmpty() {
		return nil
	}
	length := len(s.stack)
	ret := s.stack[length-1]
	s.stack = s.stack[:length-1]
	return ret
}

func (s *Stack) Push(m *Matrix) {
	s.stack = append(s.stack, m)
}

func (s *Stack) Peek() *Matrix {
	if s.IsEmpty() {
		return nil
	}
	length := len(s.stack)
	return s.stack[length-1]
}

func (s *Stack) IsEmpty() bool {
	return len(s.stack) == 0
}

func (s *Stack) String() string {
	var buffer bytes.Buffer
	length := len(s.stack)
	for i := length - 1; i >= 0; i-- {
		buffer.WriteString(s.stack[i].String())
	}
	return buffer.String()
}
