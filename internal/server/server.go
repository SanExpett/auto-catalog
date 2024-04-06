package server

import (
	"context"
	carrepo "github.com/SanExpett/auto-catalog/internal/car/repository"
	carusecases "github.com/SanExpett/auto-catalog/internal/car/usecases"
	peoplerepo "github.com/SanExpett/auto-catalog/internal/people/repository"
	peopleusecases "github.com/SanExpett/auto-catalog/internal/people/usecases"
	"github.com/SanExpett/auto-catalog/internal/server/delivery/mux"
	"github.com/SanExpett/auto-catalog/internal/server/repository"
	"github.com/SanExpett/auto-catalog/pkg/config"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"net/http"
	"strings"
	"time"
)

const (
	basicTimeout = 10 * time.Second
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(config *config.Config) error {
	baseCtx := context.Background()

	pool, err := repository.NewPgxPool(baseCtx, config.URLDataBase)
	if err != nil {
		return err //nolint:wrapcheck
	}

	logger, err := my_logger.New(strings.Split(config.OutputLogPath, " "),
		strings.Split(config.ErrorOutputLogPath, " "))
	if err != nil {
		return err //nolint:wrapcheck
	}

	defer logger.Sync()

	peopleStorage, err := peoplerepo.NewPeopleStorage(pool)
	if err != nil {
		return err
	}
	peopleService, err := peopleusecases.NewPeopleService(peopleStorage)
	if err != nil {
		return err
	}

	carStorage, err := carrepo.NewCarStorage(pool)
	if err != nil {
		return err
	}
	carService, err := carusecases.NewCarService(carStorage)
	if err != nil {
		return err
	}

	handler, err := mux.NewMux(baseCtx, mux.NewConfigMux(config.AllowOrigin,
		config.Schema, config.PortServer), peopleService, carService, logger)
	if err != nil {
		return err
	}

	s.httpServer = &http.Server{ //nolint:exhaustruct
		Addr:           ":" + config.PortServer,
		Handler:        handler,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes,
		ReadTimeout:    basicTimeout,
		WriteTimeout:   basicTimeout,
	}

	logger.Infof("Start server:%s", config.PortServer)

	return s.httpServer.ListenAndServe() //nolint:wrapcheck
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx) //nolint:wrapcheck
}
