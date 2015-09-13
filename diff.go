package goclear

func getUnchangedVarDict() VarDict {
	dict := make(VarDict)
	dict["metatype"] = "unchanged"
	dict["value"] = nil
	return dict
}

func getValueType(v interface{}) string{
	if v == nil {
		return "nil"
	}
	var t string
	switch v.(type) {
	case string:
		t = "string"
	case VarDict:
		t = "VarDict"
	case []VarDict:
		t = "array"
	case []KeyValuePair:
		t = "map"
	case map[string]interface{}:
		t = "struct"
	default:
		t = "number"
	}
	return t
}

func minInt (len1 int, len2 int) int {
	if len1<=len2{
		return len1
	}else{
		return len2
	}
}

// Compare vardict and the last one recursively, return whether they are exactly the same
// If Exactly the same, the caller should replace this vardict with a "unchanged" vardict
// In effect, this prunes the vardict, preserving only the diff
func (vardict VarDict) Compare(last VarDict) bool{
	// First make sure metatype and type are the same
	if vardict["metatype"] != last["metatype"] || vardict["type"] != last["type"] {
		return false
	}
	// Get the types of the 2 vardict's value field
	t1 := getValueType(vardict["value"])
	t2 := getValueType(last["value"])
	if t1 != t2 {
		return false
	}
	// Compare address: if both have same address or both lack address, ok, else, return false
	addr1, ok1 := vardict["address"]
	addr2, ok2 := last["address"]
	if ok1 != ok2 || addr1 != addr2 {
		return false
	}
	// Deal with various types
	switch vardict["metatype"] {
	case "ptr":
		if t1 == "string" { // 2 null pointers
			return false
		}
		// Now both values have to be VarDict
		ptrVarDict1 := vardict["value"].(VarDict)	
		ptrVarDict2 := last["value"].(VarDict)
		return ptrVarDict1.Compare(ptrVarDict2) 
	case "array", "slice":
		len1 := vardict["len"].(int)
		len2 := last["len"].(int)
		minlen := minInt(len1, len2)
		children1 := vardict["value"].([]VarDict)
		children2 := last["value"].([]VarDict)
		allSame := true
		for i:=0; i < minlen; i++{
			if !children1[i].Compare(children2[i]) {
				allSame = false
			} else{
				children1[i] = getUnchangedVarDict()
			}
		}
		if allSame && len1==len2 {
			return true
		}
		return false
	case "map":
		len1 := vardict["len"].(int)
		len2 := last["len"].(int)
		minlen := minInt(len1, len2)
		children1 := vardict["value"].([]KeyValuePair)
		children2 := last["value"].([]KeyValuePair)
		allSame := true
		for i:=0; i < minlen; i++{
			key1 := children1[i]["key"].(VarDict)
			value1 := children1[i]["value"].(VarDict)
			key2 := children2[i]["key"].(VarDict)
			value2 := children2[i]["value"].(VarDict)
			if !key1.Compare(key2) {
				allSame = false
			} else{
				children1[i]["key"] = getUnchangedVarDict()
			}
			if !value1.Compare(value2) {
				allSame = false
			} else {
				children1[i]["value"] = getUnchangedVarDict()
			}
		}
		if allSame && len1 == len2 {
			return true
		}
		return false
	case "struct":
		map1 := vardict["value"].(map[string]interface{})
		map2 := last["value"].(map[string]interface{})
		len1, len2 := len(map1), len(map2)
		allSame := true
		for k, v1 := range map1 {
			v2, exists := map2[k]
			if !exists {
				allSame = false
			} else {
				tt1, tt2 := getValueType(v1), getValueType(v2)
				if tt1 != tt2 {
					allSame = false
				} else if tt1=="VarDict"{
					vv1 := v1.(VarDict)
					vv2 := v2.(VarDict)
					if !vv1.Compare(vv2){
						allSame = false
					} else {
						map1[k] = getUnchangedVarDict()
					}
				} else if v1.(string)!= v2.(string){
					allSame = false
				}
			}
		}
		if allSame && len1 == len2 {
			return true
		}
		return false
	default:
		return vardict["value"]==last["value"]
	}

}