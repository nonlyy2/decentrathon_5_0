package seed

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func SeedAdminUser(pool *pgxpool.Pool) error {
	ctx := context.Background()

	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO users (email, password_hash, role, full_name) VALUES ($1, $2, $3, $4)`,
		"admin@invisionu.kz", string(hash), "superadmin", "Super Admin")
	if err != nil {
		return err
	}

	log.Println("Default superadmin user created (admin@invisionu.kz / admin123)")
	return nil
}

// апгрейд legacy admin → superadmin если нужно
func EnsureSuperAdminRole(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx,
		`UPDATE users SET role = 'superadmin' WHERE email = 'admin@invisionu.kz' AND role = 'admin'`)
	return err
}
