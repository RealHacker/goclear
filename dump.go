/*
	Dump a variable in JSON format like this (VarDict)
	{
		name: "variable name",
		type: "specific variable type - mostly user defined",
		metatype: "type of type - map/slice/struct/ptr/int..."
		address: "address of variable, available for slice/array/struct/map",
		value: depending on type, could be a list/object with embeded variables:
			for a slice/array - a JSON list of VarDicts
			for a struct - a dict mapping from string (field name) to VarDicts
			for a map - a JSON list like [{key: key Vardict, value:value VarDict}]
			for a pointer - the VarDict of variable it points to
			for basic type - the value itself
			for a NULL pointer - "#NULL#"
			for a variable visited through another pointer - "#VISITED#"
			for a node too deeply recursed - "#DEPTH_EXCEEDED#""
	}
*/
package goclear

import "reflect"
import "encoding/json"
import "fmt"

func printLog(args ...interface{}) {
	// TODO: write some logging utility
	fmt.Println(args...)
}

type KeyValuePair map[string]interface{}

// Both setKey and setValue should receive a VarDict as parameter
func (kv KeyValuePair) setKey(obj interface{}) {
	kv["key"] = obj
}
func (kv KeyValuePair) setValue(obj interface{}) {
	kv["value"] = obj
}

type VarDict map[string]interface{}

func (dict VarDict) Dump() string {
	v, err := json.MarshalIndent(dict, "-", "  ")
	if err != nil {
		printLog("ERROR when marshalling into JSON:", dict)
		return nil
	}
	return string(v)
}

func (dict VarDict) SetName(val string) {
	dict["name"] = val
}

func (dict VarDict) SetType(val string) {
	dict["type"] = val
}

func (dict VarDict) SetMeta(val string) {
	dict["metatype"] = val
}

func (dict VarDict) SetAddress(val string) {
	dict["address"] = val
}

func (dict VarDict) SetValue(val interface{}) {
	dict["value"] = val
}

func (dict VarDict) SetField(name string, val interface{}) {
	dict[name] = val
}

// Provide a new blank VarDict
func NewVarDict() VarDict {
	dict := make(VarDict)
	dict["metatype"] = nil
	dict["value"] = nil
	return dict
}

// Do we need to wrap this kind in a Pointer for next level?
func needsPtrForKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Struct:
		return true
	default:
		return false
	}
}

// This cache serves 2 purpose:
// 1. Avoid entering a pointer loop when recursing
// 2. Avoid dumping an object already visited with pointer
var PointerCache []int64

// The entry point for generating a VarDict
// It determines whether to call GetVarDictFromPtr or GetVarDictFromValue
func GetVarDict(name string, obj interface{}) VarDict {
	// Get a clean pointer cache
	PointerCache = make([]int64, 32)

	vardict := GetVarDictFromValue(obj, 0)
	vardict.SetName(name)
	return vardict
}

