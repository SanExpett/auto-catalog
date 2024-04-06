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
	ErrDecodePrePeople = myerrors.NewError("Некорректный json человека")
)

func ValidatePrePeople(r io.Reader) (*models.PrePeople, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(r)
	prePeople := &models.PrePeople{}
	if err := decoder.Decode(prePeople); err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, ErrDecodePrePeople)
	}

	prePeople.Trim()

	_, err = govalidator.ValidateStruct(prePeople)
	if err != nil {
		logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return prePeople, nil
}
