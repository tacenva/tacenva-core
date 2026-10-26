package app

import (
	"github.com/tacenva/database"
	"github.com/tacenva/tacenva-core/accesscontrol"
	"github.com/tacenva/tacenva-core/auth"
	"github.com/tacenva/tacenva-core/internal/user"
	"github.com/tacenva/tacenva-core/permission"
	"github.com/tacenva/tacenva-core/sync"
	"gorm.io/gorm"
)

type Services struct {
	Auth          *auth.Service
	Permission    *permission.Service
	AccessControl *accesscontrol.Service
	SyncService   *sync.Service
}

func NewServices(
	sqliteDB *gorm.DB,
	tacenvaDB *database.DB,
) *Services {
	userService := user.NewService(
		user.NewRepository(sqliteDB),
	)

	permissionService := permission.NewService(
		permission.NewRepository(sqliteDB),
	)

	syncService := sync.NewService(
		sqliteDB,
		userService,
	)

	accessControlService := accesscontrol.NewService(
		userService,
		permissionService,
	)

	authService := auth.NewService(
		userService,
		permissionService,
	)

	return &Services{
		Auth:          authService,
		Permission:    permissionService,
		AccessControl: accessControlService,
		SyncService:   syncService,
	}
}
