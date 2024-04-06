package usecases

import (
	"context"
	"fmt"
	peoplerepo "github.com/SanExpett/auto-catalog/internal/people/repository"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"go.uber.org/zap"
	"io"
)

var _ IPeopleStorage = (*peoplerepo.PeopleStorage)(nil)

type IPeopleStorage interface {
	AddPerson(ctx context.Context, prePeople *models.PrePeople) (*models.People, error)
	GetPerson(ctx context.Context, peopleID uint64) (*models.People, error)
	DeletePerson(ctx context.Context, personID uint64) error
}

type PeopleService struct {
	storage IPeopleStorage
	logger  *zap.SugaredLogger
}

func NewPeopleService(peopleStorage IPeopleStorage) (*PeopleService, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &PeopleService{storage: peopleStorage, logger: logger}, nil
}

func (p *PeopleService) AddPerson(ctx context.Context, r io.Reader) (*models.People, error) {
	prePeople, err := ValidatePrePeople(r)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	people, err := p.storage.AddPerson(ctx, prePeople)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return people, nil
}

func (p *PeopleService) GetPerson(ctx context.Context, peopleID uint64) (*models.People, error) {
	people, err := p.storage.GetPerson(ctx, peopleID)
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	people.Sanitize()

	return people, nil
}

func (p *PeopleService) DeletePerson(ctx context.Context, personID uint64) error {
	err := p.storage.DeletePerson(ctx, personID)
	if err != nil {
		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}
