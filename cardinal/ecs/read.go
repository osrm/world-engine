package ecs

import (
	"encoding/json"
	"fmt"
	"github.com/invopop/jsonschema"
)

type IRead interface {
	// Name returns the name of the read.
	Name() string
	// HandleRead is given a reference to the world, json encoded bytes that represent a read request
	// and is expected to return a json encoded response struct.
	HandleRead(*World, []byte) ([]byte, error)
	// Schema returns the json schema of the read request.
	Schema() *jsonschema.Schema
}

type ReadType[Request any, Reply any] struct {
	name    string
	handler func(world *World, req Request) (Reply, error)
}

var _ IRead = NewReadType[struct{}, struct{}]("", nil)

func NewReadType[Request any, Reply any](
	name string,
	handler func(world *World, req Request) (Reply, error),
) *ReadType[Request, Reply] {
	return &ReadType[Request, Reply]{
		name:    name,
		handler: handler,
	}
}

func (r *ReadType[req, rep]) Name() string {
	return r.name
}

func (r *ReadType[req, rep]) Schema() *jsonschema.Schema {
	return jsonschema.Reflect(new(req))
}

func (r *ReadType[req, rep]) HandleRead(w *World, bz []byte) ([]byte, error) {
	t := new(req)
	err := json.Unmarshal(bz, t)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal read request into type %T: %w", *t, err)
	}
	res, err := r.handler(w, *t)
	if err != nil {
		return nil, err
	}
	bz, err = json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal response %T: %w", res, err)
	}
	return bz, nil
}