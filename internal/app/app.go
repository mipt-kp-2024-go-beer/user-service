package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	users "github.com/mipt-kp-2024-go-beer/user-service/internal"
	"github.com/mipt-kp-2024-go-beer/user-service/internal/storage/memory"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config  *Config
	open    *http.ServeMux
	secret  *http.ServeMux
	public  *http.Server
	private *http.Server
}

func New(ctx context.Context, config *Config) (*App, error) {
	open := http.NewServeMux()
	secret := http.NewServeMux()
	return &App{
		config:  config,
		open:    open,
		secret:  secret,
		public:  &http.Server{Addr: net.JoinHostPort(config.Host, config.PublicPort), Handler: open},
		private: &http.Server{Addr: net.JoinHostPort(config.Host, config.PrivatePort), Handler: secret},
	}, nil
}

func (a *App) Setup(ctx context.Context, dsn string) error {
	//db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	//if err != nil {
	//	return err
	//}

	//store := fridgeStore.NewStorage(db)
	// store := sqlite.NewStorage(db)
	store := memory.NewStorage()

	service := users.NewAppService(store)
	handler := users.NewHandler(service, a.open, a.secret)
	handler.Register()

	ID, err := service.NewUser(ctx, users.User{Login: a.config.Login, Password: a.config.Password, Permissions: 0x1111111})
	println(ID)
	if err != nil {
		fmt.Errorf("%w", err)
	}
	// shelfService := shelf.NewAppService(store)

	return nil
}

func (a *App) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errs, ctx := errgroup.WithContext(ctx)

	errs.Go(func() error {
		log.Println("starting web server on port", a.config.PublicPort)

		go func() {
			if err := a.public.ListenAndServe(); err != nil {
				panic("ListenAndServe: " + err.Error())
			}
		}()

		log.Println("starting web server on port", a.config.PrivatePort)

		go func() {
			if err := a.private.ListenAndServe(); err != nil {
				panic("ListenAndServe: " + err.Error())
			}
		}()

		return nil
	})

	<-ctx.Done()

	a.public.Shutdown(ctx)
	a.private.Shutdown(ctx)

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	//stop()
	//log.Println("shutting down gracefully")

	// Perform application shutdown with a maximum timeout of 5 seconds.
	//timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	//if err := a.http.Shutdown(timeoutCtx); err != nil {
	//	log.Println(err.Error())
	//}

	return nil
}
