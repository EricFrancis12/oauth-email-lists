package main

import (
	"strings"
)

const (
	strIpolLeftDelim  string = "{{"
	strIpolRightDelim string = "}}"
)

type StrIpol struct {
	data       map[string]string
	leftDelim  string
	rightDelim string
}

func NewStrIpol(leftDelim string, rightDelim string) *StrIpol {
	return &StrIpol{
		data:       make(map[string]string),
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
	}
}

func (s *StrIpol) RegisterVar(name string, value string) {
	s.data[strings.Trim(name, " ")] = value
}

func (s *StrIpol) RegisterVars(data map[string]string) {
	for name, value := range data {
		s.RegisterVar(name, value)
	}
}

func (s *StrIpol) Eval(str string) string {
	partsA := strings.Split(str, s.leftDelim)
	for i, a := range partsA {
		if strings.Contains(a, s.rightDelim) {
			partsB := strings.Split(a, s.rightDelim)
			trimmed := strings.Trim(partsB[0], " ")
			value, ok := s.data[trimmed]
			if ok {
				partsB[0] = value
			} else {
				partsB[0] = ""
			}
			partsA[i] = strings.Join(partsB, "")
		}
	}
	return strings.Join(partsA, "")
}

func (to TelegramOutput) StrIpolMap(emailAddr string, name string) map[string]string {
	return map[string]string{
		"emailAddr": emailAddr,
		"name":      name,
	}
}
