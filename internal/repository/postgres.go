package repository

import (
	"context"
	"fmt"
	"github.com/OksidGen/enrich_server/internal/entity"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"strings"
)

type Repository interface {
	GetAllPeople(ctx context.Context) ([]entity.Person, error)
	GetPeopleWithFilters(ctx context.Context, filters map[string]interface{}, pagination map[string]int) ([]entity.Person, error)
	GetPersonByID(ctx context.Context, id int) (entity.Person, error)
	CreatePerson(ctx context.Context, person entity.Person) (int, error)
	UpdatePerson(ctx context.Context, id int, updates map[string]interface{}) error
	DeletePerson(ctx context.Context, id int) error
}

type postgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) Repository {
	return &postgresRepository{db}
}

func (r *postgresRepository) GetAllPeople(ctx context.Context) ([]entity.Person, error) {
	log.Debug().Msg("Calling GetAllPeople repository")

	var people []entity.Person
	err := r.db.SelectContext(ctx, &people, "SELECT id, name, surname, patronymic, age, gender, nationality FROM people")
	if err != nil {
		log.Err(err).Msg("Failed to get people")
		return nil, err
	}
	return people, nil
}

func (r *postgresRepository) GetPeopleWithFilters(ctx context.Context, filters map[string]interface{}, pagination map[string]int) ([]entity.Person, error) {
	log.Debug().Interface("filters", filters).Msg("Calling GetPeopleWithFilters repository")

	query := "SELECT * FROM people"
	var args []interface{}
	id := 1

	if len(filters) != 0 {
		query += " WHERE "
		for key, value := range filters {
			switch key {
			case "name", "surname", "patronymic", "gender", "nationality":
				query += fmt.Sprintf("%s ILIKE $%d AND ", key, id)
				args = append(args, fmt.Sprintf("%%%s%%", value))
			case "age", "minAge", "maxAge":
				switch key {
				case "age":
					query += fmt.Sprintf("age = $%d AND ", id)
				case "minAge":
					query += fmt.Sprintf("age >= $%d AND ", id)
				case "maxAge":
					query += fmt.Sprintf("age <= $%d AND ", id)
				}
				args = append(args, value)
			}
			id++
		}
		query = strings.TrimSuffix(query, " AND ")
	}

	if len(pagination) != 0 {
		page := pagination["page"]
		limit := pagination["limit"]
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", id, id+1)
		args = append(args, limit, (page-1)*limit)
	}

	var people []entity.Person
	err := r.db.SelectContext(ctx, &people, query, args...)
	if err != nil {
		return nil, err
	}

	return people, nil
}

func (r *postgresRepository) GetPersonByID(ctx context.Context, id int) (entity.Person, error) {
	log.Debug().Int("id", id).Msgf("Calling GetPersonByID repository")

	var person entity.Person
	err := r.db.GetContext(ctx, &person, "SELECT * FROM people WHERE id = $1", id)
	if err != nil {
		log.Err(err).Int("id", id).Msg("Failed to get person by ID")
		return entity.Person{}, err
	}
	return person, nil
}

func (r *postgresRepository) CreatePerson(ctx context.Context, person entity.Person) (int, error) {
	log.Debug().Interface("person", person).Msg("Calling CreatePerson repository")

	var id int
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO people (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality).Scan(&id)
	if err != nil {
		log.Err(err).Interface("person", person).Msg("Failed to create person")
		return 0, err
	}
	return id, nil
}

func (r *postgresRepository) UpdatePerson(ctx context.Context, id int, updates map[string]interface{}) error {
	log.Debug().Int("id", id).Interface("updates", updates).Msg("Calling UpdatePerson repository")

	updateQuery := "UPDATE people SET "
	var args []interface{}
	i := 1

	for field, value := range updates {
		updateQuery += fmt.Sprintf("%s = $%d, ", field, i)

		args = append(args, value)
		i++
	}

	updateQuery = updateQuery[:len(updateQuery)-2] + fmt.Sprintf(" WHERE id = $%d", i)

	args = append(args, id)

	_, err := r.db.ExecContext(ctx, updateQuery, args...)
	if err != nil {
		log.Err(err).Int("id", id).Interface("updates", updates).Msg("Failed to update person")
		return err
	}

	return nil
}

func (r *postgresRepository) DeletePerson(ctx context.Context, id int) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM people WHERE id = $1", id)
	if err != nil {
		log.Err(err).Int("id", id).Msg("Failed to delete person")
		return err
	}
	return nil
}
