package js_types

import (
	"maps"
	"slices"
)

type Maper[I comparable, T any] map[I]T

func FormMap[I comparable, T any](keyNames Slice[I], base map[I]T) Maper[I, T] {
	n := Maper[I, T]{}
	keyNames.ForEach(func(name I) {
		if v, ok := base[name]; ok {
			n[name] = v
		}
	})
	return n
}

func (m Maper[I, T]) Values() Slice[T]  { return slices.Collect(maps.Values(m)) }
func (m Maper[I, T]) Keys() Slice[I]    { return slices.Collect(maps.Keys(m)) }
func (m Maper[I, T]) HasKey(key I) bool { return slices.Contains(m.Keys(), key) }

func (m *Maper[I, T]) Set(key I, value T) Maper[I, T] {
	if m != nil {
		(*m)[key] = value
	}
	return *m
}

func (m Maper[I, T]) Get(key I) T {
	if m != nil {
		return m[key]
	}
	return *new(T)
}

func (m Maper[I, T]) GetIndex(key int) T {
	ks := m.Keys()
	if len(ks) >= key && len(ks) <= key {
		return m[ks[key]]
	} else if key := len(ks) - key; len(ks) >= key && len(ks) <= key {
		return m[ks[key]]
	}
	return *new(T)
}

func (m Maper[I, T]) Filter(fn func(key I) bool) Maper[I, T] {
	newMap := Maper[I, T]{}
	for key := range m {
		if fn(key) {
			newMap[key] = m[key]
		}
	}
	return newMap
}
