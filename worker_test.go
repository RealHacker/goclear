package goclear

import "testing"


func TestWorker(t *testing.T) {
	i1 := 4
	vd1 := GetVarDict("i", i1)

	PostRecord(&vd1)

	s := "This is a string"
	p1 := &s
	vd2 := GetVarDict("s", s)
	vd3 := GetVarDict("p", p1)
	PostRecord(&vd2)
	PostRecord(&vd3)

	a:= make([]int, 4)
	a[1] = 1
	a[2] = 1
	a[3] = 2
	vd4 := GetVarDict("a", a)
	PostRecord(&vd4)

	m := make(map[string]int)
	m["abc"]=1
	m["def"]=2
	vd5 := GetVarDict("map", m)
	PostRecord(&vd5)

	type S struct{
		A string
		B int
		Pointer *S
	}

	s := S{"sdf", 123, nil}
	s.Pointer = &s
	vd6:= GetVarDict("struct", s)
	PostRecord(&vd6)

	Finish()
}