14a15,17
> // This file is changed by Broadcom.
> // Modifications - Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or its subsidiaries.
> 
353c356,358
< 	util.DbgPrint("list after unmarshal:\n%s\n", pretty.Sprint(parent))
---
> 	if util.IsDebugLibraryEnabled() {
> 		util.DbgPrint("list after unmarshal:\n%s\n", pretty.Sprint(parent))	
> 	}
391d395
< 
397,398c401,405
< 
< 		nv, err := StringToType(ft, fieldVal)
---
> 		sf, ok := val.Elem().Type().FieldByName(fn)
> 		if ok == false {
> 			return fmt.Errorf("Field %s not present in the struct %s", fn, val.Elem())
> 		}
> 		cschema, err := childSchema(schema, sf)
401a409,485
> 		keyLeafKind := cschema.Type.Kind
> 		if keyLeafKind == yang.Yleafref {
> 			lrfschema, err := resolveLeafRef(cschema)
> 			if err != nil {
> 				return err
> 			}
> 			keyLeafKind = lrfschema.Type.Kind
> 		}
> 
> 		var nv reflect.Value
> 		if keyLeafKind == yang.Yunion {
> 			sks, err := getUnionKindsNotEnums(cschema)
> 			if err != nil {
> 				return err
> 			}
> 			for _, sk := range sks {
> 				gv, err := StringToType(reflect.TypeOf(yangBuiltinTypeToGoType(sk)), fieldVal)
> 				if err == nil {
> 					mn := "To_" + ft.Name()
> 					mapMethod := val.MethodByName(mn)
> 					if !mapMethod.IsValid() {
> 						return fmt.Errorf("%s does not have a %s function", val, mn)
> 					}
> 					ec := mapMethod.Call([]reflect.Value{gv})
> 					if len(ec) != 2 {
> 						return fmt.Errorf("%s %s function returns %d params", ft.Name(), mn, len(ec))
> 					}
> 					ei := ec[0].Interface()
> 					ee := ec[1].Interface()
> 					if ee != nil {
> 						return fmt.Errorf("unmarshaled %v type %T does not have a union type: %v", fieldVal, fieldVal, ee)
> 					}
> 					nv = reflect.ValueOf(ei)
> 					break
> 				}
> 			}
> 			
> 			if nv.IsValid() == false {
> 				ets, err := schemaToEnumTypes(cschema, elmT)
> 				if err != nil {
> 					return err
> 				}
> 				for _, et := range ets {
> 					ev, err := castToEnumValue(et, fieldVal)
> 					if err != nil {
> 						return err
> 					}
> 					if ev != nil {
> 						mn := "To_" + ft.Name()
> 						mapMethod := val.MethodByName(mn)
> 						if !mapMethod.IsValid() {
> 							return fmt.Errorf("%s does not have a %s function", val, mn)
> 						}
> 						ec := mapMethod.Call([]reflect.Value{reflect.ValueOf(ev)})
> 						if len(ec) != 2 {
> 							return fmt.Errorf("%s %s function returns %d params", ft.Name(), mn, len(ec))
> 						}
> 						ei := ec[0].Interface()
> 						ee := ec[1].Interface()
> 						if ee != nil {
> 							return fmt.Errorf("unmarshaled %v type %T does not have a union type: %v", fieldVal, fieldVal, ee)
> 						}
> 						nv = reflect.ValueOf(ei)
> 						break
> 					}
> 					fmt.Errorf("could not unmarshal %v into enum type: %s\n", fieldVal, err)
> 				}
> 				if nv.IsValid() == false {
> 					return fmt.Errorf("could not create the value type for the field name %s  with the value %s", fn, fieldVal)
> 				}
> 			}
> 		} else {
> 			nv, err = StringToType(ft, fieldVal)
> 			if err != nil {
> 				return err
> 			}
> 		}
496a581,583
>     if (len(keys) == 0) { 
>          return nil, nil
>     } 
