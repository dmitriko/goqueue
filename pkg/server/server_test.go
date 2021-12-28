package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStorage(t *testing.T) {
	s := NewStorage()
	assert.NotNil(t, s)
}

func TestStorageSet(t *testing.T) {
	s := NewStorage()
	err := s.AddItem(&Item{Key: "foo", Value: "bar"})
	assert.Nil(t, err)
	err = s.AddItem(&Item{Key: "", Value: "spam"})
	assert.NotNil(t, err)
}

func TestStorageGet(t *testing.T) {
	it1 := NewItem("foo", "bar")
	//	it2 := NewItem("spam", "egg")
	s := NewStorage()
	assert.Nil(t, s.AddItem(it1))
	it2, exists := s.GetItem("foo")
	assert.True(t, exists)
	assert.Equal(t, it1, it2)
	_, exists = s.GetItem("spam")
	assert.False(t, exists)
}

func TestRemoveItem(t *testing.T) {
	it1 := NewItem("foo", "bar")
	s := NewStorage()
	assert.Nil(t, s.AddItem(it1))
	assert.True(t, s.RemoveItem("foo"))
	assert.False(t, s.RemoveItem("spam"))
}

func TestGetAllItems(t *testing.T) {
	it1 := NewItem("foo", "bar")
	it2 := NewItem("spam", "egg")
	s := NewStorage()
	assert.Nil(t, s.AddItem(it1))
	assert.Nil(t, s.AddItem(it2))
	items := s.GetAllItems()
	if assert.Equal(t, 2, len(items)) {
		assert.Equal(t, items[0], it1)
		assert.Equal(t, items[1], it2)
	}
}
