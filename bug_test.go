package bug

import (
	"context"
	"testing"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
	"go.opentelemetry.io/otel/trace"

	"entgo.io/bug/ent"
	"entgo.io/bug/ent/enttest"
	"entgo.io/bug/ent/intercept"
	_ "entgo.io/bug/ent/runtime"
)

func TestBugSQLite(t *testing.T) {
	tracer := trace.NewNoopTracerProvider().Tracer("")

	client := enttest.Open(t, dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() {
		_ = client.Close()
	})

	// Instrument reads.
	interceptor := ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			// Get a generic query from a typed-query.
			ctx, span := tracer.Start(ctx, "Query",
				trace.WithSpanKind(trace.SpanKindClient),
			)
			defer span.End()
			q, err := intercept.NewQuery(query)
			if err != nil {
				return nil, err
			}
			return next.Query(ctx, q)
		})
	})
	client.Intercept(interceptor)

	test(t, client)
}

func test(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	client.User.Delete().ExecX(ctx)
	client.User.Create().SetName("Ariel").SetAge(30).ExecX(ctx)
	if n := client.User.Query().CountX(ctx); n != 1 {
		t.Errorf("unexpected number of users: %d", n)
	}
}
