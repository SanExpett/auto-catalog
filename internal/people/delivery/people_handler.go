package delivery

import (
	"context"
	"fmt"
	"github.com/SanExpett/auto-catalog/internal/people/usecases"
	"github.com/SanExpett/auto-catalog/internal/server/delivery"

	"github.com/SanExpett/auto-catalog/pkg/models"
	myerrors "github.com/SanExpett/auto-catalog/pkg/my_errors"
	"github.com/SanExpett/auto-catalog/pkg/my_logger"
	"github.com/SanExpett/auto-catalog/pkg/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
)

var _ IPeopleService = (*usecases.PeopleService)(nil)

type IPeopleService interface {
	AddPerson(ctx context.Context, r io.Reader) (*models.People, error)
	GetPerson(ctx context.Context, personID uint64) (*models.People, error)
	DeletePerson(ctx context.Context, personID uint64) error
}

type PeopleHandler struct {
	service IPeopleService
	logger  *zap.SugaredLogger
}

func NewPeopleHandler(PeopleService IPeopleService) (*PeopleHandler, error) {
	logger, err := my_logger.Get()
	if err != nil {
		return nil, fmt.Errorf(myerrors.ErrTemplate, err)
	}

	return &PeopleHandler{
		service: PeopleService,
		logger:  logger,
	}, nil
}

// AddPeopleHandler godoc
//
//	@Summary    add people
//	@Description  add People by data
//	@Description Error.status can be:
//	@Description StatusErrBadRequest      = 400
//	@Description  StatusErrInternalServer  = 500
//	@Tags People
//
//	@Accept      json
//	@Produce    json
//	@Param      People  body models.PrePeople true  "People data for adding"
//	@Success    200  {object} PeopleResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /people/add [post]
func (p *PeopleHandler) AddPeopleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	person, err := p.service.AddPerson(ctx, r.Body)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger, NewPeopleResponse(delivery.StatusResponseSuccessful, person))
	p.logger.Infof("in AddPeopleHandler: add people: %+v", person)
}

// GetPeopleHandler godoc
//
//	@Summary    get People
//	@Description  get People by id
//	@Tags People
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "People id"
//	@Success    200  {object} PeopleResponse
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /people/get [get]
func (p *PeopleHandler) GetPeopleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	peopleID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	people, err := p.service.GetPerson(ctx, peopleID)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger, NewPeopleResponse(delivery.StatusResponseSuccessful, people))
	p.logger.Infof("in GetPeopleHandler: get People: %+v", people)
}

// DeletePeopleHandler godoc
//
//	@Summary     delete People
//	@Description  delete People for author using user id from cookies\jwt.
//	@Description  This totally removed People. Recovery will be impossible
//	@Tags People
//	@Accept      json
//	@Produce    json
//	@Param      id  query uint64 true  "People id"
//	@Success    200  {object} delivery.Response
//	@Failure    405  {string} string
//	@Failure    500  {string} string
//	@Failure    222  {object} delivery.ErrorResponse "Error"
//	@Router      /people/delete [delete]
func (p *PeopleHandler) DeletePeopleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `Method not allowed`, http.StatusMethodNotAllowed)

		return
	}

	ctx := r.Context()

	personID, err := utils.ParseUint64FromRequest(r, "id")
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	err = p.service.DeletePerson(ctx, personID)
	if err != nil {
		delivery.HandleErr(w, p.logger, err)

		return
	}

	delivery.SendOkResponse(w, p.logger,
		delivery.NewResponse(delivery.StatusResponseSuccessful, ResponseSuccessfulDeletePeople))
	p.logger.Infof("in DeletePeopleHandler: delete People id=%d", personID)
}
