package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/SanExpett/auto-catalog/internal/server/repository"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

var (
	ErrPeopleNotFound       = myerrors.NewError("Этот человек не найден")
	ErrNoAffectedPeopleRows = myerrors.NewError("Не получилось обновить данные человека")

	NameSeqPeople = pgx.Identifier{"public", "people_id_seq"} //nolint:gochecknoglobals
)

type PeopleStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewPeopleStorage(pool *pgxpool.Pool) (*PeopleStorage, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &PeopleStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

func (p *PeopleStorage) selectCreatedAtByPeopleID(ctx context.Context, tx pgx.Tx, peopleID uint64,
) (time.Time, error) {
	SQLSelectCreatedAtByPeopleID := `SELECT created_at FROM public."people" WHERE id=$1`

	var createdAt time.Time

	createdAtRow := tx.QueryRow(ctx, SQLSelectCreatedAtByPeopleID, peopleID)
	if err := createdAtRow.Scan(&createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf(myerrors.ErrTemplate, ErrPeopleNotFound)
		}

		p.logger.Errorf("error with PeopleId=%d: %+v", peopleID, err)

		return time.Time{}, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return createdAt, nil
}

func (p *PeopleStorage) insertPeople(ctx context.Context, tx pgx.Tx, prePeople *models.PrePeople) error {
	SQLInsertPeople := `INSERT INTO public."people"(name, surname, patronymic) VALUES($1, $2, $3)`
	_, err := tx.Exec(ctx, SQLInsertPeople, prePeople.Name, prePeople.Surname, prePeople.Patronymic)

	if err != nil {
		p.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (p *PeopleStorage) AddPerson(ctx context.Context, prePeople *models.PrePeople) (*models.People, error) {
	people := &models.People{Name: prePeople.Name, Surname: prePeople.Surname, Patronymic: prePeople.Patronymic}

	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		err := p.insertPeople(ctx, tx, prePeople)
		if err != nil {
			return err
		}

		lastPeopleID, err := repository.GetLastValSeq(ctx, tx, NameSeqPeople)
		if err != nil {
			return err
		}

		people.ID = lastPeopleID

		createdAt, err := p.selectCreatedAtByPeopleID(ctx, tx, lastPeopleID)
		if err != nil {
			return err
		}

		people.CreatedAt = createdAt

		return err
	})
	if err != nil {
		p.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return people, nil
}

func (p *PeopleStorage) selectPeopleByID(ctx context.Context, tx pgx.Tx, peopleID uint64,
) (*models.People, error) {
	SQLSelectPeople := `SELECT name, surname, patronymic, created_at FROM public."people" WHERE id=$1`
	people := &models.People{ID: peopleID} //nolint:exhaustruct

	peopleRow := tx.QueryRow(ctx, SQLSelectPeople, peopleID)
	if err := peopleRow.Scan(&people.Name, &people.Surname, &people.Patronymic, &people.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(myerrors.ErrTemplate, ErrPeopleNotFound)
		}

		p.logger.Errorf("error with PeopleId=%d: %+v", peopleID, err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return people, nil
}

func (p *PeopleStorage) GetPerson(ctx context.Context, peopleID uint64) (*models.People, error) {
	var people *models.People

	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		peopleInner, err := p.selectPeopleByID(ctx, tx, peopleID)
		if err != nil {
			return err
		}

		people = peopleInner

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return people, nil
}

func (p *PeopleStorage) deletePerson(ctx context.Context, tx pgx.Tx, personID uint64) error {
	SQLDeletePeople := `DELETE FROM public."people" WHERE id=$1`

	result, err := tx.Exec(ctx, SQLDeletePeople, personID)
	if err != nil {
		p.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedPeopleRows)
	}

	return nil
}

func (p *PeopleStorage) DeletePerson(ctx context.Context, personID uint64) error {
	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		err := p.deletePerson(ctx, tx, personID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		p.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}
