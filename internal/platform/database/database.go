package database

import (
	"context"
	"net/url"

	// "github.com/jmoiron/sqlx"
	"github.com/go-pg/pg/v9"
	// orm "github.com/go-pg/pg/v9/orm"
	// _ "github.com/lib/pq" // The database driver in use.
	"go.opencensus.io/trace"
)

// Config is the required properties to use the database.
type Config struct {
	User       string
	Password   string
	Host       string
	Name       string
	DisableTLS bool
	QueryHook  bool
}

type dbLogger struct{}

// BeforeQuery is a hook for go-pg
func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	// logger := zap.S().With("package", "storage.postgres")
	trace.StartSpan(c, "platform.db.Query")
	_, err := q.FormattedQuery()
	if err != nil {
		return c, err
	}
	// if err != nil {
	// 	logger.Errorw("Error querying: ", err)
	// }
	// logger.Infow("querying ", "stmt: ", qwer)
	return c, nil
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*pg.DB, error) {
	dbURL := BuildDbURL(cfg)

	// return sqlx.Open("postgres", u.String())
	connOpt, err := pg.ParseURL(dbURL)
	if err != nil {
		return nil, err
	}

	db := pg.Connect(connOpt)
	// Define query hook
	if cfg.QueryHook {
		db.AddQueryHook(dbLogger{})
	}
	return db, nil
}

// BuildDbURL returns database url to be parsed by database adapter
func BuildDbURL(cfg Config) string {
	// Define SSL mode.
	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// Query parameters.
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")

	// Construct url.
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	return u.String()
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *pg.DB) error {
	ctx, span := trace.StartSpan(ctx, "platform.DB.StatusCheck")
	defer span.End()

	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	const q = `SELECT true`
	var tmp bool
	// return db.QueryRowContext(ctx, q).Scan(&tmp)
	_, err := db.QueryOneContext(ctx, &tmp, q)
	return err
}
