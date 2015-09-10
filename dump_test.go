package goclear

import "testing"
import "fmt"

func TestDump(t *testing.T) {
	// basic types
	s := "Hello"
	ps := &s
	
	fmt.Println(string(GetVarDict("s", s).Dump()))
	fmt.Println(string(GetVarDict("ps", ps).Dump()))
	
	i := 123
	pi := &i
	fmt.Println(string(GetVarDict("i", i).Dump()))
	fmt.Println(string(GetVarDict("pi", pi).Dump()))

	type TT struct {
		SS string
		II *int
	}
	
	type EXAMPLE struct {
		Ex TT
		ExPtr *TT
	}
	tt:= TT{
		SS: "abcdef",
		II: pi,
	}	
	m := EXAMPLE {
		Ex: tt,
	}
	ptree := &m
	fmt.Println(ptree)
	m.ExPtr = &m.Ex

	
	fmt.Println(string(GetVarDict("tt", tt).Dump()))
	fmt.Println(string(GetVarDict("m", m).Dump()))

	arr := make([]TT, 3)
	arr[0] = TT{"sfsdd", pi}
	
	fmt.Println(string(GetVarDict("array", arr).Dump()))
}
