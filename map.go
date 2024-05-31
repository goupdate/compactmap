package compactmap

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"unsafe"

	"golang.org/x/exp/constraints"
)

const maxSliceSize = 1000

type Entry[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

type CompactMap[K constraints.Ordered, V any] struct {
	sync.RWMutex

	buffers    []*[]Entry[K, V]
	changed    bool
	loadedFile string
}

func NewCompactMap[K constraints.Ordered, V any]() *CompactMap[K, V] {
	return &CompactMap[K, V]{
		buffers:    make([]*[]Entry[K, V], 0, 100),
		changed:    false,
		loadedFile: "",
	}
}

func (m *CompactMap[K, V]) Add(key K, value V) {
	m.Lock()
	defer m.Unlock()

	if len(m.buffers) == 0 {
		newBuffer := &[]Entry[K, V]{Entry[K, V]{Key: key, Value: value}}
		m.buffers = append(m.buffers, newBuffer)
		m.changed = true
		return
	}

	bufferIndex := len(m.buffers) - 1

	buffer := m.buffers[bufferIndex]
	if len(*buffer) < maxSliceSize {
		index := sort.Search(len(*buffer), func(i int) bool {
			return (*buffer)[i].Key >= key
		})

		if index < len(*buffer) && (*buffer)[index].Key == key {
			(*buffer)[index].Value = value
		} else {
			*buffer = append(*buffer, Entry[K, V]{})
			copy((*buffer)[index+1:], (*buffer)[index:])
			(*buffer)[index] = Entry[K, V]{Key: key, Value: value}
		}
		m.changed = true
		return
	}

	// If no appropriate buffer found, create a new one
	newBuffer := &[]Entry[K, V]{Entry[K, V]{Key: key, Value: value}}
	m.buffers = append(m.buffers, newBuffer)

	m.changed = true
}

func (m *CompactMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()

	for bufferIndex := range len(m.buffers) {
		buffer := m.buffers[bufferIndex]
		index := sort.Search(len(*buffer), func(i int) bool {
			return (*buffer)[i].Key >= key
		})

		if index < len(*buffer) && (*buffer)[index].Key == key {
			return (*buffer)[index].Value, true
		}
	}

	var zero V
	return zero, false
}

func (m *CompactMap[K, V]) Delete(key K) {
	m.Lock()
	defer m.Unlock()

	for bufferIndex := range len(m.buffers) {
		buffer := m.buffers[bufferIndex]
		index := sort.Search(len(*buffer), func(i int) bool {
			return (*buffer)[i].Key >= key
		})

		if index < len(*buffer) && (*buffer)[index].Key == key {
			if len(*buffer) > 1 {
				//remove element in inner buffer
				*buffer = append((*buffer)[:index], (*buffer)[index+1:]...)
			} else {
				//remove whole slice
				m.buffers = append((m.buffers)[:bufferIndex], (m.buffers)[bufferIndex+1:]...)
			}

			m.changed = true
			return
		}
	}
}

// dont modify database in iterate!
func (m *CompactMap[K, V]) Iterate(fn func(key K, val V) bool) {
	m.RLock()
	defer m.RUnlock()

	for _, buffer := range m.buffers {
		buffer_ := *buffer
		for _, k := range buffer_ {
			if !fn(k.Key, k.Value) {
				return
			}
		}
	}
}

func (m *CompactMap[K, V]) Exist(key K) bool {
	m.RLock()
	defer m.RUnlock()

	for bufferIndex := range len(m.buffers) {
		buffer := m.buffers[bufferIndex]
		index := sort.Search(len(*buffer), func(i int) bool {
			return (*buffer)[i].Key >= key
		})

		if index < len(*buffer) && (*buffer)[index].Key == key {
			return true
		}
	}
	return false
}

func (m *CompactMap[K, V]) Count() int {
	m.RLock()
	defer m.RUnlock()

	count := 0
	for _, buffer := range m.buffers {
		count += len(*buffer)
	}
	return count
}

func (m *CompactMap[K, V]) Stats() string {
	m.RLock()
	defer m.RUnlock()

	count := 0
	for _, buffer := range m.buffers {
		count += len(*buffer)
	}

	str := fmt.Sprintf("%d buffers, total len: %d", len(m.buffers), count)

	return str
}

func (m *CompactMap[K, V]) Save(filename string) error {
	m.RLock()
	defer m.RUnlock()

	if m.loadedFile == filename && !m.changed {
		return nil
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	const bufferSize = 50 * 1024 * 1024 // 50MB
	writer := bufio.NewWriterSize(file, bufferSize)

	writeToFile := func(data []byte) error {
		if _, err := writer.Write(data); err != nil {
			return err
		}
		return nil
	}

	// Write number of entries
	totalEntries := 0 //Count()
	for _, buffer := range m.buffers {
		totalEntries += len(*buffer)
	}

	totalEntriesBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalEntriesBuf, uint64(totalEntries))
	if err := writeToFile(totalEntriesBuf); err != nil {
		return err
	}

	var buf4 [4]byte

	for _, buffer_ := range m.buffers {
		buffer := *buffer_
		// Write keys and values
		for _, entry := range buffer {
			keyData, err := serialize(entry.Key)
			if err != nil {
				return err
			}
			valueData, err := serialize(entry.Value)
			if err != nil {
				return err
			}

			// Write key size and key
			binary.LittleEndian.PutUint32(buf4[:], uint32(len(keyData)))
			if err := writeToFile(buf4[:]); err != nil {
				return err
			}
			if err := writeToFile(keyData); err != nil {
				return err
			}

			// Write value size and value
			binary.LittleEndian.PutUint32(buf4[:], uint32(len(valueData)))
			if err := writeToFile(buf4[:]); err != nil {
				return err
			}
			if err := writeToFile(valueData); err != nil {
				return err
			}
		}
	}

	err = writer.Flush()

	m.changed = false
	m.loadedFile = filename
	return err
}

func (m *CompactMap[K, V]) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 50*1024*1024) // 50MB buffer

	var numEntries int64
	if err := binary.Read(reader, binary.LittleEndian, &numEntries); err != nil {
		return err
	}

	// Read keys and values
	for i := int64(0); i < numEntries; i++ {
		var keySize int32
		if err := binary.Read(reader, binary.LittleEndian, &keySize); err != nil {
			return err
		}
		keyData := make([]byte, keySize)
		if _, err := reader.Read(keyData); err != nil {
			return err
		}
		key, err := deserialize[K](keyData)
		if err != nil {
			return err
		}

		var valueSize int32
		if err := binary.Read(reader, binary.LittleEndian, &valueSize); err != nil {
			return err
		}
		valueData := make([]byte, valueSize)
		if _, err := reader.Read(valueData); err != nil {
			return err
		}
		value, err := deserialize[V](valueData)
		if err != nil {
			return err
		}

		m.Add(key, value)
	}

	m.changed = false
	m.loadedFile = filename
	return nil
}

