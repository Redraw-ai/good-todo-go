package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			NotEmpty().
			Immutable(),
		field.String("tenant_id").
			NotEmpty().
			Immutable(),
		field.String("email").
			NotEmpty(),
		field.String("password_hash").
			NotEmpty().
			Sensitive(),
		field.String("name").
			Default(""),
		field.Enum("role").
			Values("admin", "member").
			Default("member"),
		field.Bool("email_verified").
			Default(false),
		field.String("verification_token").
			Optional().
			Nillable(),
		field.Time("verification_token_expires_at").
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

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tenant", Tenant.Type).
			Ref("users").
			Field("tenant_id").
			Required().
			Unique().
			Immutable(),
		edge.To("todos", Todo.Type),
	}
}

// Indexes of the User.
func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "email").Unique(),
		index.Fields("tenant_id"),
	}
}
