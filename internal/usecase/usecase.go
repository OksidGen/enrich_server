package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/OksidGen/enrich_server/internal/entity"
	"github.com/OksidGen/enrich_server/internal/repository"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
)

type Usecase interface {
	GetPeople(ctx context.Context, params map[string]interface{}) ([]entity.Person, error)
	GetPersonByID(ctx context.Context, id int) (entity.Person, error)
	CreatePerson(ctx context.Context, params map[string]interface{}) (int, error)
	UpdatePerson(ctx context.Context, id int, updates map[string]interface{}) error
	DeletePerson(ctx context.Context, id int) error
}

type usecase struct {
	repo repository.Repository
}

func NewUsecase(repo repository.Repository) Usecase {
	return &usecase{repo}
}

func (uc *usecase) GetPeople(ctx context.Context, params map[string]interface{}) ([]entity.Person, error) {
	log.Debug().Msg("Calling GetPeople usecase")

	hasQueryParams := len(params) != 0

	if !hasQueryParams {
		return uc.repo.GetAllPeople(ctx)
	}

	filters := make(map[string]interface{})
	pagination := make(map[string]int)

	if pageStr, ok := params["page"]; ok {
		page, err := strconv.Atoi(pageStr.(string))
		if err != nil {
			log.Error().Err(err).Str("param", "page").Str("value", pageStr.(string)).Msg("Failed to convert page to int")
			return nil, err
		}
		if limitStr, ok := params["limit"]; ok {
			limit, err := strconv.Atoi(limitStr.(string))
			if err != nil {
				log.Error().Err(err).Str("param", "limit").Str("value", limitStr.(string)).Msg("Failed to convert limit to int")
				return nil, err
			}
			pagination["page"] = page
			pagination["limit"] = limit
		} else {
			pagination["page"] = page
			pagination["limit"] = 10
		}

	} else {
		if limitStr, ok := params["limit"]; ok {
			limit, err := strconv.Atoi(limitStr.(string))
			if err != nil {
				log.Error().Err(err).Str("param", "limit").Str("value", limitStr.(string)).Msg("Failed to convert limit to int")
				return nil, err
			}
			pagination["page"] = 1
			pagination["limit"] = limit
		}
	}

	delete(params, "page")
	delete(params, "limit")

	for param, value := range params {
		switch param {
		case "age", "minAge", "maxAge":
			numValue, err := strconv.Atoi(value.(string))
			if err != nil {
				log.Error().Err(err).Str("param", param).Str("value", value.(string)).Msg("Failed to convert value of param to int")
				return nil, err
			}
			filters[param] = numValue
		case "name", "surname", "patronymic", "gender", "nationality":
			filters[param] = value
		default:
			err := fmt.Errorf("invalid query param: %s", param)
			log.Err(err).Str("param", param).Msg("Invalid query param")
			return nil, err
		}
	}
	if _, hasAge := filters["age"]; hasAge {
		delete(filters, "minAge")
		delete(filters, "maxAge")
	} else {
		if _, hasMinAge := filters["minAge"]; hasMinAge {
			if _, hasMaxAge := filters["maxAge"]; !hasMaxAge {
				filters["maxAge"] = 777
			}
		} else if _, hasMaxAge := filters["maxAge"]; hasMaxAge {
			filters["minAge"] = 0
		}
	}

	return uc.repo.GetPeopleWithFilters(ctx, params, pagination)
}

func (uc *usecase) GetPersonByID(ctx context.Context, id int) (entity.Person, error) {
	log.Debug().Int("id", id).Msgf("Calling GetPersonByID usecase")
	return uc.repo.GetPersonByID(ctx, id)
}

func (uc *usecase) CreatePerson(ctx context.Context, params map[string]interface{}) (int, error) {
	log.Debug().Interface("params", params).Msg("Calling CreatePerson usecase")

	if err := validateFields(params); err != nil {
		log.Err(err).Msg("Failed to validate fields")
		return 0, err
	}
	var person entity.Person
	if err := person.MapToPerson(params); err != nil {
		log.Err(err).Msg("Failed to map person")
		return 0, err
	}
	enrichPersonData(&person)

	return uc.repo.CreatePerson(ctx, person)
}

