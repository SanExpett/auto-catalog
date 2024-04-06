package mux

import (
	"context"
	"github.com/SanExpett/auto-catalog/pkg/middleware"
	"net/http"

	cardelivery "github.com/SanExpett/auto-catalog/internal/car/delivery"
	peopledelivery "github.com/SanExpett/auto-catalog/internal/people/delivery"

	"go.uber.org/zap"
)

type ConfigMux struct {
	addrOrigin string
	schema     string
	portServer string
}

func NewConfigMux(addrOrigin string, schema string, portServer string) *ConfigMux {
	return &ConfigMux{
		addrOrigin: addrOrigin,
		schema:     schema,
		portServer: portServer,
	}
}

func NewMux(ctx context.Context, configMux *ConfigMux, peopleService peopledelivery.IPeopleService,
	carService cardelivery.ICarService, logger *zap.SugaredLogger,
) (http.Handler, error) {
	router := http.NewServeMux()

	peopleHandler, err := peopledelivery.NewPeopleHandler(peopleService)
	if err != nil {
		return nil, err
	}

	carHandler, err := cardelivery.NewCarHandler(carService)
	if err != nil {
		return nil, err
	}

	router.Handle("/api/v1/people/add", middleware.Context(ctx,
		middleware.SetupCORS(peopleHandler.AddPeopleHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/people/get", middleware.Context(ctx,
		middleware.SetupCORS(peopleHandler.GetPeopleHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/people/delete", middleware.Context(ctx,
		middleware.SetupCORS(peopleHandler.DeletePeopleHandler, configMux.addrOrigin, configMux.schema)))

	router.Handle("/api/v1/car/add", middleware.Context(ctx,
		middleware.SetupCORS(carHandler.AddCarHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/car/get", middleware.Context(ctx,
		middleware.SetupCORS(carHandler.GetCarHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/car/delete", middleware.Context(ctx,
		middleware.SetupCORS(carHandler.DeleteCarHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/car/update", middleware.Context(ctx,
		middleware.SetupCORS(carHandler.UpdateCarHandler, configMux.addrOrigin, configMux.schema)))
	router.Handle("/api/v1/car/get_list", middleware.Context(ctx,
		middleware.SetupCORS(carHandler.GetCarsListHandler, configMux.addrOrigin, configMux.schema)))

	mux := http.NewServeMux()
	mux.Handle("/", middleware.Panic(router, logger))

	return mux, nil
}
