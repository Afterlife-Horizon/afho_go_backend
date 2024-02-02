package utils

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
)

type Collection[T any] struct {
	Data []T
	sync.RWMutex
}

func NewCollection[T any](slice []T) Collection[T] {
	return Collection[T]{
		Data: slice,
	}
}

func (collection *Collection[T]) Get(fn func(T) bool) (T, error) {
	var err error
	collection.RLock()
	for _, item := range collection.Data {
		if fn(item) {
			return item, err
		}
	}
	collection.RUnlock()
	var zeroValue T
	err = errors.New("no value found")
	return zeroValue, err
}

func (collection *Collection[T]) GetIndex(fn func(T) bool) (int, error) {
	var err error
	collection.RLock()
	for index, item := range collection.Data {
		if fn(item) {
			return index, err
		}
	}
	collection.RUnlock()
	err = errors.New("no value found")
	return -1, err
}

func (collection *Collection[T]) Insert(value ...T) {
	collection.Lock()
	collection.Data = append(collection.Data, value...)
	collection.Unlock()
}

func (collection *Collection[T]) InsertAt(position int, value ...T) {
	collection.Lock()
	collection.Data = append(collection.Data[:position], append(value, collection.Data[position:]...)...)
	collection.Unlock()
}

func (collection *Collection[T]) Update(index int, value T) error {
	var err error
	if index < 0 || index > len(collection.Data) {
		err = errors.New("index  out of range")
		return err
	}
	collection.Lock()
	collection.Data[index] = value
	collection.Unlock()
	return nil
}

func (collection *Collection[T]) RemoveItemAtIndex(index int) error {
	var err error
	if index < 0 || index > len(collection.Data) {
		err = errors.New("index  out of range")
		return err
	}
	collection.Lock()
	collection.Data = append(collection.Data[:index], collection.Data[index+1:]...)
	collection.Unlock()
	return nil
}

func (collection *Collection[T]) RemoveItem(fn func(T) bool) {
	for index, item := range collection.Data {
		if fn(item) {
			collection.Lock()
			collection.Data = append(collection.Data[:index], collection.Data[index+1:]...)
			collection.Unlock()
		}
	}
}

func (collection *Collection[T]) Shift(skipCount int) {
	if skipCount > len(collection.Data) {
		skipCount = len(collection.Data)
	}
	collection.Lock()
	collection.Data = collection.Data[skipCount:]
	collection.Unlock()
}

func (collection *Collection[T]) Shuffle(start int, end int, shuffleCount int) {
	if start < 0 || end > len(collection.Data) {
		return
	}

	for i := 0; i < shuffleCount; i++ {
		for j := start; j < end; j++ {
			if j == end {
				break
			}
			var randomIndex = rand.Intn(end-start) + start
			collection.Lock()
			collection.Data[randomIndex], collection.Data[j] = collection.Data[j], collection.Data[randomIndex]
			collection.Unlock()
		}
	}
}

func (collection *Collection[T]) ToString() string {
	var result string = "\n--------------------------\n"
	collection.RLock()
	for index, item := range collection.Data {
		result += fmt.Sprintf("index: %d, value: %v\n", index, item)
	}
	collection.RUnlock()
	result += "--------------------------\n"
	return result
}

func Map[T, U any](collection *Collection[T], fn func(T) U) *Collection[U] {
	var result = NewCollection[U](make([]U, len(collection.Data)))

	collection.RLock()
	for index, item := range collection.Data {
		result.Data[index] = fn(item)
	}
	collection.RUnlock()
	return &result
}

func (collection *Collection[T]) Filter(fn func(T) bool) *Collection[T] {
	var result = NewCollection[T](make([]T, 0))

	collection.RLock()
	for _, item := range collection.Data {
		if fn(item) {
			result.Data = append(result.Data, item)
		}
	}
	collection.RUnlock()
	return &result
}
