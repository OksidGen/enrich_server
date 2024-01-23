package entity

import "fmt"

type Person struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Surname     string `json:"surname"`
	Patronymic  string `json:"patronymic,omitempty"`
	Age         int    `json:"age,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Nationality string `json:"nationality,omitempty"`
}

func (person *Person) MapToPerson(data map[string]interface{}) error {

	for key, value := range data {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				person.Name = name
			} else {
				return fmt.Errorf("Ошибка в поле 'name'")
			}
		case "surname":
			if surname, ok := value.(string); ok {
				person.Surname = surname
			} else {
				return fmt.Errorf("Ошибка в поле 'surname'")
			}
		case "patronymic":
			if patronymic, ok := value.(string); ok {
				person.Patronymic = patronymic
			} else {
				return fmt.Errorf("Ошибка в поле 'patronymic'")
			}
		case "age":
			if age, ok := value.(int); ok {
				person.Age = age
			} else {
				return fmt.Errorf("Ошибка в поле 'age'")
			}
		case "gender":
			if gender, ok := value.(string); ok {
				person.Gender = gender
			} else {
				return fmt.Errorf("Ошибка в поле 'gender'")
			}
		case "nationality":
			if nationality, ok := value.(string); ok {
				person.Nationality = nationality
			} else {
				return fmt.Errorf("Ошибка в поле 'nationality'")
			}
		default:
			return fmt.Errorf("Недопустимое поле: %s", key)
		}
	}

	return nil
}
