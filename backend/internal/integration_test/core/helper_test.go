package integration_test

import (
	"good-todo-go/internal/ent"
	"good-todo-go/internal/infrastructure/database"
	"good-todo-go/internal/infrastructure/repository"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/pkg/cerror"
	"good-todo-go/internal/presentation/public/controller"
	"good-todo-go/internal/presentation/public/presenter"
	"good-todo-go/internal/presentation/public/router/context_keys"
	"good-todo-go/internal/usecase"

	"github.com/labstack/echo/v4"
)

// TestDependencies holds all the dependencies for integration tests
type TestDependencies struct {
	Client         *ent.Client
	AuthController *controller.AuthController
	TodoController *controller.TodoController
	UserController *controller.UserController
	JWTService     *pkg.JWTService
}

// BuildTestDependencies creates all dependencies for integration tests
func BuildTestDependencies(client *ent.Client) *TestDependencies {
	// Repositories
	authRepo := repository.NewAuthRepository(client)
	todoRepo := repository.NewTodoRepository(client)
	userRepo := repository.NewUserRepository(client)

	// Services
	uuidGen := pkg.NewUUIDGenerator()
	jwtService := pkg.NewJWTService("test-secret-key-for-integration-tests", 3600, 86400)

	// Usecases
	authInteractor := usecase.NewAuthInteractor(authRepo, jwtService, uuidGen)
	todoInteractor := usecase.NewTodoInteractor(todoRepo, userRepo, uuidGen)
	userInteractor := usecase.NewUserInteractor(userRepo)

	// Presenters
	authPresenter := presenter.NewAuthPresenter()
	todoPresenter := presenter.NewTodoPresenter()
	userPresenter := presenter.NewUserPresenter()

	// Controllers
	authController := controller.NewAuthController(authInteractor, authPresenter)
	todoController := controller.NewTodoController(todoInteractor, todoPresenter)
	userController := controller.NewUserController(userInteractor, userPresenter)

	return &TestDependencies{
		Client:         client,
		AuthController: authController,
		TodoController: todoController,
		UserController: userController,
		JWTService:     jwtService,
	}
}

// SetupEcho creates a new Echo instance with custom error handler
func SetupEcho() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = cerror.CustomHTTPErrorHandler
	return e
}

// SetAuthContext sets authentication context for the Echo context
func SetAuthContext(c echo.Context, userID, tenantID string) {
	c.Set(context_keys.UserIDContextKey, userID)
	c.Set(context_keys.TenantIDContextKey, tenantID)

	// Also set tenant ID in request context for repository layer
	ctx := database.WithTenantID(c.Request().Context(), tenantID)
	c.SetRequest(c.Request().WithContext(ctx))
}
