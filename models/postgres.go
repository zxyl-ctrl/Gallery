package models

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf(`host=%s port=%s user=%s password=%s dbname=%s sslmode=%s`,
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "baloo",
		Password: "junglebook",
		Database: "lenslocked",
		SSLMode:  "disable",
	}
}

func Open(config PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.String())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return db, nil
}

// goose使用全局变量，因此可以调用类似SetDialect，并且从这个函数调用所做的更改在之后调用Up函数会被保存下来
// 这样设计的目的实施的.go迁移文件更容易
// 为了便于在不存在目录的情况下执行Migrate，需要将.sql文件嵌入
func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("migrate:%w", err)
	}
	return nil
}

// ∵goose使用了全局变量，因此，MIgrateFS调用Migrate时，仍然能够使用设置的基础文件系统，一旦函数完成，可以将
// 延迟执行的代码将nil传递给函数，以取消设置的基础文件系统。这确保了我们在函数结束前撤销了所做的任何全局变量更改，
// 从而允许其他可能不使用我们的文件系统的迁移运行。
func MigrateFS(db *sql.DB, migrationsFS fs.FS, dir string) error {
	// In case the dir is an empty string, they probably meant the current directory and goose wants a period for that.
	if dir == "" {
		dir = "."
	}
	goose.SetBaseFS(migrationsFS)
	defer func() {
		// Ensure that we remove the FS on the off chance some other part of our app uses goose for migrations and doesn't want to use our FS.
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}
