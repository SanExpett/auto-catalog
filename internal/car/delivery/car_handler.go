package delivery

import (
	"context"
	"fmt"
	"github.com/SanExpett/auto-catalog/internal/car/usecases"
	"github.com/SanExpett/auto-catalog/internal/server/delivery"
	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/SanExpett/auto-catalog/pkg/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var _ ICarService = (*usecases.CarService)(nil)

type ICarService interface {
	AddCar(ctx context.Context, r io.Reader) (*models.Car, error)
	GetCar(ctx context.Context, carID uint64) (*models.Car, error)
	DeleteCar(ctx context.Context, carID uint64) error
	UpdateCar(ctx context.Context, r io.Reader, isPartialUpdate bool, carID uint64) error
	GetCarsList(ctx context.Context, limit uint64, offset uint64, model string, mark string, ownerID uint64,
		sortByYearType uint64) ([]*models.Car, error)
}

type CarHandler struct {
	service ICarService
	logger  *zap.SugaredLogger
}

func NewCarHandler(CarService ICarService) (*CarHandler, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &CarHandler{
		service: CarService,
		logger:  logger,
	}, nil
}

// AddCarHandler godoc
//
//	@Summary    add Car
//	@Description  add Car by data
//	@Description Error.status can be:
//	@Description StatusErrBadRequest      = 400
//	@Description  StatusErrInternalServer  = 500
//	@Tags Car
//
//	@Accept      json
//	@Produce    json
//	@Param      Car  body models.PreCar true  "Car data for adding"
//	@Success    200  {object} CarResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /car/add [post]
func (p *CarHandler) AddCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	person, err := p.service.AddCar(ctx, r.Body)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger, NewCarResponse(delivery.StatusResponseSuccessful, person))
	p.logger.Infof("in AddCarHandler: add Car: %+v", person)
}

// GetCarHandler godoc
//
//	@Summary    get Car
//	@Description  get Car by id
//	@Tags Car
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Car id"
//	@Success    200  {object} CarResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /car/get [get]
func (p *CarHandler) GetCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	CarID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	Car, err := p.service.GetCar(ctx, CarID)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger, NewCarResponse(delivery.StatusResponseSuccessful, Car))
	p.logger.Infof("in GetCarHandler: get Car: %+v", Car)
}

// DeleteCarHandler godoc
//
//	@Summary     delete Car
//	@Description  delete Car for author using user id from cookies\jwt.
//	@Description  This totally removed Car. Recovery will be impossible
//	@Tags Car
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "Car id"
//	@Success    200  {object} delivery.Response
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /car/delete [delete]
func (c *CarHandler) DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	carID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, c.logger, err)

		return
	}

	err = c.service.DeleteCar(ctx, carID)
	if err != nil {
		delivery.HandleErr(w, c.logger, err)

		return
	}

	delivery.SendOkResponse(w, c.logger,
		delivery.NewResponse(delivery.StatusResponseSuccessful, ResponseSuccessfulDeleteCar))
	c.logger.Infof("in DeleteCarHandler: delete Car id=%d", carID)
}

// UpdateCarHandler godoc
//
//	@Summary    update Car
//	@Description  update Car by id
//	@Tags Car
//	@Accept      json
//	@Produce    json
//	@Param      id query uint64 true  "Car id"
//	@Param      preCar  body models.PreCar false  "полностью опционален"
//	@Success    200  {object} delivery.ResponseID
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /car/update [patch]
//	@Router      /car/update [put]
func (c *CarHandler) UpdateCarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	carID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, c.logger, err)

		return
	}

	ctx := r.Context()

	if r.Method == http.MethodPatch {
		err = c.service.UpdateCar(ctx, r.Body, true, carID)
	} else {
		err = c.service.UpdateCar(ctx, r.Body, false, carID)
	}

	if err != nil {
		delivery.HandleErr(w, c.logger, err)

		return
	}

	delivery.SendOkResponse(w, c.logger, delivery.NewResponseID(carID))
	c.logger.Infof("in UpdateCarHandler: updated Car with id = %+v", carID)
}

// GetCarsListHandler godoc
//
//	@Summary    get Cars list
//	@Description  get Cars by count and last_id return old Cars
//	@Tags Car
//	@Accept      json
//	@Produce    json
//	@Param      limit  query uint64 true  "limit Cars"
//	@Param      offset  query uint64 true  "offset of Cars"
//	@Param      mark  query string true  "mark of cars in list"
//	@Param      model  query string true  "model of cars in list"
//	@Param      sort_by_year_type query uint64 true  "type of sort(0 - by year desc, 1 - by year asc)"
//	@Success    200  {object} CarListResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /car/get_list [get]
func (c *CarHandler) GetCarsListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	limit, err := utils.ParseUint64FromRequest(r, "limit")
	if err != nil {
		limit = 10
	}

	offset, err := utils.ParseUint64FromRequest(r, "offset")
	if err != nil {
		offset = 0
	}

	ownerID, err := utils.ParseUint64FromRequest(r, "owner_id")
	if err != nil {
		ownerID = 0
	}

	sortByYearType, err := utils.ParseUint64FromRequest(r, "sort_by_year_type")
	if err != nil {
		sortByYearType = 0
	}

	model := utils.ParseStringFromRequest(r, "model")
	mark := utils.ParseStringFromRequest(r, "mark")

	cars, err := c.service.GetCarsList(ctx, limit, offset, model, mark, ownerID, sortByYearType)
	if err != nil {
		delivery.HandleErr(w, c.logger, err)

		return
	}

	delivery.SendOkResponse(w, c.logger, NewCarListResponse(delivery.StatusResponseSuccessful, cars))
	c.logger.Infof("in GetCarListHandler: get Car list: %+v", cars)
}
