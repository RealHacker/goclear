/*
	Dump a variable in JSON format like this
	{
		varname: "variable name",
		typename: "variable type",
		metatype: "type of type - map/slice/struct/ptr"
		address: "address of variable, available for slice/array/struct/map",
		value: depending on type, could be a list/object with embeded variables:
			for a slice/array - a JSON list
			for a struct - a dict
			for a map - if the key is string, a dict
				else, a JSON list like [{key:keyvariable, value:valuevariable}]
			for a pointer - the variable it points to
			for basic type - the value itself
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

func (dict VarDict) Dump() []byte {
	v, err := json.MarshalIndent(dict, "-", "\t")
	if err != nil {
		printLog("ERROR when marshalling into JSON:", dict)
		return nil
	}
	return v
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
func NewVarDict() *VarDict {
	dict := make(VarDict)
	dict["type"] = nil
	dict["metatype"] = nil
	dict["value"] = nil
	return &dict
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

// The entry point for generating a VarDict
// It determines whether to call GetVarDictFromPtr or GetVarDictFromValue
func GetVarDict(name string, obj interface{}) *VarDict {
	vardict := GetVarDictFromValue(obj, 0)
	vardict.SetName(name)
	return vardict
}

// Generate a VarDict for the object ptr pointed to, also record the pointer address,
// 	call this function for struct/slice/map (complex type)
// func GetVarDictFromPtr(ptr interface{}, depth int) *VarDict {
// 	// vardict := NewVarDict()
// 	// Get the pointer address
// 	address := fmt.Sprintf("%d", reflect.ValueOf(ptr).Pointer())
// 	vardict.SetAddress(address)
// 	// Dereferencing the pointer
// 	v := reflect.ValueOf(ptr).Elem()
// 	t := v.Type()
// 	vardict.SetType(t.Name())
// 	kind := v.Kind()

// 	switch kind {

// 	default:
// 		// Cases where simple value also needs their ptr to be saved
// 		// For instance, when the simple value is referenced by a pointer
// 		// Call GetVarDictFromValue and copy the fields over
// 		_vardict := *GetVarDictFromValue(v.Interface(), depth)
// 		meta, ok := _vardict["metatype"].(string)
// 		//TODO: error handling
// 		if !ok {
// 			printLog("Error")
// 		}
// 		vardict.SetMeta(meta)
// 		vardict.SetValue(_vardict["value"])
// 	}
// 	return vardict
// }

// Generate a VarDict for the object itself, for basic types like int, string, and pointer
func GetVarDictFromValue(variable interface{}, depth int) *VarDict {
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
		vardict.SetValue("#DEPTH EXCEEDED#")
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
		address := fmt.Sprintf("%d", reflect.ValueOf(variable).Pointer())
		ptrval := reflect.ValueOf(variable)
		if ptrval != nil && ptrval.Elem().CanInterface() {
			objval := ptrval.Elem()
			obj := objval.Interface()
			childVarDict := GetVarDictFromValue(obj, depth+1)
			childVarDict.SetAddress(address)
			vardict.SetValue(childVarDict)
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
			childvardict := *GetVarDictFromValue(obji, depth+1)
			fmt.Println(vi.Kind())
			if needsPtrForKind(vi.Kind()) && vi.CanAddr() {

				address := fmt.Sprintf("%d", vi.Addr().Pointer())
				childvardict.SetAddress(address)
			}
			varDictArray[i] = childvardict
		}
		vardict.SetValue(varDictArray)
	// case reflect.Slice:
	// 	vardict.SetMeta("slice")
	// 	slicelen := v.Len()
	// 	vardict.SetField("len", slicelen)
	// 	vardict.SetField("cap", v.Cap())
	// 	varDictArray := make([]VarDict, slicelen)
	// 	// For slice, we will directly convert obj to []interface{}
	// 	for i, obji := range []interface{}(obj) {
	// 		if needsPtrForKind(reflect.ValueOf(obji).Kind()) {
	// 			varDictArray[i] = *GetVarDictFromPtr(&obji, depth+1)
	// 		} else {
	// 			varDictArray[i] = *GetVarDictFromValue(obji, depth+1)
	// 		}
	// 	}
	// 	vardict.SetValue(varDictArray)
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
			// if needsPtrForKind(key.Kind()) {
			// 	address := fmt.Sprintf("%d", reflect.ValueOf(&keyobj).Pointer())
			// 	keyVarDict.SetAddress(address)
			// }
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
			valueobj := value.Interface()
			valueVarDict := GetVarDictFromValue(valueobj, depth+1)
			fmt.Println("field"+fieldName, value.CanAddr())
			if needsPtrForKind(value.Kind()) && value.CanAddr() {
				address := fmt.Sprintf("%d", value.Addr().Pointer())
				valueVarDict.SetAddress(address)
			}
			varDictDict[fieldName] = valueVarDict
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
		vardict.SetValue("#COMPLEX#")
	case reflect.Interface:
		vardict.SetMeta("interface")
		vardict.SetValue("#INTERFACE#")
	case reflect.Func:
		vardict.SetMeta("function")
		vardict.SetValue("#FUNCTION#")
	case reflect.Chan:
		vardict.SetMeta("chan")
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
