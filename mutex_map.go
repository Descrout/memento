package main

import "sync"

type MutexMap[K comparable, V any] struct {
	sync.Mutex
	data map[K]V
}

func NewMutexMap[K comparable, V any]() *MutexMap[K, V] {
	return &MutexMap[K, V]{
		data: make(map[K]V),
	}
}

func (m *MutexMap[K, V]) Len() int {
	return len(m.data)
}

func (m *MutexMap[K, V]) Set(key K, value V) {
	m.Lock()
	m.data[key] = value
	m.Unlock()
}

func (m *MutexMap[K, V]) SetIfNotExists(key K, value V) V {
	m.Lock()
	defer m.Unlock()
	if !m.Has(key) {
		m.data[key] = value
	}

	return m.data[key]
}

func (m *MutexMap[K, V]) Get(key K) (value V) {
	var ok bool
	if value, ok = m.data[key]; ok {
		return value
	}

	return
}

func (m *MutexMap[K, V]) Update(key K, mutate func(V) V) {
	m.Lock()

	if value, ok := m.data[key]; ok {
		m.data[key] = mutate(value)
	}

	m.Unlock()
}

func (m *MutexMap[K, V]) SetOrUpdate(key K, mutate func(V) V) {
	m.Lock()
	m.data[key] = mutate(m.Get(key))
	m.Unlock()
}

func (m *MutexMap[K, V]) Has(key K) bool {
	_, ok := m.data[key]
	return ok
}

func (m *MutexMap[K, V]) Delete(key K) {
	m.Lock()
	delete(m.data, key)
	m.Unlock()
}

func (m *MutexMap[K, V]) Filter(keep func(K, V) bool) []K {
	m.Lock()

	results := []K{}
	for key, value := range m.data {
		if !keep(key, value) {
			delete(m.data, key)
			results = append(results, key)
		}
	}

	m.Unlock()
	return results
}
