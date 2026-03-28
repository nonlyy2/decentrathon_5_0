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
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)`,
		"admin@invisionu.kz", string(hash), "admin")
	if err != nil {
		return err
	}

	log.Println("Default admin user created (admin@invisionu.kz / admin123)")
	return nil
}
