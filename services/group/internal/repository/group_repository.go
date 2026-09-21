package repository

import (
	"context"
	"errors"

	"github.com/adk-saugat/rally/shared/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("group not found")
	ErrAlreadyMember = errors.New("already a member")
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *types.Group, creatorUserID string) error
	GetGroupByInviteCode(ctx context.Context, inviteCode string) (*types.Group, error)
	JoinGroup(ctx context.Context, groupID, userID string) error
	ListMembers(ctx context.Context, groupID string) ([]*types.User, error)
}

type groupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) GroupRepository {
	return &groupRepository{pool: pool}
}

func (repo *groupRepository) CreateGroup(ctx context.Context, group *types.Group, creatorUserID string) error {
	transaction, err := repo.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	err = transaction.QueryRow(ctx,
		"INSERT INTO groups (name, invite_code) VALUES ($1, $2) RETURNING id",
		group.Name, group.InviteCode).Scan(&group.ID)
	if err != nil {
		return err
	}

	_, err = transaction.Exec(ctx,
		"INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)",
		group.ID, creatorUserID)
	if err != nil {
		return err
	}

	return transaction.Commit(ctx)
}

func (repo *groupRepository) GetGroupByInviteCode(ctx context.Context, inviteCode string) (*types.Group, error) {
	var group types.Group
	err := repo.pool.QueryRow(ctx,
		"SELECT id, name, invite_code FROM groups WHERE invite_code = $1",
		inviteCode).Scan(&group.ID, &group.Name, &group.InviteCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (repo *groupRepository) JoinGroup(ctx context.Context, groupID, userID string) error {
	_, err := repo.pool.Exec(ctx, "INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)", groupID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyMember
		}
		return err
	}
	return nil
}

func (repo *groupRepository) ListMembers(ctx context.Context, groupID string) ([]*types.User, error) {
	rows, err := repo.pool.Query(ctx, "SELECT user_id FROM group_members WHERE group_id = $1", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*types.User, 0)
	for rows.Next() {
		var user types.User
		err := rows.Scan(&user.ID)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
