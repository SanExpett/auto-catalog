package usecases

import (
	"context"
	"fmt"
	carrepo "github.com/SanExpett/auto-catalog/internal/car/repository"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/SanExpett/auto-catalog/pkg/utils"
	"go.uber.org/zap"
	"io"
)

var _ ICarStorage = (*carrepo.CarStorage)(nil)

type ICarStorage interface {
	AddCar(ctx context.Context, preCar *models.PreCar) (*models.Car, error)
	GetCar(ctx context.Context, CarID uint64) (*models.Car, error)
	DeleteCar(ctx context.Context, carID uint64) error
	UpdateCar(ctx context.Context, carID uint64, updateFields map[string]interface{}) error
	GetCarsList(ctx context.Context, limit uint64, offset uint64, model string, mark string, ownerID uint64,
		sortByYearType uint64) ([]*models.Car, error)
}

type CarService struct {
	storage ICarStorage
	logger  *zap.SugaredLogger
}

func NewCarService(CarStorage ICarStorage) (*CarService, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &CarService{storage: CarStorage, logger: logger}, nil
}

func (p *CarService) AddCar(ctx context.Context, r io.Reader) (*models.Car, error) {
	preCar, err := ValidatePreCar(r)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	car, err := p.storage.AddCar(ctx, preCar)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return car, nil
}

func (p *CarService) GetCar(ctx context.Context, carID uint64) (*models.Car, error) {
	car, err := p.storage.GetCar(ctx, carID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	car.Sanitize()

	return car, nil
}

func (c *CarService) DeleteCar(ctx context.Context, carID uint64) error {
	err := c.storage.DeleteCar(ctx, carID)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (c *CarService) UpdateCar(ctx context.Context, r io.Reader, isPartialUpdate bool, carID uint64) error {
	var preCar *models.PreCar

	var err error

	if isPartialUpdate {
		preCar, err = ValidatePartOfPreCar(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	} else {
		preCar, err = ValidatePreCar(r)
		if err != nil {
			return fmt.Errorf(myerrors.ErrTemplate, err)
		}
	}

	updateFieldsMap := utils.StructToMap(preCar)

	err = c.storage.UpdateCar(ctx, carID, updateFieldsMap)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (c *CarService) GetCarsList(ctx context.Context, limit uint64, offset uint64, model string, mark string,
	ownerID uint64, sortByYearType uint64,
) ([]*models.Car, error) {
	cars, err := c.storage.GetCarsList(ctx, limit, offset, model, mark, ownerID, sortByYearType)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	for _, car := range cars {
		car.Sanitize()
	}

	return cars, nil
}
