package utils_test

import (
	"afho_backend/utils"
	"testing"
)

func TestNewCollection(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	if len(collection.Data) != len(slice) {
		t.Errorf("Expected %d but got %d", len(slice), len(collection.Data))
	}

	for i, item := range collection.Data {
		if item != slice[i] {
			t.Errorf("Expected %d but got %d", slice[i], item)
		}
	}
}

func TestCollectionGet(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	value, err := collection.Get(func(i int) bool {
		return i == 3
	})

	if err != nil {
		t.Errorf("Expected nil but got %s", err.Error())
	}

	if value != 3 {
		t.Errorf("Expected 3 but got %d", value)
	}

	_, err = collection.Get(func(i int) bool {
		return i == 6
	})

	if err == nil {
		t.Errorf("Expected error but got nil")
	}
}

func TestCollectionGetIndex(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	index, err := collection.GetIndex(func(i int) bool {
		return i == 3
	})

	if err != nil {
		t.Errorf("Expected nil but got %s", err.Error())
	}

	if index != 2 {
		t.Errorf("Expected 2 but got %d", index)
	}

	_, err = collection.GetIndex(func(i int) bool {
		return i == 6
	})

	if err == nil {
		t.Errorf("Expected error but got nil")
	}
}

func TestCollectionInsert(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	collection.Insert(6)

	if len(collection.Data) != len(slice)+1 {
		t.Errorf("Expected %d but got %d", len(slice)+1, len(collection.Data))
	}

	if collection.Data[len(collection.Data)-1] != 6 {
		t.Errorf("Expected 6 but got %d", collection.Data[len(collection.Data)-1])
	}
}

func TestCollectionInsertAt(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	collection.InsertAt(2, 6)

	if len(collection.Data) != len(slice)+1 {
		t.Errorf("Expected %d but got %d", len(slice)+1, len(collection.Data))
	}

	if collection.Data[2] != 6 {
		t.Errorf("Expected 6 but got %d", collection.Data[2])
	}
}

func TestCollectionUpdate(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	err := collection.Update(2, 6)

	if err != nil {
		t.Errorf("Expected nil but got %s", err.Error())
	}

	if collection.Data[2] != 6 {
		t.Errorf("Expected 6 but got %d", collection.Data[2])
	}

	err = collection.Update(6, 6)

	if err == nil {
		t.Errorf("Expected error but got nil")
	}
}

func TestCollectionRemoveItemAtIndex(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	err := collection.RemoveItemAtIndex(2)

	if err != nil {
		t.Errorf("Expected nil but got %s", err.Error())
	}

	if len(collection.Data) != len(slice)-1 {
		t.Errorf("Expected %d but got %d", len(slice)-1, len(collection.Data))
	}

	if collection.Data[2] != 4 {
		t.Errorf("Expected 4 but got %d", collection.Data[2])
	}

	err = collection.RemoveItemAtIndex(6)

	if err == nil {
		t.Errorf("Expected error but got nil")
	}
}

func TestCollectionRemoveItem(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	collection.RemoveItem(func(i int) bool {
		return i == 3
	})

	if collection.Data[2] != 4 {
		t.Errorf("Expected 4 at index 2 but got %d", collection.Data[2])
	}

	if len(collection.Data) != len(slice)-1 {
		t.Errorf("Expected %d but got %d", len(slice)-1, len(collection.Data))
	}

	collection.RemoveItem(func(i int) bool {
		return i == 6
	})

	if len(collection.Data) != len(slice)-1 {
		t.Errorf("Expected same length but got different length")
	}
}

func TestCollectionShift(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	collection := utils.NewCollection(slice)

	collection.Shift(2)

	if len(collection.Data) != len(slice)-2 {
		t.Errorf("Expected %d but got %d", len(slice)-2, len(collection.Data))
	}

	if collection.Data[0] != 3 {
		t.Errorf("Expected 3 at index 0 but got %d", collection.Data[0])
	}

	collection.Shift(6)

	if len(collection.Data) != 0 {
		t.Errorf("Expected 0 but got %d", len(collection.Data))
	}
}
