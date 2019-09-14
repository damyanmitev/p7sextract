package main

import "github.com/lxn/walk"

type NamedIconCache struct {
	walk.Disposables
	namedIcons map[string]*walk.Icon
}

func NewNamedIconCache() *NamedIconCache {
	return &NamedIconCache{
		namedIcons: make(map[string]*walk.Icon),
	}
}

func (d *NamedIconCache) AddNamed(name string, icon *walk.Icon) {
	d.Add(icon)
	d.namedIcons[name] = icon
}

func (d *NamedIconCache) Get(name string) *walk.Icon {
	return d.namedIcons[name]
}
