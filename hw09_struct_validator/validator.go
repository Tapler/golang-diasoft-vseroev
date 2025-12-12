package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation errors: ")
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(fmt.Sprintf("%s: %v", err.Field, err.Err))
	}
	return sb.String()
}

// Sentinel errors для валидации.
var (
	ErrInvalidLength  = errors.New("invalid length")
	ErrNotInSet       = errors.New("value not in set")
	ErrRegexpMismatch = errors.New("regexp mismatch")
	ErrLessThanMin    = errors.New("value less than min")
	ErrGreaterThanMax = errors.New("value greater than max")
)

// Программные ошибки.
var (
	ErrNotStruct  = errors.New("input is not a struct")
	ErrInvalidTag = errors.New("invalid validation tag")
)

// Validate валидирует публичные поля структуры на основе тега validate.
func Validate(v interface{}) error {
	if v == nil {
		return ErrNotStruct
	}

	val := reflect.ValueOf(v)
	typ := val.Type()

	// Разыменовываем указатель если нужно
	if typ.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = val.Type()
	}

	// Проверяем что это структура
	if typ.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	var validationErrs ValidationErrors

	// Проходим по всем полям структуры
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Пропускаем неэкспортируемые поля
		if !field.IsExported() {
			continue
		}

		// Получаем тег validate
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Валидируем поле
		errs, err := validateField(field.Name, fieldValue, tag)
		if err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
		validationErrs = append(validationErrs, errs...)
	}

	if len(validationErrs) > 0 {
		return validationErrs
	}

	return nil
}

// validateField валидирует конкретное поле.
func validateField(fieldName string, fieldValue reflect.Value, tag string) (ValidationErrors, error) {
	var validationErrs ValidationErrors

	// Разбиваем тег на отдельные правила валидации (разделитель |)
	rules := strings.Split(tag, "|")

	//nolint:exhaustive // Нужно только строки, числа, слайсы
	switch fieldValue.Kind() {
	case reflect.String:
		return validateStringField(fieldName, fieldValue.String(), rules)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateIntField(fieldName, fieldValue.Int(), rules)

	case reflect.Slice:
		// Валидируем каждый элемент слайса
		for j := 0; j < fieldValue.Len(); j++ {
			elem := fieldValue.Index(j)
			// Добавляем индекс элемента к имени поля для точной идентификации ошибок
			elemFieldName := fmt.Sprintf("%s[%d]", fieldName, j)
			elemErrs, err := validateField(elemFieldName, elem, tag)
			if err != nil {
				return nil, err
			}
			validationErrs = append(validationErrs, elemErrs...)
		}
		return validationErrs, nil

	default:
		return nil, nil
	}
}

// validateString валидирует строковое значение согласно правилу.
func validateString(value, rule string) error {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: %s", ErrInvalidTag, rule)
	}

	validator := parts[0]
	param := parts[1]

	switch validator {
	case "len":
		expectedLen, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("%w: invalid len parameter: %s", ErrInvalidTag, param)
		}
		if len(value) != expectedLen {
			return fmt.Errorf("%w: expected %d, got %d", ErrInvalidLength, expectedLen, len(value))
		}

	case "regexp":
		re, err := regexp.Compile(param)
		if err != nil {
			return fmt.Errorf("%w: invalid regexp: %s", ErrInvalidTag, param)
		}
		if !re.MatchString(value) {
			return fmt.Errorf("%w: value '%s' does not match pattern '%s'", ErrRegexpMismatch, value, param)
		}

	case "in":
		allowedValues := strings.Split(param, ",")
		found := false
		for _, allowed := range allowedValues {
			if value == allowed {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: value '%s' not in allowed set %v", ErrNotInSet, value, allowedValues)
		}

	default:
		return fmt.Errorf("%w: unknown string validator: %s", ErrInvalidTag, validator)
	}

	return nil
}

// validateInt валидирует int значение согласно правилу.
func validateInt(value int64, rule string) error {
	parts := strings.SplitN(rule, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("%w: %s", ErrInvalidTag, rule)
	}

	validator := parts[0]
	param := parts[1]

	switch validator {
	case "min":
		minVal, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: invalid min parameter: %s", ErrInvalidTag, param)
		}
		if value < minVal {
			return fmt.Errorf("%w: value %d is less than %d", ErrLessThanMin, value, minVal)
		}

	case "max":
		maxVal, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: invalid max parameter: %s", ErrInvalidTag, param)
		}
		if value > maxVal {
			return fmt.Errorf("%w: value %d is greater than %d", ErrGreaterThanMax, value, maxVal)
		}

	case "in":
		allowedValues := strings.Split(param, ",")
		found := false
		for _, allowed := range allowedValues {
			allowedInt, err := strconv.ParseInt(allowed, 10, 64)
			if err != nil {
				return fmt.Errorf("%w: invalid in parameter: %s", ErrInvalidTag, allowed)
			}
			if value == allowedInt {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: value %d not in allowed set %s", ErrNotInSet, value, param)
		}

	default:
		return fmt.Errorf("%w: unknown int validator: %s", ErrInvalidTag, validator)
	}

	return nil
}

// validateStringField валидирует строковое поле по набору правил.
func validateStringField(fieldName, value string, rules []string) (ValidationErrors, error) {
	return validateFieldWithRules(fieldName, rules, func(rule string) error {
		return validateString(value, rule)
	})
}

// validateIntField валидирует int поле по набору правил.
func validateIntField(fieldName string, value int64, rules []string) (ValidationErrors, error) {
	return validateFieldWithRules(fieldName, rules, func(rule string) error {
		return validateInt(value, rule)
	})
}

// validateFieldWithRules общая функция для валидации поля по набору правил.
func validateFieldWithRules(
	fieldName string,
	rules []string,
	validateFunc func(string) error,
) (ValidationErrors, error) {
	var validationErrs ValidationErrors

	for _, rule := range rules {
		if err := validateFunc(rule); err != nil {
			// Программные ошибки возвращаем сразу
			if errors.Is(err, ErrInvalidTag) {
				return nil, err
			}
			// Ошибки валидации накапливаем
			validationErrs = append(validationErrs, ValidationError{
				Field: fieldName,
				Err:   err,
			})
		}
	}

	return validationErrs, nil
}
