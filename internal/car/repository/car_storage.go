package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/SanExpett/auto-catalog/internal/server/repository"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"strconv"
	"time"
)

var (
	ErrCarNotFound       = myerrors.NewError("Эта машина не найдена")
	ErrNoAffectedCarRows = myerrors.NewError("Не получилось обновить данные автомобиля")
	ErrNoUpdateFields    = myerrors.NewError("Вы пытаетесь обновить пустое количество полей автомобиля")
	ErrNoPerson          = myerrors.NewError("Вы пытаетесь добавить машину для несуществующего человека")

	NameSeqCar = pgx.Identifier{"public", "car_id_seq"} //nolint:gochecknoglobals
)

const (
	byYearDESC = 0
	byYearASC  = 1
)

type CarStorage struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewCarStorage(pool *pgxpool.Pool) (*CarStorage, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &CarStorage{
		pool:   pool,
		logger: logger,
	}, nil
}

func (p *CarStorage) selectCreatedAtByCarID(ctx context.Context, tx pgx.Tx, carID uint64,
) (time.Time, error) {
	SQLSelectCreatedAtByCarID := `SELECT created_at FROM public."car" WHERE id=$1`

	var createdAt time.Time

	createdAtRow := tx.QueryRow(ctx, SQLSelectCreatedAtByCarID, carID)
	if err := createdAtRow.Scan(&createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf(myerrors.ErrTemplate, ErrCarNotFound)
		}

		p.logger.Errorf("error with CarId=%d: %+v", carID, err)

		return time.Time{}, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return createdAt, nil
}

func (p *CarStorage) checkPersonExistByID(ctx context.Context, tx pgx.Tx, personID uint64) (bool, error) {
	SQLCheckPersonExistByID := `SELECT EXISTS (SELECT 1 FROM public."people" WHERE id=$1);`

	var exist bool

	existRow := tx.QueryRow(ctx, SQLCheckPersonExistByID, personID)
	if err := existRow.Scan(&exist); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf(myerrors.ErrTemplate, ErrCarNotFound)
		}

		p.logger.Errorf("error with personID=%d: %+v", personID, err)

		return false, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return exist, nil
}

