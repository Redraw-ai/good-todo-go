package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Todo holds the schema definition for the Todo entity.
type Todo struct {
	ent.Schema
}

// Fields of the Todo.
func (Todo) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),
		field.String("tenant_id").
			NotEmpty().
			Immutable(),
		field.String("user_id").
			NotEmpty().
			Immutable(),
		field.String("title").
			NotEmpty(),
		field.Text("description").
			Optional().
			Default(""),
		field.Bool("completed").
			Default(false),
		field.Time("due_date").
			Optional().
			Nillable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(func() time.Time {
				return time.Now().UTC()
			}).
			Immutable(),
		field.Time("updated_at").
			Default(func() time.Time {
				return time.Now().UTC()
			}).
			UpdateDefault(func() time.Time {
				return time.Now().UTC()
			}),
	}
}

// Edges of the Todo.
func (Todo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("todos").
			Field("user_id").
			Required().
			Unique().
			Immutable(),
	}
}

// Indexes of the Todo.
func (Todo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("user_id"),
		index.Fields("tenant_id", "user_id"),
	}
}
