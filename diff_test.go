package goclear

import "testing"
import "fmt"

func _simple() {
	i1 := 4
	i2 := 5
	vd1 := GetVarDict("i", i1)
	vd2 := GetVarDict("i", i2)
	vd3 := GetVarDict("i", i2)

	fmt.Println(vd2.Compare(vd1), vd2)
	fmt.Println(vd3.Compare(vd2), vd3)
}

func _ptr() {
	s := "This is a string"
	s1 := "This is a string"

	p1 := &s
	p2 := &s1
	p3 := &s

	vd1 := GetVarDict("p", p1)
	vd2 := GetVarDict("p", p2)
	vd3 := GetVarDict("p", p3)

	fmt.Println(vd2.Compare(vd1), string(vd2.Dump()))
	fmt.Println(vd3.Compare(vd1), string(vd3.Dump()))

	p4 := &p1
	p5 := &p3
	p6 := &p3

	vd4 := GetVarDict("p", p4)
	vd5 := GetVarDict("p", p5)
	vd6 := GetVarDict("p", p6)
	
	fmt.Println(vd5.Compare(vd4), string(vd5.Dump()))
	fmt.Println(vd6.Compare(vd5), string(vd6.Dump()))
}

func _array() {
	a:= make([]int, 4)
	b:= make([]int, 4)
	a[1] = 1
	a[2] = 1
	a[3] = 2
	b[1] = 1
	b[2] = 1
	b[3] = 2
	vd1 := GetVarDict("a", a)
	vd2 := GetVarDict("a", b)
	fmt.Println(vd2.Compare(vd1), string(vd2.Dump()))
	aa := make([][]int, 2)
	aa[0] = a
	aa[1] = b
	vd3:= GetVarDict("aa", aa)
	fmt.Println(string(vd3.Dump()))
	b[2] = 33
	vd4:= GetVarDict("aa", aa)
	fmt.Println(vd4.Compare(vd3), string(vd4.Dump()))
}

func _map() {
	m := make(map[string]int)
	m["abc"]=1
	m["ds"]=2
	vd1:= GetVarDict("map", m)
	fmt.Println(string(vd1.Dump()))

	m["abc"] = 3
	m["dss"] = 4
	vd2:= GetVarDict("map", m)
	fmt.Println(vd2.Compare(vd1), string(vd2.Dump()))	

	m["abc"] = 1
	delete(m, "dss")
	vd3:= GetVarDict("map", m)
	fmt.Println(vd3.Compare(vd1), string(vd3.Dump()))	

}

func _struct(){
	type S struct{
		A string
		B int
		Pointer *S
	}

	s := S{"sdf", 123, nil}
	s.Pointer = &s
	vd1:= GetVarDict("struct", s)
	fmt.Println(string(vd1.Dump()))

	s.B = 123
	vd2:= GetVarDict("struct", s)
	fmt.Println(vd2.Compare(vd1), string(vd2.Dump()))
}

func TestCompare(t *testing.T){
	_simple()
	_ptr()
	_array()
	_map()
	_struct()
}
