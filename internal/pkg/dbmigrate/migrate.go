// Package dbmigrate 提供基于 golang-migrate 的嵌入式数据库版本化迁移。
//
// 迁移文件存放在 migrations/ 子目录，以版本号命名：
//
//	000001_initial_schema.up.sql   / 000001_initial_schema.down.sql
//	000002_add_xxx.up.sql          / 000002_add_xxx.down.sql
//	...
//
// 每次应用启动时调用 RunMigrations，它会自动把还未执行过的迁移文件
// 按版本顺序应用到数据库，并在 schema_migrations 表中记录版本号。
package dbmigrate

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations 将所有未执行的 SQL 迁移文件应用到数据库。
// db 必须是已连接的 *sql.DB（从 gorm.DB.DB() 获取）。
// 如果数据库已是最新版本，则静默跳过，不返回错误。
func RunMigrations(db *sql.DB, dbName string, logger *zap.Logger) error {
	// 创建 iofs source
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("dbmigrate: create iofs source: %w", err)
	}

	// 创建 MySQL driver（使用已有连接，不新建）
	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: dbName,
	})
	if err != nil {
		return fmt.Errorf("dbmigrate: create mysql driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("dbmigrate: create migrator: %w", err)
	}

	logger.Info("running database migrations...")

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("database schema is already up to date")
			return nil
		}
		return fmt.Errorf("dbmigrate: apply migrations: %w", err)
	}

	version, dirty, verErr := m.Version()
	if verErr == nil {
		logger.Info("database migrations applied successfully",
			zap.Uint("schema_version", version),
			zap.Bool("dirty", dirty),
		)
	}

	return nil
}
