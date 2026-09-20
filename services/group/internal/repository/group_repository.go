package repository

import (
	"context"

	"github.com/adk-saugat/rally/shared/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupRepository interface {
	CreateGroup(ctx context.Context, group *types.Group) error
	JoinGroup(ctx context.Context, groupID, userID string) error
	ListMembers(ctx context.Context, groupID string) ([]*types.User, error)
}

type groupRepository struct {
	pool *pgxpool.Pool
}

func NewGroupRepository(pool *pgxpool.Pool) GroupRepository {
	return &groupRepository{pool: pool}
}

func (repo *groupRepository) CreateGroup(ctx context.Context, group *types.Group) error {
	_, err := repo.pool.Exec(ctx, "INSERT INTO groups (id, name, invite_code) VALUES ($1, $2, $3)", group.ID, group.Name, group.InviteCode)
	return err
}

func (repo *groupRepository) JoinGroup(ctx context.Context, groupID, userID string) error {
	_, err := repo.pool.Exec(ctx, "INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)", groupID, userID)
	return err
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