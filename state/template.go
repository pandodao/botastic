package state

import (
	"sync"
	"text/template"
)

type templateCache struct {
	sync.Mutex
	km map[string]string
	m  map[string]*template.Template
}

func newTemplateCache() *templateCache {
	return &templateCache{
		km: map[string]string{},
		m:  map[string]*template.Template{},
	}
}

func (c *templateCache) getTemplate(key string, v string) (*template.Template, error) {
	c.Lock()
	defer c.Unlock()

	if v, ok := c.m[v]; ok {
		return v, nil
	}

	delete(c.m, c.km[key])
	delete(c.km, key)

	r, err := template.New(key).Parse(v)
	if err != nil {
		return nil, err
	}

	c.km[key] = v
	c.m[v] = r
	return r, nil
}
