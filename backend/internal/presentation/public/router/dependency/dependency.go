package dependency

import (
	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/infrastructure/environment"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/presentation/public/controller"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/usecase"

	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	// environment
	container.Provide(environment.LoadConfig)

	// infrastructure
	container.Provide(database.NewEntClient)

	// pkg
	container.Provide(func(cfg *environment.Config) *pkg.JWTService {
		return pkg.NewJWTService(cfg.JWTSecret, cfg.JWTExpiresIn, cfg.JWTRefreshExpiresIn)
	})
	container.Provide(pkg.NewUUIDGenerator)

	// repository
	container.Provide(repository.NewAuthRepository)
	container.Provide(repository.NewUserRepository)
	container.Provide(repository.NewTodoRepository)

	// usecase
	container.Provide(usecase.NewAuthInteractor)
	container.Provide(usecase.NewUserInteractor)
	container.Provide(usecase.NewTodoInteractor)

	// presenter
	container.Provide(presenter.NewAuthPresenter)
	container.Provide(presenter.NewUserPresenter)
	container.Provide(presenter.NewTodoPresenter)

	// controller
	container.Provide(controller.NewAuthController)
	container.Provide(controller.NewUserController)
	container.Provide(controller.NewTodoController)

	return container
}
