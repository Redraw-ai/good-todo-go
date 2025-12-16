package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// TenantUserView holds the schema definition for the TenantUserView entity.
// This is a read-only view for tenant-scoped user access.
type TenantUserView struct {
	ent.Schema
}

// Annotations of the TenantUserView.
func (TenantUserView) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Define the view using SQL builder
		entsql.ViewFor("postgres", func(s *sql.Selector) {
			u := sql.Table("users").As("u")
			t := sql.Table("tenants").As("t")
			s.From(u).
				Join(t).On(u.C("tenant_id"), t.C("id")).
				Select(
					u.C("id"),
					u.C("tenant_id"),
					sql.As(t.C("name"), "tenant_name"),
					sql.As(t.C("slug"), "tenant_slug"),
					u.C("email"),
					u.C("name"),
					u.C("role"),
					u.C("email_verified"),
					u.C("created_at"),
					u.C("updated_at"),
				)
		}),
		// Skip migration for view (view is created manually or via ViewFor)
		entsql.Skip(),
	}
}

// Fields of the TenantUserView.
func (TenantUserView) Fields() []ent.Field {
	return []ent.Field{
		field.String("id"),
		field.String("tenant_id"),
		field.String("tenant_name"),
		field.String("tenant_slug"),
		field.String("email"),
		field.String("name"),
		field.String("role"),
		field.Bool("email_verified"),
		field.Time("created_at").
			Default(func() time.Time {
				return time.Now().UTC()
			}),
		field.Time("updated_at").
			Default(func() time.Time {
				return time.Now().UTC()
			}),
	}
}

// Edges of the TenantUserView.
func (TenantUserView) Edges() []ent.Edge {
	return nil
}
