package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CheckinLog records a user's daily check-in. Hard-delete only (no soft-delete mixin).
type CheckinLog struct {
	ent.Schema
}

func (CheckinLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "checkin_logs"},
	}
}

func (CheckinLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id"),
		// reward credited for this check-in
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		// consecutive check-in streak this counts as
		field.Int("consecutive_days").Default(1),
		// check-in date (UTC date), part of the per-user unique constraint
		field.Time("checkin_date").
			SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (CheckinLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("checkin_logs").
			Field("user_id").
			Required().
			Unique(),
	}
}

func (CheckinLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "checkin_date").Unique(),
		index.Fields("user_id"),
		index.Fields("checkin_date"),
	}
}
