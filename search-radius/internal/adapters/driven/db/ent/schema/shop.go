package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	e "search-radius/pkg/database/ent"
)

// Shop holds the schema definition for the Shop entity.
type Shop struct {
	ent.Schema
}

// Mixin of the Shop.
func (Shop) Mixin() []ent.Mixin {
	return []ent.Mixin{
		e.BaseMixin{},
	}
}

// Fields of the Shop.
func (Shop) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.Float("lat"),
		field.Float("lng"),
	}
}

// Edges of the Shop.
func (Shop) Edges() []ent.Edge {
	return nil
}
