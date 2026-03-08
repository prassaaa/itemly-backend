package usecase

import (
	"context"
	"log/slog"

	"github.com/prassaaa/itemly-backend/internal/repository"
	"github.com/redis/go-redis/v9"
)

type redisPermissionUsecase struct {
	permRepo repository.PermissionRepository
	rdb      *redis.Client
}

func NewRedisPermissionUsecase(permRepo repository.PermissionRepository, rdb *redis.Client) PermissionUsecase {
	return &redisPermissionUsecase{
		permRepo: permRepo,
		rdb:      rdb,
	}
}

func (u *redisPermissionUsecase) LoadPermissions() error {
	ctx := context.Background()

	// Delete existing permission keys
	iter := u.rdb.Scan(ctx, 0, "perm:*", 100).Iterator()
	var keysToDelete []string
	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keysToDelete) > 0 {
		if err := u.rdb.Del(ctx, keysToDelete...).Err(); err != nil {
			return err
		}
	}

	rolePerms, err := u.permRepo.GetAllRolePermissions()
	if err != nil {
		return err
	}

	pipe := u.rdb.Pipeline()
	for _, rp := range rolePerms {
		pipe.HSet(ctx, "perm:"+string(rp.Role), rp.Permission.Name, "1")
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (u *redisPermissionUsecase) HasPermission(role string, permissionName string) bool {
	ctx := context.Background()
	exists, err := u.rdb.HExists(ctx, "perm:"+role, permissionName).Result()
	if err != nil {
		slog.Error("failed to check permission in Redis", "role", role, "permission", permissionName, "error", err)
		return false // fail closed for security
	}
	return exists
}

func (u *redisPermissionUsecase) GetPermissionsByRole(role string) []string {
	ctx := context.Background()
	keys, err := u.rdb.HKeys(ctx, "perm:"+role).Result()
	if err != nil {
		slog.Error("failed to get permissions from Redis", "role", role, "error", err)
		return nil
	}
	return keys
}