func (p *CarStorage) insertCar(ctx context.Context, tx pgx.Tx, preCar *models.PreCar) error {
	SQLInsertCar := `INSERT INTO public."car"(owner_id, reg_num, mark, model, year) VALUES($1, $2, $3, $4, $5)`
	_, err := tx.Exec(ctx, SQLInsertCar, preCar.OwnerID, preCar.RegNum, preCar.Mark, preCar.Model, preCar.Year)

	if err != nil {
		p.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (p *CarStorage) AddCar(ctx context.Context, preCar *models.PreCar) (*models.Car, error) {
	car := &models.Car{OwnerID: preCar.OwnerID, RegNum: preCar.RegNum, Mark: preCar.Mark,
		Model: preCar.Model, Year: preCar.Year}

	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		exist, err := p.checkPersonExistByID(ctx, tx, car.OwnerID)
		if err != nil {
			return err
		}
		if !exist {
			return fmt.Errorf(myerrors.ErrTemplate, ErrNoPerson)
		}

		err = p.insertCar(ctx, tx, preCar)
		if err != nil {
			return err
		}

		lastCarID, err := repository.GetLastValSeq(ctx, tx, NameSeqCar)
		if err != nil {
			return err
		}

		car.ID = lastCarID

		createdAt, err := p.selectCreatedAtByCarID(ctx, tx, lastCarID)
		if err != nil {
			return err
		}

		car.CreatedAt = createdAt

		return err
	})
	if err != nil {
		p.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return car, nil
}

func (p *CarStorage) selectCarByID(ctx context.Context, tx pgx.Tx, carID uint64,
) (*models.Car, error) {
	SQLSelectCar := `SELECT owner_id, reg_num, mark, model, year, created_at FROM public."car" WHERE id=$1`
	car := &models.Car{ID: carID} //nolint:exhaustruct

	carRow := tx.QueryRow(ctx, SQLSelectCar, carID)
	if err := carRow.Scan(&car.OwnerID, &car.RegNum, &car.Mark, &car.Model, &car.Year, &car.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(myerrors.ErrTemplate, ErrCarNotFound)
		}

		p.logger.Errorf("error with CarId=%d: %+v", carID, err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return car, nil
}

func (p *CarStorage) GetCar(ctx context.Context, CarID uint64) (*models.Car, error) {
	var car *models.Car

	err := pgx.BeginFunc(ctx, p.pool, func(tx pgx.Tx) error {
		carInner, err := p.selectCarByID(ctx, tx, CarID)
		if err != nil {
			return err
		}

		car = carInner

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return car, nil
}

func (c *CarStorage) deleteCar(ctx context.Context, tx pgx.Tx, carID uint64) error {
	SQLDeleteCar := `DELETE FROM public."car" WHERE id=$1`

	result, err := tx.Exec(ctx, SQLDeleteCar, carID)
	if err != nil {
		c.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedCarRows)
	}

	return nil
}

func (c *CarStorage) DeleteCar(ctx context.Context, carID uint64) error {
	err := pgx.BeginFunc(ctx, c.pool, func(tx pgx.Tx) error {
		err := c.deleteCar(ctx, tx, carID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (c *CarStorage) updateCar(ctx context.Context, tx pgx.Tx,
	carID uint64, updateFields map[string]interface{},
) error {
	if len(updateFields) == 0 {
		return ErrNoUpdateFields
	}

	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Update(`public."car"`).
		Where(squirrel.Eq{"id": carID}).SetMap(updateFields)

	queryString, args, err := query.ToSql()
	if err != nil {
		c.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	result, err := tx.Exec(ctx, queryString, args...)
	if err != nil {
		c.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf(myerrors.ErrTemplate, ErrNoAffectedCarRows)
	}

	return nil
}

func (c *CarStorage) UpdateCar(ctx context.Context, carID uint64, updateFields map[string]interface{}) error {
	err := pgx.BeginFunc(ctx, c.pool, func(tx pgx.Tx) error {
		err := c.updateCar(ctx, tx, carID, updateFields)

		return err
	})
	if err != nil {
		c.logger.Errorln(err)

		return fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return nil
}

func (c *CarStorage) selectCarsWithWhereOrderLimitOffset(ctx context.Context, tx pgx.Tx,
	limit uint64, offset uint64, whereClause any, orderByClause []string,
) ([]*models.Car, error) {
	query := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Select("id, owner_id, reg_num, " +
		"mark, model, year, created_at").From(`public."car"`).
		Where(whereClause).OrderBy(orderByClause...).Limit(limit).Offset(offset)

	SQLQuery, args, err := query.ToSql()
	if err != nil {
		c.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	rowsCars, err := tx.Query(ctx, SQLQuery, args...)
	if err != nil {
		c.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	curCar := new(models.Car)

	var slCar []*models.Car

	_, err = pgx.ForEachRow(rowsCars, []any{
		&curCar.ID, &curCar.OwnerID, &curCar.RegNum, &curCar.Mark,
		&curCar.Model, &curCar.Year, &curCar.CreatedAt,
	}, func() error {
		slCar = append(slCar, &models.Car{ //nolint:exhaustruct
			ID:        curCar.ID,
			OwnerID:   curCar.OwnerID,
			RegNum:    curCar.RegNum,
			Mark:      curCar.Mark,
			Model:     curCar.Model,
			Year:      curCar.Year,
			CreatedAt: curCar.CreatedAt,
		})

		return nil
	})
	if err != nil {
		c.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slCar, nil
}

func (c *CarStorage) GetCarsList(ctx context.Context, limit uint64, offset uint64,
	model string, mark string, ownerID uint64, sortByYearType uint64,
) ([]*models.Car, error) {
	var slCar []*models.Car

	var orderByClause []string

	switch sortByYearType {
	case byYearDESC:
		orderByClause = []string{"year DESC"}
	case byYearASC:
		orderByClause = []string{"year ASC"}
	}

	whereClause := ""
	if mark != "" {
		whereClause += "mark = " + mark + " AND "
	}
	if model != "" {
		whereClause += "model = " + model + " AND "
	}
	if ownerID != 0 {
		whereClause += "owner_id = " + strconv.FormatUint(ownerID, 10) + " AND "
	}

	if whereClause != "" {
		whereClause = whereClause[:len(whereClause)-5]
	}

	err := pgx.BeginFunc(ctx, c.pool, func(tx pgx.Tx) error {
		var err error
		slCar, err = c.selectCarsWithWhereOrderLimitOffset(ctx,
			tx, limit, offset, whereClause, orderByClause)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		c.logger.Errorln(err)

		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return slCar, nil
}
