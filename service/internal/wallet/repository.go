package wallet

import (
	"context"
	"errors"
	"service/internal/config"
	"service/internal/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type RepositoryInterface interface {
	Deposit(walletID string, amount int64, ctx context.Context) error
	Withdraw(walletID string, amount int64, ctx context.Context) error
	GetBalance(walletID string, ctx context.Context) (int64, error)
	Close()
}

type DBPool interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close()
}

type Repository struct {
	db  DBPool
	lg  logger.Logger
	ctx context.Context
}

const (
	maxconns = 2000
)

var errWalletid, errWithdraw = errors.New("walletid not found"), errors.New("insufficient funds or walletid not found")

func NewRepository(lg logger.Logger, ctx context.Context, cfg *config.ConfigAdr) RepositoryInterface {
	conf, err := pgxpool.ParseConfig(cfg.Database_url)
	if err != nil {
		lg.FatalCtx(ctx, "Could not parse database URL: ", err)
	}
	conf.MaxConns = maxconns

	pg, err := pgxpool.NewWithConfig(ctx, conf)
	if err != nil {
		lg.FatalCtx(ctx, "Could not create connection pool: ", err)
	}
	rep := new(Repository)
	rep.db = pg
	rep.lg = lg
	rep.ctx = ctx
	return rep
}

func (r *Repository) Close() {
	r.db.Close()
}

func (r *Repository) Deposit(walletID string, amount int64, ctx context.Context) error {
	r.ctx = ctx
	result, err := r.db.Exec(r.ctx, "UPDATE wallets SET balance = balance + $1 WHERE id = $2", amount, walletID)
	if err != nil {
		r.lg.ErrorCtx(r.ctx, "func deposit sql query failed")
		return err
	}
	rowsAffected := result.RowsAffected()

	if rowsAffected == 0 {
		r.lg.ErrorCtx(r.ctx, "func deposit walletid not found")
		return errWalletid
	}
	return err
}

func (r *Repository) Withdraw(walletID string, amount int64, ctx context.Context) error {
	r.ctx = ctx
	result, err := r.db.Exec(r.ctx, "UPDATE wallets SET balance = balance - $1 WHERE id = $2 AND balance >= $1", amount, walletID)
	if err != nil {
		r.lg.ErrorCtx(r.ctx, "func withdraw sql query failed")
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.lg.ErrorCtx(r.ctx, "func withdraw insufficient funds or walletid not found")
		return errWithdraw
	}
	return err
}

func (r *Repository) GetBalance(walletID string, ctx context.Context) (int64, error) {
	r.ctx = ctx
	var balance int64
	err := r.db.QueryRow(r.ctx, "SELECT balance FROM wallets WHERE id = $1", walletID).Scan(&balance)
	if err == pgx.ErrNoRows {
		r.lg.ErrorCtx(r.ctx, "func getbalance walletid not found")
		return 0, errWalletid
	} else if err != nil {
		r.lg.ErrorCtx(r.ctx, "Could not scan wallet")
		return 0, err
	}
	return balance, nil
}
