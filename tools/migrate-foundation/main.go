package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"go.yaml.in/yaml/v3"
)

type sourceConfig struct {
	Database struct {
		MySQL struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Database string `yaml:"database"`
		} `yaml:"mysql"`
	} `yaml:"database"`
}

type tableSpec struct {
	name    string
	columns []string
	selects []string
}

var baseColumns = []string{"id", "create_time", "update_time", "delete_time"}

func table(name string, fields ...string) tableSpec {
	return tableSpec{name: name, columns: append(append([]string{}, baseColumns...), fields...)}
}

var tables = []tableSpec{
	table("sys_role", "role_name", "role_code", "description", "status"),
	table("sys_menu", "menu_name", "parent_id", "menu_type", "route_path", "sort", "status", "component", "icon", "hide_in_menu"),
	table("sys_permission", "perm_name", "module", "perm_key", "request_method", "api_path", "status"),
	{
		name:    "sys_admin",
		columns: append(append([]string{}, baseColumns...), "username", "password", "admin_name", "role_id", "avatar", "status", "last_login_time"),
		selects: []string{"id", "create_time", "update_time", "delete_time", "username", "password", "COALESCE(NULLIF(admin_name, ''), adminname)", "role_id", "avatar", "status", "last_login_time"},
	},
	table("sys_role_menu", "role_id", "menu_id"),
	table("sys_role_perm", "role_id", "perm_id"),
	table("tag", "tag_name", "status", "is_hot"),
	table("sensitive_word", "word", "status"),
	table("msg_template", "template_name", "msg_type", "content_template", "fixed_tip"),
}

func main() {
	if len(os.Args) != 2 {
		fail(fmt.Errorf("usage: migrate-foundation <local-old-config.yaml>"))
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fail(err)
	}
	var cfg sourceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fail(err)
	}
	source := cfg.Database.MySQL
	if source.Host == "" || source.Username == "" || source.Database == "" {
		fail(fmt.Errorf("incomplete source MySQL configuration"))
	}
	dsn := (&mysql.Config{
		User: source.Username, Passwd: source.Password, Net: "tcp",
		Addr: source.Host + ":" + strconv.Itoa(source.Port), DBName: source.Database,
		AllowNativePasswords: true,
	}).FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fail(err)
	}
	defer db.Close()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		fail(err)
	}
	defer tx.Rollback()

	// Keep the target transaction open until every source table has been read.
	fmt.Fprintln(os.Stdout, "SET NAMES utf8mb4;\nSTART TRANSACTION;")
	for _, spec := range tables {
		if err := exportTable(ctx, tx, os.Stdout, spec); err != nil {
			fail(fmt.Errorf("%s: %w", spec.name, err))
		}
	}
	fmt.Fprintln(os.Stdout, "COMMIT;")
}

func exportTable(ctx context.Context, tx *sql.Tx, out io.Writer, spec tableSpec) error {
	selects := spec.selects
	if len(selects) == 0 {
		selects = spec.columns
	}
	rows, err := tx.QueryContext(ctx, "SELECT "+strings.Join(selects, ",")+" FROM "+spec.name+" ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		values := make([]sql.RawBytes, len(spec.columns))
		dest := make([]any, len(values))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		encoded := make([]string, len(values))
		for i, value := range values {
			if value == nil {
				encoded[i] = "NULL"
			} else {
				encoded[i] = "CONVERT(0x" + hex.EncodeToString(value) + " USING utf8mb4)"
			}
		}
		if _, err := fmt.Fprintf(out, "INSERT INTO %s (%s) VALUES (%s);\n", spec.name, strings.Join(spec.columns, ","), strings.Join(encoded, ",")); err != nil {
			return err
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s: %d rows\n", spec.name, count)
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
