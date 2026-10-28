package app

import (
	"github.com/tacenva/database"
	"github.com/tacenva/tacenva-core/accesscontrol"
	"github.com/tacenva/tacenva-core/auth"
	"github.com/tacenva/tacenva-core/internal/accessgroup"
	"github.com/tacenva/tacenva-core/internal/accessunit"
	"github.com/tacenva/tacenva-core/internal/user"
	"github.com/tacenva/tacenva-core/sync"
	"gorm.io/gorm"
)

type Services struct {
	Auth          *auth.Service
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

	accessunitService := accessunit.NewService(
		accessunit.NewRepository(sqliteDB),
	)

	accessgroupService := accessgroup.NewService(
		accessgroup.NewRepository(sqliteDB),
		accessunitService.CreateUnit,
	)

	syncService := sync.NewService(
		sqliteDB,
		userService,
	)

	accessControlService := accesscontrol.NewService(
		userService,
		accessunitService,
		accessgroupService,
	)

	authService := auth.NewService(
		userService,
		accessunitService,
		accessgroupService,
		"",
	)

	return &Services{
		Auth:          authService,
		AccessControl: accessControlService,
		SyncService:   syncService,
	}
}
