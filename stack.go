package main

import (
	"bytes"
)

// Stack is a stack of matrices
type Stack struct {
	stack []*Matrix
}

// NewStack returns a new stack
func NewStack() *Stack {
	return &Stack{
		stack: make([]*Matrix, 0, 10),
	}
}

// Pop returns and removes the top matrix in the stack
func (s *Stack) Pop() *Matrix {
	if s.IsEmpty() {
		return nil
	}
	length := len(s.stack)
	ret := s.stack[length-1]
	s.stack = s.stack[:length-1]
	return ret
}

// Push pushes a new matrix onto the stack
func (s *Stack) Push(m *Matrix) {
	s.stack = append(s.stack, m)
}

// Peek returns the top matrix in the stack
func (s *Stack) Peek() *Matrix {
	if s.IsEmpty() {
		return nil
	}
	length := len(s.stack)
	return s.stack[length-1]
}

// IsEmpty returns true if the stack is empty, false otherwise
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
