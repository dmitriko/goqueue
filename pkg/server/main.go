package server

import (
	"container/list"
	"errors"
)

type Item struct {
	Key   string
	Value string
	el    *list.Element
}

func NewItem(key, value string) *Item {
	return &Item{Key: key, Value: value}
}

type Storage struct {
	items map[string]*Item
	list  *list.List
}

func NewStorage() *Storage {
	return &Storage{
		items: make(map[string]*Item),
		list:  list.New(),
	}
}

func (s *Storage) AddItem(item *Item) error {
	if item.Key == "" {
		return errors.New("Key could not be empty string.")
	}
	item.el = s.list.PushBack(item)
	s.items[item.Key] = item
	return nil
}

func (s *Storage) GetItem(key string) (*Item, bool) {
	i, e := s.items[key]
	return i, e
}

//Returns true if item was deleted, false otherwise
func (s *Storage) RemoveItem(key string) bool {
	item, exists := s.items[key]
	if !exists {
		return false
	}
	s.list.Remove(item.el)
	delete(s.items, key)
	return true
}

func el2item(el *list.Element) *Item {
	if el == nil {
		return nil
	}
	return el.Value.(*Item)
}

func (s *Storage) Oldest() *Item {
	return el2item(s.list.Front())
}

func (item *Item) Next() *Item {
	return el2item(item.el.Next())
}

func (s *Storage) GetAllItems() []*Item {
	r := []*Item{}
	for item := s.Oldest(); item != nil; item = item.Next() {
		r = append(r, item)
	}
	return r
}
