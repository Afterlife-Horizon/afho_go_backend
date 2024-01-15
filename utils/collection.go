package utils

import (
	"errors"
)

type Collection[T any] struct {
	Data []T
}

func NewCollection[T any](slice []T) Collection[T] {
	return Collection[T]{
		Data: slice,
	}
}

func (collection *Collection[T]) Get(fn func(T) bool) (T, error) {
	var err error
	for _, item := range collection.Data {
		if fn(item) {
			return item, err
		}
	}
	var zeroValue T
	err = errors.New("no value found")
	return zeroValue, err
}

func (collection *Collection[T]) GetIndex(fn func(T) bool) (int, error) {
	var err error
	for index, item := range collection.Data {
		if fn(item) {
			return index, err
		}
	}
	err = errors.New("no value found")
	return -1, err
}

func (collection *Collection[T]) Insert(value T) {
	collection.Data = append(collection.Data, value)
}

func (collection *Collection[T]) Update(index int, value T) error {
	var err error
	if index < 0 || index > len(collection.Data) {
		err = errors.New("index  out of range")
		return err
	}
	collection.Data[index] = value
	return nil
}

func (collection *Collection[T]) RemoveItemAtIndex(index int) error {
	var err error
	if index < 0 || index > len(collection.Data) {
		err = errors.New("index  out of range")
		return err
	}
	collection.Data = append(collection.Data[:index], collection.Data[index+1:]...)
	return nil
}

func (collection *Collection[T]) RemoveItem(fn func(T) bool) {
	for index, item := range collection.Data {
		if fn(item) {
			collection.Data = append(collection.Data[:index], collection.Data[index+1:]...)
		}
	}
}
