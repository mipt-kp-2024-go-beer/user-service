package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/jmoiron/sqlx"
	"golang.org/x/sync/errgroup"
)

type App struct {
	config  *Config
	public  *http.ServeMux
	private *http.ServeMux
}

func New(ctx context.Context, config *Config) (*App, error) {
	pub := http.NewServeMux()
	priv := http.NewServeMux()
	return &App{
		config:  config,
		public:  pub,
		private: priv,
	}, nil
}

func (a *App) Setup(ctx context.Context, dsn string) error {
	db, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return err
	}

	//store := fridgeStore.NewStorage(db)
	// store := sqlite.NewStorage(db)

	//service := fridge.NewAppService(store)
	//handler := fridge.NewHandler(a.router, service)
	//handler.Register()

	// shelfService := shelf.NewAppService(store)

	return nil
}

func (a *App) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	errs, ctx := errgroup.WithContext(ctx)

	log.Println("starting web server on port %s", a.config.PublicPort)

	errs.Go(func() error {
		http.ListenAndServe(a.public)
		if err := a.http.ListenAndServe(); err != nil {
			return fmt.Errorf("listen and serve error: %w", err)
		}
		return nil
	})

	//<-ctx.Done()

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
