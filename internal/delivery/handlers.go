package delivery

import (
	"context"
	"github.com/OksidGen/enrich_server/internal/usecase"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

type Delivery struct {
	usecase usecase.Usecase
}

func NewDelivery(usecase usecase.Usecase) *Delivery {
	return &Delivery{usecase}
}

func (d *Delivery) RegisterRoutes(e *echo.Echo) {
	e.GET("/", d.Root)
	e.GET("/ping", d.Ping)
	e.GET("/people", d.GetPeople)
	e.GET("/people/:id", d.GetPerson)
	e.POST("/people", d.CreatePerson)
	e.PUT("/people/:id", d.UpdatePerson)
	e.DELETE("/people/:id", d.DeletePerson)
}

func (d *Delivery) Root(c echo.Context) error {
	log.Debug().Msg("Calling Root handler")
	return c.JSON(http.StatusOK, map[string]string{"message": "Hello. This is Enrich Server."})
}

func (d *Delivery) Ping(c echo.Context) error {
	log.Debug().Msg("Calling Ping handler")
	return c.JSON(http.StatusOK, map[string]string{"message": "pong"})
}

func (d *Delivery) GetPeople(c echo.Context) error {
	log.Debug().Msg("Calling GetPeople handler")

	queryParams := c.QueryParams()
	params := make(map[string]interface{})

	for key, value := range queryParams {
		params[key] = value[0]
	}

	people, err := d.usecase.GetPeople(context.Background(), params)
	if err != nil {
		log.Error().Err(err).Msg("Failed to call usecase.GetPeople")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, people)
}

func (d *Delivery) GetPerson(c echo.Context) error {
	log.Debug().Msg("Calling GetPerson handler")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to convert id to int")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	person, err := d.usecase.GetPersonByID(context.Background(), id)
	if err != nil {
		log.Err(err).Msg("Failed to call usecase.GetPersonByID")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, person)
}

func (d *Delivery) CreatePerson(c echo.Context) error {
	log.Debug().Msg("Calling CreatePerson handler")

	params := make(map[string]interface{})
	if err := c.Bind(&params); err != nil {
		log.Error().Err(err).Msg("Failed to bind params")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	id, err := d.usecase.CreatePerson(context.Background(), params)
	if err != nil {
		log.Err(err).Msg("Failed to call usecase.CreatePerson")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]int{"id": id})
}

func (d *Delivery) UpdatePerson(c echo.Context) error {
	log.Debug().Msg("Calling UpdatePerson handler")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Err(err).Msg("Failed to convert id to int")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	updates := make(map[string]interface{})
	if err := c.Bind(&updates); err != nil {
		log.Err(err).Msg("Failed to bind updates")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	delete(updates, "id")

	err = d.usecase.UpdatePerson(context.Background(), id, updates)
	if err != nil {
		log.Err(err).Msg("Failed to call usecase.UpdatePerson")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Person updated"})
}

func (d *Delivery) DeletePerson(c echo.Context) error {
	log.Debug().Msg("Calling DeletePerson handler")

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Err(err).Msg("Failed to convert id to int")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err = d.usecase.DeletePerson(context.Background(), id)
	if err != nil {
		log.Err(err).Msg("Failed to call usecase.DeletePerson")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Person deleted"})
}