// Generate a VarDict for the object itself, for basic types like int, string, and pointer
func GetVarDictFromValue(variable interface{}, depth int) VarDict {
	vardict := NewVarDict()

	if variable == nil {
		vardict.SetType("nil")
		vardict.SetMeta("nil")
		vardict.SetValue(nil)
		return vardict
	}
	if depth > Config.MaxDepth {
		vardict.SetType("depth")
		vardict.SetMeta("depth")
		vardict.SetValue("#DEPTH_EXCEEDED#")
		return vardict
	}
	v := reflect.ValueOf(variable)
	t := reflect.TypeOf(variable)
	vardict.SetType(t.Name())
	kind := v.Kind()

	switch kind {
	case reflect.Invalid:
		vardict.SetMeta("invalid")
		vardict.SetValue(nil)
	case reflect.Ptr:
		vardict.SetMeta("ptr")
		// No matter what the type is, we need to get VarDict from a pointer
		ptr := reflect.ValueOf(variable).Pointer()		
		if ptr != 0 {
			address := fmt.Sprintf("%d", ptr)
			// Check if the object pointed by ptr has been visited
			visited := false
			for _, p := range PointerCache {
				if p == int64(ptr) {
					visited = true
					break
				}
			}
			if !visited {
				// Put the pointer in for future checking
				PointerCache = append(PointerCache, int64(ptr))
				objval := reflect.ValueOf(variable).Elem()
				if objval.CanInterface(){			
					obj := objval.Interface()
					childVarDict := GetVarDictFromValue(obj, depth+1)
					childVarDict.SetAddress(address)
					vardict.SetValue(childVarDict)
				} else{
					vardict.SetValue("#NULL#")
				}
			} else{
				// Make a special VISITED vardict
				visitedVarDict := NewVarDict()
				visitedVarDict.SetAddress(address)
				visitedVarDict.SetMeta("visited")
				visitedVarDict.SetValue("#VISITED#")
				vardict.SetValue(visitedVarDict)
			}
		} else {
			vardict.SetValue("#NULL#")
		}
	case reflect.Array, reflect.Slice:
		if kind == reflect.Array {
			vardict.SetMeta("array")
		} else {
			vardict.SetMeta("slice")
		}
		// For an array, there is not easy way to covert obj to [len]interface{},
		// because the len is not a constant. We have to use reflection
		arraylen := v.Len()
		vardict.SetField("len", arraylen)
		vardict.SetField("cap", v.Cap())
		varDictArray := make([]VarDict, arraylen)
		for i := 0; i < arraylen; i++ {
			vi := v.Index(i)
			obji := vi.Interface()
			childvardict := GetVarDictFromValue(obji, depth+1)
			if needsPtrForKind(vi.Kind()) && vi.CanAddr() {
				address := fmt.Sprintf("%d", vi.Addr().Pointer())
				childvardict.SetAddress(address)
			}
			varDictArray[i] = childvardict
		}
		vardict.SetValue(varDictArray)
	case reflect.Map:
		vardict.SetMeta("map")
		// For map, we can convert to map[interface{}]interface{}
		// But let's try reflect first
		keys := v.MapKeys()
		vardict.SetField("len", len(keys))
		varDictArray := make([]KeyValuePair, len(keys))
		for i, key := range keys {
			kv := make(KeyValuePair)
			// Get key's VarDict
			keyobj := key.Interface()
			keyVarDict := GetVarDictFromValue(keyobj, depth+1)
			if needsPtrForKind(key.Kind()) && key.CanAddr() {
				address := fmt.Sprintf("%d", key.Addr().Pointer())
				keyVarDict.SetAddress(address)
			}
			kv.setKey(keyVarDict)
			// Get Value's VarDict
			value := v.MapIndex(key)
			valueobj := value.Interface()
			valueVarDict := GetVarDictFromValue(valueobj, depth+1)
			if needsPtrForKind(value.Kind()) && value.CanAddr() {
				address := fmt.Sprintf("%d", value.Addr().Pointer())
				valueVarDict.SetAddress(address)
			}
			kv.setValue(valueVarDict)
			varDictArray[i] = kv
		}
		vardict.SetValue(varDictArray)
	case reflect.Struct:
		vardict.SetMeta("struct")
		// For struct, use reflect
		varDictDict := make(map[string]interface{})
		numFields := t.NumField()
		for i := 0; i < numFields; i++ {
			fieldName := t.Field(i).Name
			value := v.FieldByName(fieldName)
			// Ignore all function/interface fields in a struct
			// if value.Kind() == reflect.Func || value.Kind() == reflect.Interface {
			// 	continue
			// }
			if value.CanInterface() {
				valueobj := value.Interface()
				valueVarDict := GetVarDictFromValue(valueobj, depth+1)
				if needsPtrForKind(value.Kind()) && value.CanAddr() {
					address := fmt.Sprintf("%d", value.Addr().Pointer())
					valueVarDict.SetAddress(address)
				}
				varDictDict[fieldName] = valueVarDict
			} else {
				varDictDict[fieldName] = "#UNEXPORTED#"
			}
		}
		vardict.SetValue(varDictDict)
	case reflect.Uintptr, reflect.UnsafePointer:
		vardict.SetMeta("unsafeptr")
		vardict.SetValue(fmt.Sprintf("%v", v.Interface()))
	case reflect.String:
		vardict.SetMeta("string")
		vardict.SetField("len", v.Len())
		vardict.SetValue(variable)
	case reflect.Bool:
		vardict.SetMeta("bool")
		vardict.SetValue(variable)
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		vardict.SetMeta("int")
		vardict.SetValue(variable)
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint:
		vardict.SetMeta("uint")
		vardict.SetValue(variable)
	case reflect.Float32, reflect.Float64:
		vardict.SetMeta("float")
		vardict.SetValue(variable)
	case reflect.Complex64, reflect.Complex128:
		vardict.SetMeta("complex")
		vardict.SetValue(fmt.Sprintf("%v", variable))
	case reflect.Interface:
		vardict.SetMeta("interface")
		vardict.SetValue("#INTERFACE#")
	case reflect.Func:
		vardict.SetMeta("function")
		vardict.SetType(v.Type().String())
		vardict.SetValue("#FUNCTION#")
	case reflect.Chan:
		vardict.SetMeta("chan")
		vardict.SetType(v.Type().String())
		vardict.SetValue("#CHANNEL#")
	default:
		vardict.SetMeta("unknown")
		if v.CanInterface() {
			vardict.SetValue(fmt.Sprintf("%v", v.Interface()))
		} else {
			vardict.SetValue(fmt.Sprintf("%v", v.String()))
		}
	}
	return vardict
}
