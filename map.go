package compactmap

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"sync"

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

func (m *CompactMap[K, V]) Clear() {
	m.Lock()
	defer m.Unlock()

	m.buffers = m.buffers[0:0]
	m.changed = true
}

// sync.Map analog
func (m *CompactMap[K, V]) LoadOrStore(key K, value V) (old V, loaded bool) {
	m.RLock()
	defer m.RUnlock()

	old, loaded = m.get(key)
	if loaded {
		return old, true
	}

	m.addOrSet(key, value)
	return value, false
}

func (m *CompactMap[K, V]) LoadAndDelete(key K) (old V, loaded bool) {
	m.RLock()
	defer m.RUnlock()

	old, loaded = m.get(key)
	if loaded {
		m.delete(key)
		return old, true
	}

	var zero V
	return zero, false
}

// sync.map - compatible
func (m *CompactMap[K, V]) Store(key K, value V) {
	m.AddOrSet(key, value)
}

// Add or Set
func (m *CompactMap[K, V]) AddOrSet(key K, value V) (overwrited bool) {
	m.Lock()
	defer m.Unlock()

	return m.addOrSet(key, value)
}

func (m *CompactMap[K, V]) addOrSet(key K, value V) (overwrited bool) {
	if len(m.buffers) == 0 {
		newBuffer := &[]Entry[K, V]{Entry[K, V]{Key: key, Value: value}}
		m.buffers = append(m.buffers, newBuffer)
		m.changed = true
		overwrited = false
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
			overwrited = true
		} else {
			*buffer = append(*buffer, Entry[K, V]{})
			copy((*buffer)[index+1:], (*buffer)[index:])
			(*buffer)[index] = Entry[K, V]{Key: key, Value: value}
			overwrited = false
		}
		m.changed = true
		return
	}

	// If no appropriate buffer found, create a new one
	newBuffer := &[]Entry[K, V]{Entry[K, V]{Key: key, Value: value}}
	m.buffers = append(m.buffers, newBuffer)
	overwrited = false

	m.changed = true
	return
}

// alias map-compatible
func (m *CompactMap[K, V]) Load(key K) (V, bool) {
	return m.Get(key)
}

func (m *CompactMap[K, V]) Get(key K) (V, bool) {
	m.RLock()
	defer m.RUnlock()

	return m.get(key)
}

func (m *CompactMap[K, V]) get(key K) (V, bool) {
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

	m.delete(key)
}

func (m *CompactMap[K, V]) delete(key K) {
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

// sync.Map alias
func (m *CompactMap[K, V]) Range(fn func(key K, val V) bool) {
	m.Iterate(fn)
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

func (m *CompactMap[K, V]) Init(filename string) error {
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

		m.AddOrSet(key, value)
	}

	m.changed = false
	m.loadedFile = filename
	return nil
}

func serialize[T any](data T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserialize[T any](data []byte) (T, error) {
	var result T
	dec := gob.NewDecoder(bytes.NewBuffer(data))
	err := dec.Decode(&result)
	if err != nil {
		return result, err
	}
	return result, nil
}
