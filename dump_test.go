package goclear

import "testing"
import "fmt"
import  _ "reflect"


type TT struct {
		String string
		IntPtr *int
		TTPtr *TT
		Parent *EXAMPLE
}	
type EXAMPLE struct {
		Ex TT
		ExPtr *TT
		Func func(int)(int)
}
func Abc(a int) int {
	return a
}
func (a EXAMPLE) Dothings() int {
	return 1
}

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

	// struct
	tt:= TT{
		String: "abcdef",
		IntPtr: pi,
		TTPtr: nil,
	}
	// struct with function field and method
	m := EXAMPLE {
		Ex: tt,
		Func: Abc,
	}
	// Recursive structure
	tt.Parent = &m
	m.ExPtr = &tt

	fmt.Println(string(GetVarDict("tt", tt).Dump()))
	fmt.Println(string(GetVarDict("m", m).Dump()))

	// array/slice
	arr := make([]TT, 3)
	arr[0] = TT{"abc", pi, nil, nil}	
	arr[1] = TT{"xyz", pi, nil, &m}	
	fmt.Println(string(GetVarDict("array", arr).Dump()))

	// map
	dic := make(map[string]*TT)
	dic["abc"] = &TT{"ksajdf", pi, &tt, &m}
	dic["sdf"] = &TT{"sdfs", nil, nil, nil}
	fmt.Println(string(GetVarDict("dic", dic).Dump()))
	fmt.Println(string(GetVarDict("dicptr", &dic).Dump()))
}

