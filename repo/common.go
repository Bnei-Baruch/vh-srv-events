package repo

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"gitlab.bbdev.team/vh/vh-srv-events/common"
)

func GetDBURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		url.QueryEscape(common.Config.DBUser),
		url.QueryEscape(common.Config.DBPass),
		common.Config.DBHost,
		common.Config.DBPort,
		url.QueryEscape(common.Config.DBName))
}

// sqlArgs collects query arguments; add returns the $n placeholder for the
// value it appends.
type sqlArgs []interface{}

func (a *sqlArgs) add(v interface{}) string {
	*a = append(*a, v)
	return "$" + strconv.Itoa(len(*a))
}

// columns collects column = value pairs for a dynamic INSERT or UPDATE.
// Column names are always literals from this package; values become $n args.
type columns struct {
	names []string
	args  sqlArgs
}

func (c *columns) set(name string, v interface{}) {
	c.names = append(c.names, name)
	c.args = append(c.args, v)
}

// setIf adds the column only when v is non-nil.
func setIf[T any](c *columns, name string, v *T) {
	if v != nil {
		c.set(name, *v)
	}
}

func (c *columns) empty() bool { return len(c.names) == 0 }

// insert returns INSERT INTO table (a,b) VALUES ($1,$2) and its args.
func (c *columns) insert(table string) (string, []interface{}) {
	nums := make([]string, len(c.names))
	for i := range c.names {
		nums[i] = "$" + strconv.Itoa(i+1)
	}
	return "INSERT INTO " + table + " (" + strings.Join(c.names, ",") + ") VALUES (" + strings.Join(nums, ",") + ")", c.args
}

// setClause returns a=$1,b=$2 for an UPDATE. Placeholders for a WHERE clause
// can then be added with c.args.add.
func (c *columns) setClause() string {
	sets := make([]string, len(c.names))
	for i, n := range c.names {
		sets[i] = n + "=$" + strconv.Itoa(i+1)
	}
	return strings.Join(sets, ",")
}

// update returns UPDATE table SET a=$1,b=$2 WHERE whereCol=$3 and its args.
func (c *columns) update(table, whereCol string, whereVal interface{}) (string, []interface{}) {
	sql := "UPDATE " + table + " SET " + c.setClause()
	sql += " WHERE " + whereCol + "=" + c.args.add(whereVal)
	return sql, c.args
}

func SyncDBStructInsertionAndMigrations() error {
	slog.Info("running db migrations")
	m, err := migrate.New(
		"file://./db/migrations", GetDBURL()+"?sslmode=disable")
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			slog.Info("no changes in migrations")
			return nil
		}
		return fmt.Errorf("migrate.Up: %w", err)
	}

	return nil
}