func serialize[T any](data T) ([]byte, error) {
	var buf []byte
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		size := v.Type().Size()
		buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(v.Int()))
		return buf[:size], nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		size := v.Type().Size()
		buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, v.Uint())
		return buf[:size], nil
	case reflect.Float32, reflect.Float64:
		size := v.Type().Size()
		buf = make([]byte, size)
		binary.LittleEndian.PutUint64(buf, *(*uint64)(unsafe.Pointer(&data)))
		return buf[:size], nil
	case reflect.String:
		str := v.String()
		strLen := uint32(len(str))
		buf = make([]byte, 4+strLen)
		binary.LittleEndian.PutUint32(buf, strLen)
		copy(buf[4:], str)
		return buf, nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			slice := v.Bytes()
			sliceLen := uint32(len(slice))
			buf = make([]byte, 4+sliceLen)
			binary.LittleEndian.PutUint32(buf, sliceLen)
			copy(buf[4:], slice)
			return buf, nil
		}
	}
	return nil, errors.New("unsupported type " + reflect.TypeOf(v).String())
}

func deserialize[T any](data []byte) (T, error) {
	var result T
	v := reflect.ValueOf(&result).Elem()
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		size := v.Type().Size()
		if len(data) < int(size) {
			return result, errors.New("data is too short")
		}
		switch size {
		case 1:
			v.SetInt(int64(data[0]))
		case 2:
			v.SetInt(int64(binary.LittleEndian.Uint16(data[:2])))
		case 4:
			v.SetInt(int64(binary.LittleEndian.Uint32(data[:4])))
		case 8:
			v.SetInt(int64(binary.LittleEndian.Uint64(data[:8])))
		}
		return result, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		size := v.Type().Size()
		if len(data) < int(size) {
			return result, errors.New("data is too short")
		}
		switch size {
		case 1:
			v.SetUint(uint64(data[0]))
		case 2:
			v.SetUint(uint64(binary.LittleEndian.Uint16(data[:2])))
		case 4:
			v.SetUint(uint64(binary.LittleEndian.Uint32(data[:4])))
		case 8:
			v.SetUint(binary.LittleEndian.Uint64(data[:8]))
		}
		return result, nil
	case reflect.Float32, reflect.Float64:
		size := v.Type().Size()
		if len(data) < int(size) {
			return result, errors.New("data is too short")
		}
		switch size {
		case 4:
			v.SetFloat(float64(*(*float32)(unsafe.Pointer(&data[0]))))
		case 8:
			v.SetFloat(*(*float64)(unsafe.Pointer(&data[0])))
		}
		return result, nil
	case reflect.String:
		if len(data) < 4 {
			return result, errors.New("data is too short")
		}
		strLen := binary.LittleEndian.Uint32(data[:4])
		if len(data) < int(4+strLen) {
			return result, errors.New("data is too short")
		}
		v.SetString(string(data[4 : 4+strLen]))
		return result, nil
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			if len(data) < 4 {
				return result, errors.New("data is too short")
			}
			sliceLen := binary.LittleEndian.Uint32(data[:4])
			if len(data) < int(4+sliceLen) {
				return result, errors.New("data is too short")
			}
			v.SetBytes(data[4 : 4+sliceLen])
			return result, nil
		}
	}
	return result, errors.New("unsupported type")
}
