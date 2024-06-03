package structmap

import (
	"reflect"
	"strings"
	"sync/atomic"

	"github.com/goupdate/compactmap"
)

type StructMap[V any] struct {
	cm *compactmap.CompactMap[int64, V] //im-memory database

	info        *compactmap.CompactMap[int64, int64] //store maxId
	storageFile string

	maxId int64 //max stored id, incremented after Add
}

func New[V any](storageFile string, failIfNotLoaded bool) (*StructMap[V], error) {
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
	store, ex := p.cm.Get(id)
	if !ex {
		return false
	}
	val := reflect.ValueOf(store).Elem()
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

type FindCondition struct {
	Field string
	Value interface{}
	Op    string
}

func compareValues(v1, v2 interface{}, op string) bool {
	switch op {
	case "equal", "eq", "=":
		return v1 == v2
	case "gt", "more", ">":
		return reflect.ValueOf(v1).Float() > reflect.ValueOf(v2).Float()
	case "lt", "less", "<":
		return reflect.ValueOf(v1).Float() < reflect.ValueOf(v2).Float()
	case "like", "contains":
		str1, ok1 := v1.(string)
		str2, ok2 := v2.(string)
		return ok1 && ok2 && strings.Contains(str1, str2)
	default:
		return false
	}
}

func (p *StructMap[V]) Find(where ...FindCondition) []V {
	var ret []V
	p.cm.Iterate(func(key int64, v V) bool {
		match := true
		val := reflect.ValueOf(v).Elem()
		for _, cond := range where {
			f := val.FieldByName(cond.Field)
			if !f.IsValid() || !compareValues(f.Interface(), cond.Value, cond.Op) {
				match = false
				break
			}
		}
		if match {
			ret = append(ret, v)
		}
		return true
	})
	return ret
}

func (p *StructMap[V]) Get(id int64) (V, bool) {
	return p.cm.Get(id)
}

func (p *StructMap[V]) Add(v V) int64 {
	id := atomic.AddInt64(&p.maxId, 1)
	p.cm.Add(id, v)
	return id
}
