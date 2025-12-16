package router

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"good-todo-go/internal/ent"
	"good-todo-go/internal/infrastructure/environment"
	"good-todo-go/internal/pkg"
	"good-todo-go/internal/presentation/public/api"
	"good-todo-go/internal/presentation/public/controller"
	"good-todo-go/internal/presentation/public/router/dependency"
	"good-todo-go/internal/presentation/public/router/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	env            *environment.Config
	jwtService     *pkg.JWTService
	authController *controller.AuthController
	userController *controller.UserController
}

func NewServer(
	env *environment.Config,
	jwtService *pkg.JWTService,
	authController *controller.AuthController,
	userController *controller.UserController,
) *Server {
	return &Server{
		env:            env,
		jwtService:     jwtService,
		authController: authController,
		userController: userController,
	}
}

func NewRouter() (*echo.Echo, *environment.Config, *ent.Client, error) {
	e := echo.New()

	// ミドルウェア設定
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// DIコンテナを作成
	container := dependency.BuildContainer()
	container.Provide(NewServer)

	var (
		server    *Server
		client    *ent.Client
		jwtSvc    *pkg.JWTService
	)

	if err := container.Invoke(func(s *Server) {
		server = s
	}); err != nil {
		return nil, nil, nil, err
	}

	if err := container.Invoke(func(c *ent.Client) {
		client = c
	}); err != nil {
		return nil, nil, nil, err
	}

	if err := container.Invoke(func(j *pkg.JWTService) {
		jwtSvc = j
	}); err != nil {
		return nil, nil, nil, err
	}

	// JWT認証ミドルウェア
	e.Use(middleware.JWTAuthMiddleware(jwtSvc))

	// 依存解決したハンドラーをルーティングに登録
	api.RegisterHandlers(e, server)

	// グレースフルシャットダウンを仕込む
	gracefulShutdown(e)

	return e, server.env, client, nil
}

func gracefulShutdown(e *echo.Echo) {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-shutdownCh
		log.Printf("Received signal %s, shutting down...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			log.Printf("failed to shut down echo server gracefully: %v", err)
			if err := e.Close(); err != nil {
				log.Printf("failed to force close echo server: %v", err)
			}
		}
	}()
}
