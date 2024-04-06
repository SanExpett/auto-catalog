package usecases

import (
	"encoding/json"
	"fmt"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/asaskevich/govalidator"
	"io"
)

var (
	ErrDecodePreCar = myerrors.NewError("Некорректный json машины")
)

func validatePreCar(r io.Reader) (*models.PreCar, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(r)

	preCar := &models.PreCar{}
	if err := decoder.Decode(preCar); err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, ErrDecodePreCar)
	}

	preCar.Trim()

	_, err = govalidator.ValidateStruct(preCar)
	if err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return preCar, nil
}

func ValidatePreCar(r io.Reader) (*models.PreCar, error) {
	preCar, err := validatePreCar(r)
	if err != nil {
		return nil, myerrors.NewError(err.Error())
	}

	return preCar, nil
}

func ValidatePartOfPreCar(r io.Reader) (*models.PreCar, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	preCar, err := validatePreCar(r)
	if preCar == nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	if err != nil {
		validationErrors := govalidator.ErrorsByField(err)

		for field, err := range validationErrors {
			if err != "non zero value required" {
				logger.Errorln(err)

				return nil, myerrors.NewError("%s error: %s", field, err)
			}
		}
	}

	return preCar, nil
}
