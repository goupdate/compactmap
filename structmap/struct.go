package structmap

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/goupdate/compactmap"
)

/*
	Extended version of compactmap
*/

type StructMap[V any] struct {
	sync.RWMutex

	cm   *compactmap.CompactMap[int64, V]     // In-memory database
	info *compactmap.CompactMap[int64, int64] // Store maxId

	storageFile string

	maxId int64 // Max stored id, incremented after Add
}

// V - should be pointer to struct
func New[V any](storageFile string, failIfNotLoaded bool) (*StructMap[V], error) {
	var zero V
	valType := reflect.TypeOf(&zero).Elem()

	// Check if V is a pointer
	if valType.Kind() != reflect.Pointer {
		panic(fmt.Sprintf("cant use %v need pointer", valType.Name()))
	}

	// Check if V is a pointer to a struct
	structType := valType.Elem()
	if structType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("%v is not a struct", structType.Name()))
	}

	// Check if the struct has a field named "Id"
	id, ok := structType.FieldByName("Id")
	if !ok {
		panic(fmt.Sprintf("struct %v does not have a field named 'Id'", structType.Name()))
	}
	if id.Type.Kind() != reflect.Int64 {
		panic("Id should be int64")
	}

	cm := compactmap.NewCompactMap[int64, V]()
	err := cm.Load(storageFile)
	if err != nil && failIfNotLoaded {
		return nil, err
	}

	info := compactmap.NewCompactMap[int64, int64]()
	err = info.Load(storageFile + "i")
	if err != nil && failIfNotLoaded {
		return nil, err
	}

	maxId, ex := info.Get(1)
	if !ex {
		info.AddOrSet(1, 1)
		maxId = 1
	}

	return &StructMap[V]{cm: cm, maxId: maxId, info: info, storageFile: storageFile}, nil
}

// GetMaxId returns the current max ID
func (p *StructMap[V]) GetMaxId() int64 {
	return atomic.LoadInt64(&p.maxId)
}

// Save stores the current state of the map to a file
func (p *StructMap[V]) Save() error {
	return p.SaveAs(p.storageFile)
}

// Save stores the current state of the map to a file
func (p *StructMap[V]) SaveAs(name string) error {
	p.Lock()
	defer p.Unlock()

	p.info.AddOrSet(1, p.maxId)
	err := p.cm.Save(name)
	if err != nil {
		return err
	}
	err2 := p.info.Save(name + "i")
	if err2 != nil {
		return err2
	}
	return nil
}

// SetField sets a specific field to a value for a struct by ID
func (p *StructMap[V]) SetField(id int64, field string, value interface{}) bool {
	return p.SetFields(id, map[string]interface{}{field: value})
}

// SetFields sets multiple fields for a struct by ID
func (p *StructMap[V]) SetFields(id int64, fields map[string]interface{}) bool {
	p.Lock()
	defer p.Unlock()

	store, ex := p.cm.Get(id)
	if !ex {
		return false
	}
	val := reflect.Indirect(reflect.ValueOf(store))

	for field, value := range fields {
		//f := val.FieldByName(field)
		f := FindFieldByName(val, field)
		if !f.IsValid() || !f.CanSet() {
			return false
		}
		// Check and set the value with type conversion
		fieldType := f.Type()
		valueVal := reflect.ValueOf(value)

		if valueVal.Type().ConvertibleTo(fieldType) {
			f.Set(valueVal.Convert(fieldType))
		} else {
			panic(fmt.Sprintf("value of type %v is not assignable to type %v", valueVal.Type(), fieldType))
		}
	}
	return true
}

/*
	Update updates multiple fields for structs that match the given conditions

"condition" logic can be:

	"" - doesnt matter
	OR - where1 || where2 || where3 ...
	AND - where1 && where2 && where3

	Returns: elements updated
*/
func (p *StructMap[V]) Update(condition string, where []FindCondition, fields map[string]interface{}) int {
	ids := p.UpdateCount(condition, where, fields, 0)
	return len(ids)
}

/*
	Update updates multiple fields for structs that match the given conditions

condition logic can be:
"" - doesnt matter
OR - where1 || where2 || where3 ...
AND - where1 && where2 && where3

elCount - count of first elements to update, 0 if no limit

Returns: slice of Ids of updated elements
*/
func (p *StructMap[V]) UpdateCount(condition string, where []FindCondition, fields map[string]interface{}, elCount int) []int64 {
	var ids []int64
	count := 0
	p.FindFn(condition, where, func(id int64, v V) bool {
		ids = append(ids, id)
		count++
		if elCount > 0 && count == elCount {
			return false
		}
		return true
	})
	for _, id := range ids {
		p.SetFields(id, fields)
	}
	return ids
}

// GetAll retrieves all structs from the map
func (p *StructMap[V]) GetAll() []V {
	var ret []V
	p.cm.Iterate(func(k int64, v V) bool {
		ret = append(ret, v)
		return true
	})
	return ret
}

