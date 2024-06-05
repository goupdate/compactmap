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
	return p.SaveAs(p.storageFile + "i")
}

// Save stores the current state of the map to a file
func (p *StructMap[V]) SaveAs(name string) error {
	p.Lock()
	defer p.Unlock()

	p.info.AddOrSet(1, p.maxId)
	err := p.cm.Save(p.storageFile)
	if err != nil {
		return err
	}
	err2 := p.info.Save(name)
	if err2 != nil {
		return err2
	}
	return nil
}

// SetField sets a specific field to a value for a struct by ID
func (p *StructMap[V]) SetField(id int64, field string, value interface{}) bool {
	p.Lock()
	defer p.Unlock()

	store, ex := p.cm.Get(id)
	if !ex {
		return false
	}
	val := reflect.Indirect(reflect.ValueOf(store))
	//f := val.FieldByName(field)
	f := FindFieldByName(val, field)
	if !f.IsValid() || !f.CanSet() {
		return false
	}
	f.Set(reflect.ValueOf(value))
	return true
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

// Update updates multiple fields for structs that match the given conditions
/*
condition logic can be:
"" - doesnt matter
OR - where1 || where2 || where3 ...
AND - where1 && where2 && where3
*/
func (p *StructMap[V]) Update(condition string, where []FindCondition, fields map[string]interface{}) int {
	elems := p.Find(condition, where...)
	updatedCount := 0

	for _, elem := range elems {
		idField := reflect.ValueOf(elem).Elem().FieldByName("Id")
		if !idField.IsValid() {
			continue
		}

		id := idField.Interface().(int64)
		if p.SetFields(id, fields) {
			updatedCount++
		}
	}

	return updatedCount
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

	switch op {
	case "equal", "eq", "=":
		return reflect.DeepEqual(v1Val.Interface(), v2Val.Interface())
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
	default:
		panic("unknown field condition: " + op)
	}
	return false
}

// findFieldByName finds a struct field by name, case-insensitively
func FindFieldByName(val reflect.Value, name string) reflect.Value {
	valType := val.Type()
	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		if strings.EqualFold(field.Name, name) {
			return val.Field(i)
		}
	}
	return reflect.Value{}
}

type FindCondition struct {
	Field string
	Value interface{}
	Op    string // "equal", "eq", =, "gt", "more", >, "lt", "less", <
}

/*
condition logic can be:
"" - doesnt matter
OR - where1 || where2 || where3 ...
AND - where1 && where2 && where3
*/
func (p *StructMap[V]) Find(condition string, where ...FindCondition) []V {
	var ret []V
	p.cm.Iterate(func(key int64, v V) bool {
		match := false
		val := reflect.Indirect(reflect.ValueOf(v))

		for _, cond := range where {
			//f := val.FieldByName(cond.Field)
			f := FindFieldByName(val, cond.Field)
			if !f.IsValid() {
				match = false
				break
			}

			// Determine if the condition is met
			isMatch := compareValues(f.Interface(), cond.Value, cond.Op)

			if condition == "AND" {
				if !isMatch {
					match = false
					break
				}
				match = true
			} else if condition == "OR" {
				if isMatch {
					match = true
					break
				}
			} else {
				match = isMatch
			}
		}

		if match {
			ret = append(ret, v)
		}
		return true
	})
	return ret
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