func (uc *usecase) UpdatePerson(ctx context.Context, id int, updates map[string]interface{}) error {
	log.Debug().Int("id", id).Interface("updates", updates).Msg("Calling UpdatePerson usecase")

	if err := validateFields(updates); err != nil {
		log.Err(err).Msg("Failed to validate fields")
		return err
	}

	err := uc.repo.UpdatePerson(ctx, id, updates)
	if err != nil {
		log.Err(err).Msg("Failed to update person")
		return err
	}

	return nil
}
func (uc *usecase) DeletePerson(ctx context.Context, id int) error {
	log.Debug().Int("id", id).Msg("Calling DeletePerson usecase")
	return uc.repo.DeletePerson(ctx, id)
}

const (
	agifyAPI       = "https://api.agify.io/?name=%s"
	genderizeAPI   = "https://api.genderize.io/?name=%s"
	nationalizeAPI = "https://api.nationalize.io/?name=%s"
)

func enrichPersonData(person *entity.Person) {
	log.Debug().Str("name", person.Name).Msg("Enriching person data")
	if person.Name != "" {
		person.Age = getAge(person.Name)
		person.Gender = getGender(person.Name)
		person.Nationality = getNationality(person.Name)
	}
}

func getAge(name string) int {
	log.Debug().Str("name", name).Msg("Getting age")

	url := fmt.Sprintf(agifyAPI, name)
	resp, err := http.Get(url)
	if err != nil {
		log.Err(err).Msg("Failed to get age")
		return 0
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("Failed to read response body")
		return 0
	}

	var ageResponse map[string]interface{}
	err = json.Unmarshal(body, &ageResponse)
	if err != nil {
		log.Err(err).Msg("Failed to unmarshal age response")
		return 0
	}

	age, ok := ageResponse["age"].(float64)
	if !ok {
		log.Err(err).Interface("ageResponse", ageResponse).Msg("Failed to get age from response")
		return 0
	}

	return int(age)
}

func getGender(name string) string {
	log.Debug().Str("name", name).Msg("Getting gender")

	url := fmt.Sprintf(genderizeAPI, name)
	resp, err := http.Get(url)
	if err != nil {
		log.Err(err).Msg("Failed to get gender")
		return ""
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("Failed to read response body")
		return ""
	}

	var genderResponse map[string]interface{}
	err = json.Unmarshal(body, &genderResponse)
	if err != nil {
		log.Err(err).Msg("Failed to unmarshal gender response")
		return ""
	}

	gender, ok := genderResponse["gender"].(string)
	if !ok {
		log.Err(err).Interface("genderResponse", genderResponse).Msg("Failed to get gender from response")
		return ""
	}

	return gender
}

func getNationality(name string) string {
	log.Debug().Str("name", name).Msg("Getting nationality")

	url := fmt.Sprintf(nationalizeAPI, name)
	resp, err := http.Get(url)
	if err != nil {
		log.Err(err).Msg("Failed to get nationality")
		return ""
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Err(err).Msg("Failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Err(err).Msg("Failed to read response body")
		return ""
	}

	var nationalityResponse map[string]interface{}
	err = json.Unmarshal(body, &nationalityResponse)
	if err != nil {
		log.Err(err).Msg("Failed to unmarshal nationality response")
		return ""
	}

	countryArray, ok := nationalityResponse["country"].([]interface{})
	if !ok || len(countryArray) == 0 {
		log.Err(err).Interface("nationalityResponse", nationalityResponse).Msg("Failed to get nationality from response")
		return ""
	}

	countryInfo := countryArray[0].(map[string]interface{})
	countryCode, ok := countryInfo["country_id"].(string)
	if !ok {
		log.Err(err).Interface("countryInfo", countryInfo).Msg("Failed to get country code from response")
		return ""
	}

	return countryCode
}

func validateFields(data map[string]interface{}) error {
	log.Debug().Interface("data", data).Msg("Validating fields")

	allowedFields := map[string]bool{
		"name":        true,
		"surname":     true,
		"patronymic":  true,
		"age":         true,
		"gender":      true,
		"nationality": true,
	}

	for field := range data {
		if _, ok := allowedFields[field]; !ok {
			return fmt.Errorf("invalid field: %s", field)
		}
	}

	return nil
}