// compareValues compares two values based on the given operator
func compareValues(v1, v2 interface{}, op string) bool {
	v1Val := reflect.Indirect(reflect.ValueOf(v1))
	v2Val := reflect.Indirect(reflect.ValueOf(v2))

	// Handle nil values
	if !v1Val.IsValid() || !v2Val.IsValid() {
		switch op {
		case "equal", "eq", "=":
			return !v1Val.IsValid() && !v2Val.IsValid() // both are nil
		default:
			return false
		}
	}

	switch op {
	case "gt", "more", ">":
		if v1Val.Kind() == reflect.Int && v2Val.Kind() == reflect.Int {
			return v1Val.Int() > v2Val.Int()
		} else if v1Val.Kind() == reflect.Float32 || v1Val.Kind() == reflect.Float64 {
			return v1Val.Float() > v2Val.Float()
		} else if v1Val.Kind() == reflect.String && v2Val.Kind() == reflect.String {
			return v1Val.String() > v2Val.String()
		}
	case "lt", "less", "<":
		if v1Val.Kind() == reflect.Int && v2Val.Kind() == reflect.Int {
			return v1Val.Int() < v2Val.Int()
		} else if v1Val.Kind() == reflect.Float32 || v1Val.Kind() == reflect.Float64 {
			return v1Val.Float() < v2Val.Float()
		} else if v1Val.Kind() == reflect.String && v2Val.Kind() == reflect.String {
			return v1Val.String() < v2Val.String()
		}
	case "like", "contains":
		str1, ok1 := v1Val.Interface().(string)
		str2, ok2 := v2Val.Interface().(string)
		return ok1 && ok2 && strings.Contains(str1, str2)
	case "equal", "eq", "=":
		fallthrough
	default: // "="
		return reflect.DeepEqual(v1Val.Interface(), v2Val.Interface())
	}
	return false
}

// FindFieldByName searches for a field by name in a given value.
// It searches through the entire depth of nested structures.
func FindFieldByName(val reflect.Value, name string) reflect.Value {
	// Attempt to find the nested field first
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		return findNestedField(val, parts)
	}

	// If not nested, search for the direct field
	return findFieldByName(val, name)
}

// findNestedField recursively searches for a nested field by following the parts slice.
func findNestedField(val reflect.Value, parts []string) reflect.Value {
	if len(parts) == 0 {
		return val
	}

	currentPart := parts[0]
	remainingParts := parts[1:]

	field := findFieldByName(val, currentPart)
	if !field.IsValid() || len(remainingParts) == 0 {
		return field
	}

	return findNestedField(field, remainingParts)
}

// findFieldByName searches for a single field by name in a given value.
// If not found directly, it searches within nested structs.
func findFieldByName(val reflect.Value, name string) reflect.Value {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	valType := val.Type()
	if valType.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		fieldVal := val.Field(i)
		if strings.EqualFold(field.Name, name) {
			return fieldVal
		}

		// If the field is a nested struct, search within it
		if fieldVal.Kind() == reflect.Struct || (fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct) {
			nestedField := findFieldByName(fieldVal, name)
			if nestedField.IsValid() {
				return nestedField
			}
		}
	}

	return reflect.Value{}
}

type FindCondition struct {
	Field string
	Value interface{}
	Op    string // "equal", "eq", =, "gt", "more", >, "lt", "less", <
	// if op is "" eq not set, used "equal" operator
}

/*
condition logic can be:
"" - doesnt matter
OR - where1 || where2 || where3 ...
AND - where1 && where2 && where3
*/
func (p *StructMap[V]) Find(condition string, where ...FindCondition) []V {
	var ret []V
	p.FindFn(condition, where, func(key int64, v V) bool {
		ret = append(ret, v)
		return true
	})
	return ret
}

/*
Same as Find but with callback function for found elements
*/
func (p *StructMap[V]) FindFn(condition string, where []FindCondition, fn func(key int64, v V) bool) {
	p.cm.Iterate(func(key int64, v V) bool {
		match := false
		val := reflect.Indirect(reflect.ValueOf(v))

		for _, cond := range where {
			f := FindFieldByName(val, cond.Field)
			if !f.IsValid() {
				match = false
				break
			}

			// Determine if the condition is met
			isMatch := compareValues(f.Interface(), cond.Value, cond.Op)

			switch condition {
			case "AND":
				if !isMatch {
					match = false
					break
				}
				match = true
			case "OR":
				if isMatch {
					match = true
					break
				}
			default:
				match = isMatch
			}
		}

		if match {
			if !fn(key, v) {
				return false
			}
		}
		return true
	})
}

// Iterate iterates over all structs in the map and applies the given function
func (p *StructMap[V]) Iterate(fn func(v V) bool) {
	p.cm.Iterate(func(_ int64, v V) bool {
		return fn(v)
	})
}

// Get retrieves a struct by ID
func (p *StructMap[V]) Get(id int64) (V, bool) {
	return p.cm.Get(id)
}

// Delete removes a struct by ID
func (p *StructMap[V]) Delete(id int64) {
	p.cm.Delete(id)
}

// Clear removes all structs from the map and resets maxId
func (p *StructMap[V]) Clear() {
	p.maxId = 1
	p.cm.Clear()
}

// Add adds a new struct to the map, setting its ID if it is 0
func (p *StructMap[V]) Add(v V) int64 {
	if reflect.ValueOf(v).Kind() != reflect.Pointer {
		panic(fmt.Sprintf("cant add %v need pointer", reflect.ValueOf(v).Type().Name()))
	}
	val := reflect.ValueOf(v).Elem()
	idField := val.FieldByName("Id")

	if !idField.IsValid() || !idField.CanSet() {
		panic("struct does not have a settable Id field")
	}

	var id int64
	if idField.Int() == 0 {
		id = atomic.AddInt64(&p.maxId, 1)
		idField.SetInt(id)
	} else {
		id = idField.Int()
	}

	p.cm.AddOrSet(id, v)
	return id
}
