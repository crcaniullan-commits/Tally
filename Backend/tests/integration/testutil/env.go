package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	accesskeys "github.com/crcaniullan-commits/Tally/internal/accessKeys"
	"github.com/crcaniullan-commits/Tally/internal/auth"
	"github.com/crcaniullan-commits/Tally/internal/categories"
	"github.com/crcaniullan-commits/Tally/internal/dbtx"
	"github.com/crcaniullan-commits/Tally/internal/debtors"
	"github.com/crcaniullan-commits/Tally/internal/empresas"
	errorhandler "github.com/crcaniullan-commits/Tally/internal/error"
	"github.com/crcaniullan-commits/Tally/internal/expenses"
	"github.com/crcaniullan-commits/Tally/internal/goals"
	"github.com/crcaniullan-commits/Tally/internal/incomes"
	"github.com/crcaniullan-commits/Tally/internal/middleware"
	"github.com/crcaniullan-commits/Tally/internal/users"
	"github.com/go-chi/chi/v5"
	chimit "github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

const (
	defaultDSN = "postgres://postgres:tally_database_env@localhost/tally_database_env?sslmode=disable"
	jwtSecret  = "integration-test-secret"
	jwtIssuer  = "Tally"
	jwtExp     = 24 * time.Hour
)

type TestEnv struct {
	DB     *sql.DB
	Router http.Handler
}

var current *TestEnv

func GetEnv() *TestEnv {
	if current == nil {
		panic("testutil: el entorno no fue inicializado; llamá a SetupSuite desde TestMain")
	}
	return current
}

func newEnv() (*TestEnv, func(), error) {
	db, cleanup, err := prepareDB()
	if err != nil {
		return nil, nil, err
	}

	return &TestEnv{DB: db, Router: buildRouter(db)}, cleanup, nil
}

func buildRouter(db *sql.DB) http.Handler {
	logger := zap.NewNop().Sugar()

	middl := middleware.NewAuthMiddleware(
		errorhandler.NewErrorResponse(logger),
		auth.NewJWTAuthenticator(jwtSecret, jwtIssuer, jwtIssuer),
		users.NewStorage(db),
	)

	empresaRoute, finder := empresas.InitModule(db, logger, middl)
	emMiddl := middleware.NewEmpresaMiddleware(logger, finder)
	transactor := dbtx.NewTransactor(db)

	r := chi.NewRouter()
	r.Use(
		chimit.RequestID,
		chimit.ClientIPFromHeader("X-Real-IP"),
		chimit.Recoverer,
	)

	r.Route("/v1", func(r chi.Router) {
		auth.InitModule(r, db, logger, transactor, jwtSecret, jwtIssuer, jwtExp)

		r.Route("/app", func(r chi.Router) {
			r.Use(middl.AuthTokenMiddleware)

			users.InitModule(r, db, logger, middl)
			accesskeys.InitModule(r, db, logger, middl, noopSender{})

			r.Route("/empresas", func(r chi.Router) {
				empresaRoute.Collection(r)

				r.Route("/{empresaID}", func(r chi.Router) {
					r.Use(emMiddl.EmpresaOfUser)

					empresaRoute.Item(r)
					incomes.InitModule(r, db, logger, middl)
					expenses.InitModule(r, db, logger, middl)
					categories.InitModule(r, db, logger, middl)
					goals.InitModule(r, db, logger, middl)
					debtors.InitModule(r, db, logger, middl)
				})
			})
		})
	})

	return r
}

type noopSender struct{}

func (noopSender) SendActivationCode(context.Context, string, string, time.Time) error {
	return nil
}

func prepareDB() (*sql.DB, func(), error) {
	if addr := os.Getenv("TEST_DB_ADDR"); addr != "" {
		db, err := sql.Open("postgres", addr)
		if err != nil {
			return nil, nil, fmt.Errorf("abriendo TEST_DB_ADDR: %w", err)
		}
		if err := db.Ping(); err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("conectando a TEST_DB_ADDR: %w", err)
		}
		if err := ensureSchema(db); err != nil {
			db.Close()
			return nil, nil, err
		}
		return db, func() { db.Close() }, nil
	}

	base := os.Getenv("DB_ADDR")
	if base == "" {
		base = defaultDSN
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, nil, fmt.Errorf("DB_ADDR no es una URL válida: %w", err)
	}

	adminURL := *baseURL
	adminURL.Path = "/postgres"

	admin, err := sql.Open("postgres", adminURL.String())
	if err != nil {
		return nil, nil, fmt.Errorf("abriendo conexión administrativa: %w", err)
	}
	if err := admin.Ping(); err != nil {
		admin.Close()
		return nil, nil, fmt.Errorf("conectando a la base administrativa: %w", err)
	}

	name := fmt.Sprintf("tally_it_%d", os.Getpid())
	quoted := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`

	terminate := func() {
		_, _ = admin.Exec(
			`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`,
			name,
		)
	}

	terminate()
	if _, err := admin.Exec(`DROP DATABASE IF EXISTS ` + quoted); err != nil {
		admin.Close()
		return nil, nil, fmt.Errorf("borrando base de test previa: %w", err)
	}
	if _, err := admin.Exec(`CREATE DATABASE ` + quoted); err != nil {
		admin.Close()
		return nil, nil, fmt.Errorf("creando base de test: %w", err)
	}

	testURL := *baseURL
	testURL.Path = "/" + name

	db, err := sql.Open("postgres", testURL.String())
	if err != nil {
		admin.Close()
		return nil, nil, fmt.Errorf("abriendo base de test: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		admin.Close()
		return nil, nil, fmt.Errorf("conectando a la base de test: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		db.Close()
		admin.Close()
		return nil, nil, err
	}
	if err := ensureGlobalCategories(db); err != nil {
		db.Close()
		admin.Close()
		return nil, nil, err
	}

	cleanup := func() {
		db.Close()
		terminate()
		_, _ = admin.Exec(`DROP DATABASE IF EXISTS ` + quoted)
		admin.Close()
	}

	return db, cleanup, nil
}

func ensureSchema(db *sql.DB) error {
	var hasUsers bool

	if err := db.QueryRow(`SELECT to_regclass('public.users') IS NOT NULL`).Scan(&hasUsers); err != nil {
		return fmt.Errorf("verificando esquema: %w", err)
	}

	if !hasUsers {
		if err := applyMigrations(db); err != nil {
			return err
		}
	}

	return ensureGlobalCategories(db)
}

func ensureGlobalCategories(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT INTO categories (nombre, empresa_id)
		SELECT nombre, NULL
		FROM (VALUES ('Ventas'), ('Servicios'), ('Pedidos')) AS defaults(nombre)
		WHERE NOT EXISTS (
			SELECT 1 FROM categories c
			WHERE c.empresa_id IS NULL AND c.nombre = defaults.nombre
		)
	`)
	if err != nil {
		return fmt.Errorf("asegurando categorías globales: %w", err)
	}

	return nil
}

func applyMigrations(db *sql.DB) error {
	dir, err := migrationsDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("leyendo migraciones: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("leyendo migración %s: %w", name, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("aplicando migración %s: %w", name, err)
		}
	}

	return nil
}

func migrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("no se pudo ubicar el archivo de testutil")
	}

	dir := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "migrations"))
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("no se encontró migrations en %s: %w", dir, err)
	}

	return dir, nil
}
