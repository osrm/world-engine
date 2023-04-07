package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/argus-labs/cardinal/component"
	"github.com/argus-labs/cardinal/entity"
)

type EntryStorage interface {
	Length() int
	Push(*Entry)
	Set(entity.ID, *Entry)
	Get(entity.ID) *Entry
}

type entryStorageImpl struct {
	entries []*Entry
}

func NewEntryStorage() EntryStorage {
	return &entryStorageImpl{entries: make([]*Entry, 1, 256)}
}

func (e entryStorageImpl) Length() int {
	return len(e.entries)
}

func (e *entryStorageImpl) Push(entry *Entry) {
	e.entries = append(e.entries, entry)
}

func (e *entryStorageImpl) Set(id entity.ID, entry *Entry) {
	e.entries[id] = entry
}

func (e entryStorageImpl) Get(id entity.ID) *Entry {
	return e.entries[id]
}

// Entry is a struct that contains an entity and a location in an archetype.
type Entry struct {
	w WorldAccessor // TODO(technicallyty): fix, replace with iface probably..

	id     entity.ID
	entity Entity
	loc    *Location
}

func NewEntry(id entity.ID, e entity.Entity, loc *Location, w WorldAccessor) *Entry {
	return &Entry{
		w:      w, // TODO(technicallyty): todo, fix
		id:     id,
		entity: e,
		loc:    loc,
	}
}

func (e *Entry) SetEntity(en entity.Entity) {
	e.entity = en
}

func (e *Entry) SetLocation(loc *Location) {
	e.loc = loc
}

// Get returns the component from the entry
func Get[T any](e *Entry, cType component.IComponentType) (*T, error) {
	var comp *T
	compBz := e.Component(cType)
	var buf bytes.Buffer
	buf.Write(compBz)
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(comp)
	return comp, err
}

// Add adds the component to the entry.
func Add[T any](e *Entry, cType component.IComponentType, component *T) error {
	bz, err := MarshalComponent[T](component)
	if err != nil {
		return err
	}
	e.AddComponent(cType, bz)
	return nil
}

// Set sets the component of the entry.
func Set[T any](e *Entry, ctype component.IComponentType, component *T) error {
	bz, err := MarshalComponent[T](component)
	if err != nil {
		return err
	}
	e.SetComponent(ctype, bz)
	return nil
}

// SetValue sets the value of the component.
func SetValue[T any](e *Entry, ctype component.IComponentType, value T) error {
	c, err := Get[T](e, ctype)
	if err != nil {
		return err
	}
	*c = value
	return nil
}

// Remove removes the component from the entry.
func Remove[T any](e *Entry, ctype component.IComponentType) {
	e.RemoveComponent(ctype)
}

// Valid returns true if the entry is valid.
func Valid(e *Entry) bool {
	if e == nil {
		return false
	}
	return e.Valid()
}

// Id returns the entity id.
func (e *Entry) Id() entity.ID {
	return e.id
}

// Entity returns the entity.
func (e *Entry) Entity() Entity {
	return e.entity
}

// Component returns the component.
func (e *Entry) Component(cType component.IComponentType) []byte {
	c := e.loc.CompIndex
	a := e.loc.ArchIndex
	return e.w.Component(cType, a, c)
}

// SetComponent sets the component.
func (e *Entry) SetComponent(cType component.IComponentType, component []byte) {
	c := e.loc.CompIndex
	a := e.loc.ArchIndex
	e.w.SetComponent(cType, component, a, c)
}

// AddComponent adds the component to the entity.
func (e *Entry) AddComponent(cType component.IComponentType, components ...[]byte) {
	if len(components) > 1 {
		panic("AddComponent: component argument must be a single value")
	}
	if !e.HasComponent(cType) {
		c := e.loc.CompIndex
		a := e.loc.ArchIndex

		baseLayout := e.w.GetLayout(a)
		targetArc := e.w.GetArchetypeForComponents(append(baseLayout, cType))
		e.w.TransferArchetype(a, targetArc, c)

		e.loc = e.w.Entry(e.entity).loc
	}
	if len(components) == 1 {
		e.SetComponent(cType, components[0])
	}
}

// RemoveComponent removes the component from the entity.
func (e *Entry) RemoveComponent(cType component.IComponentType) {
	// if the entry doesn't even have this component, we can just return.
	if !e.Archetype().Layout().HasComponent(cType) {
		return
	}

	c := e.loc.CompIndex
	a := e.loc.ArchIndex

	baseLayout := e.w.GetLayout(a)
	targetLayout := make([]component.IComponentType, 0, len(baseLayout)-1)
	for _, c2 := range baseLayout {
		if c2 == cType {
			continue
		}
		targetLayout = append(targetLayout, c2)
	}

	targetArc := e.w.GetArchetypeForComponents(targetLayout)
	e.w.TransferArchetype(e.loc.ArchIndex, targetArc, c)

	e.loc = e.w.Entry(e.entity).loc
}

// Remove removes the entity from the world.
func (e *Entry) Remove() {
	e.w.Remove(e.entity)
}

// Valid returns true if the entry is valid.
func (e *Entry) Valid() bool {
	return e.w.Valid(e.entity)
}

// Archetype returns the archetype.
func (e *Entry) Archetype() ArchetypeStorage {
	a := e.loc.ArchIndex
	return e.w.Archetype(a)
}

// HasComponent returns true if the entity has the given component type.
func (e *Entry) HasComponent(componentType component.IComponentType) bool {
	return e.Archetype().Layout().HasComponent(componentType)
}

func (e *Entry) String() string {
	var out bytes.Buffer
	out.WriteString("Entry: {")
	out.WriteString(e.Entity().String())
	out.WriteString(", ")
	out.WriteString(e.Archetype().Layout().String())
	out.WriteString(", Valid: ")
	out.WriteString(fmt.Sprintf("%v", e.Valid()))
	out.WriteString("}")
	return out.String()
}