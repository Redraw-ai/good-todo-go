package dependency

import (
	"good-todo-go/internal/ent"
	"good-todo-go/internal/infrastructure/environment"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/presentation/public/controller"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/presentation/public/router"
	"good-todo-go/internal/usecase"

	"go.uber.org/dig"
)

func BuildContainer(client *ent.Client, cfg *environment.Config) (*dig.Container, error) {
	container := dig.New()

	// Provide ent client
	if err := container.Provide(func() *ent.Client { return client }); err != nil {
		return nil, err
	}

	// Provide config
	if err := container.Provide(func() *environment.Config { return cfg }); err != nil {
		return nil, err
	}

	// Provide JWT service
	if err := container.Provide(func(cfg *environment.Config) *pkg.JWTService {
		return pkg.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn, cfg.JWTRefreshExpiresIn)
	}); err != nil {
		return nil, err
	}

	// Provide UUID generator
	if err := container.Provide(func() pkg.IUUIDGenerator {
		return pkg.NewUUIDGenerator()
	}); err != nil {
		return nil, err
	}

	// Provide repositories
	if err := container.Provide(repository.NewAuthRepository); err != nil {
		return nil, err
	}

	// Provide usecases
	if err := container.Provide(usecase.NewAuthInteractor); err != nil {
		return nil, err
	}

	// Provide presenters
	if err := container.Provide(presenter.NewAuthPresenter); err != nil {
		return nil, err
	}

	// Provide controllers
	if err := container.Provide(controller.NewAuthController); err != nil {
		return nil, err
	}

	// Provide server
	if err := container.Provide(router.NewServer); err != nil {
		return nil, err
	}

	return container, nil
}
