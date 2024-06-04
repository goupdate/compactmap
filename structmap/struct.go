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

	cm   *compactmap.CompactMap[int64, V]     //im-memory database
	info *compactmap.CompactMap[int64, int64] //store maxId

	storageFile string

	maxId int64 //max stored id, incremented after Add
}

// V - should be pointer to struct
func New[V any](storageFile string, failIfNotLoaded bool) (*StructMap[V], error) {
	var zero V
	if reflect.ValueOf(&zero).Elem().Kind() != reflect.Pointer {
		panic(fmt.Sprintf("cant add %v need pointer", reflect.TypeOf(zero).Name()))
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
		info.Add(1, 1)
		maxId = 1
	}

	return &StructMap[V]{cm: cm, maxId: maxId, info: info, storageFile: storageFile}, nil
}

func (p *StructMap[V]) GetMaxId() int64 {
	return atomic.LoadInt64(&p.maxId)
}

func (p *StructMap[V]) Save() error {
	p.info.Add(1, p.maxId)
	err := p.cm.Save(p.storageFile)
	if err != nil {
		return err
	}
	err2 := p.info.Save(p.storageFile + "i")
	if err2 != nil {
		return err2
	}
	return nil
}

func (p *StructMap[V]) SetField(id int64, field string, value interface{}) bool {
	p.Lock()
	defer p.Unlock()

	store, ex := p.cm.Get(id)
	if !ex {
		return false
	}
	val := reflect.Indirect(reflect.ValueOf(store))
	f := val.FieldByName(field)
	if !f.IsValid() || !f.CanSet() {
		return false
	}
	f.Set(reflect.ValueOf(value))
	return true
}

func (p *StructMap[V]) GetAll() []V {
	var ret []V
	p.cm.Iterate(func(k int64, v V) bool {
		ret = append(ret, v)
		return true
	})
	return ret
}

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
	}
	return false
}

type FindCondition struct {
	Field string
	Value interface{}
	Op    string // "equal", "eq", =
	// "gt", "more", >
	// "lt", "less", <
}

/*
condition = "" - no, for single where  / AND / OR
*/
func (p *StructMap[V]) Find(condition string, where ...FindCondition) []V {
	var ret []V
	p.cm.Iterate(func(key int64, v V) bool {
		match := false
		val := reflect.Indirect(reflect.ValueOf(v))

		for _, cond := range where {
			f := val.FieldByName(cond.Field)
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

func (p *StructMap[V]) Iterate(fn func(v V) bool) {
	p.cm.Iterate(func(_ int64, v V) bool {
		return fn(v)
	})
}

func (p *StructMap[V]) Get(id int64) (V, bool) {
	return p.cm.Get(id)
}

func (p *StructMap[V]) Delete(id int64) {
	p.cm.Delete(id)
}

func (p *StructMap[V]) Clear() {
	p.maxId = 1
	p.cm.Clear()
}

func (p *StructMap[V]) Add(v V) int64 {
	if reflect.ValueOf(v).Kind() != reflect.Pointer {
		panic(fmt.Sprintf("cant add %v need pointer", reflect.ValueOf(v).Type().Name()))
	}
	id := atomic.AddInt64(&p.maxId, 1)
	p.cm.Add(id, v)
	return id
}
