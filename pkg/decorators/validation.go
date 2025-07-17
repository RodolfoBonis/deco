package decorators

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Validator global validator instance
var validate *validator.Validate

// init initializes the validator
func init() {
	validate = validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("phone", validatePhone); err != nil {
		log.Printf("Failed to register phone validation: %v", err)
	}
	if err := validate.RegisterValidation("cpf", validateCPF); err != nil {
		log.Printf("Failed to register CPF validation: %v", err)
	}
	if err := validate.RegisterValidation("cnpj", validateCNPJ); err != nil {
		log.Printf("Failed to register CNPJ validation: %v", err)
	}
	if err := validate.RegisterValidation("datetime", validateDateTime); err != nil {
		log.Printf("Failed to register datetime validation: %v", err)
	}
}

// ValidationResponse validation error response
type ValidationResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Fields  []ValidationField      `json:"fields,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ValidationField field-specific error
type ValidationField struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}

// ValidateStruct middleware for automatic struct validation
func ValidateStruct(config *ValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Continue processing
		c.Next()

		// If there are no binding errors, we do not need to validate
		if len(c.Errors) == 0 {
			return
		}

		// Process validation errors
		var validationErrors []ValidationField

		for _, err := range c.Errors {
			if validatorErr, ok := err.Err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validatorErr {
					validationErrors = append(validationErrors, ValidationField{
						Field:   fieldErr.Field(),
						Value:   fmt.Sprintf("%v", fieldErr.Value()),
						Tag:     fieldErr.Tag(),
						Message: getValidationMessage(fieldErr, config),
						Param:   fieldErr.Param(),
					})
				}
			}
		}

		if len(validationErrors) > 0 {
			response := ValidationResponse{
				Error:   "validation_failed",
				Message: "Invalid data provided",
				Fields:  validationErrors,
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
	}
}

// ValidateJSON middleware for automatic JSON validation
func ValidateJSON(target interface{}, config *ValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new instance of the target type
		targetType := reflect.TypeOf(target)
		if targetType.Kind() == reflect.Ptr {
			targetType = targetType.Elem()
		}

		newInstance := reflect.New(targetType).Interface()

		// Bind the JSON
		if err := c.ShouldBindJSON(newInstance); err != nil {
			var validationErrors []ValidationField

			// If it is a validation error, process fields
			if validatorErr, ok := err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validatorErr {
					validationErrors = append(validationErrors, ValidationField{
						Field:   fieldErr.Field(),
						Value:   fmt.Sprintf("%v", fieldErr.Value()),
						Tag:     fieldErr.Tag(),
						Message: getValidationMessage(fieldErr, config),
						Param:   fieldErr.Param(),
					})
				}
			} else {
				// JSON parsing error
				validationErrors = append(validationErrors, ValidationField{
					Field:   "json",
					Message: "Formato JSON inválido",
				})
			}

			response := ValidationResponse{
				Error:   "validation_failed",
				Message: "Invalid data provided",
				Fields:  validationErrors,
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		// Validate the instance
		if err := validate.Struct(newInstance); err != nil {
			var validationErrors []ValidationField

			if validatorErr, ok := err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validatorErr {
					validationErrors = append(validationErrors, ValidationField{
						Field:   fieldErr.Field(),
						Value:   fmt.Sprintf("%v", fieldErr.Value()),
						Tag:     fieldErr.Tag(),
						Message: getValidationMessage(fieldErr, config),
						Param:   fieldErr.Param(),
					})
				}
			}

			if len(validationErrors) > 0 {
				response := ValidationResponse{
					Error:   "validation_failed",
					Message: "Invalid data provided",
					Fields:  validationErrors,
				}

				c.AbortWithStatusJSON(http.StatusBadRequest, response)
				return
			}
		}

		// Save in context for later use
		c.Set("validated_data", newInstance)
		c.Next()
	}
}

// ValidateQuery middleware for query parameter validation
func ValidateQuery(target interface{}, config *ValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetType := reflect.TypeOf(target)
		if targetType.Kind() == reflect.Ptr {
			targetType = targetType.Elem()
		}

		newInstance := reflect.New(targetType).Interface()

		// Bind the query parameters
		if err := c.ShouldBindQuery(newInstance); err != nil {
			var validationErrors []ValidationField

			if validatorErr, ok := err.(validator.ValidationErrors); ok {
				for _, fieldErr := range validatorErr {
					validationErrors = append(validationErrors, ValidationField{
						Field:   fieldErr.Field(),
						Value:   fmt.Sprintf("%v", fieldErr.Value()),
						Tag:     fieldErr.Tag(),
						Message: getValidationMessage(fieldErr, config),
						Param:   fieldErr.Param(),
					})
				}
			}

			response := ValidationResponse{
				Error:   "validation_failed",
				Message: "Invalid query parameters",
				Fields:  validationErrors,
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		// Save no contexto
		c.Set("validated_query", newInstance)
		c.Next()
	}
}

// ValidateParams middleware for path parameter validation
func ValidateParams(rules map[string]string, config *ValidationConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var validationErrors []ValidationField

		for param, rule := range rules {
			value := c.Param(param)
			if value == "" {
				validationErrors = append(validationErrors, ValidationField{
					Field:   param,
					Value:   "",
					Tag:     "required",
					Message: fmt.Sprintf("Parameter '%s' is required", param),
				})
				continue
			}

			// Validate based on the rule
			if !validateParamValue(value, rule) {
				validationErrors = append(validationErrors, ValidationField{
					Field:   param,
					Value:   value,
					Tag:     rule,
					Message: fmt.Sprintf("Parameter '%s' does not meet rule '%s'", param, rule),
				})
			}
		}

		if len(validationErrors) > 0 {
			response := ValidationResponse{
				Error:   "validation_failed",
				Message: "Invalid route parameters",
				Fields:  validationErrors,
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}

		c.Next()
	}
}

// getValidationMessage returns custom error message
func getValidationMessage(fieldErr validator.FieldError, config *ValidationConfig) string {
	field := fieldErr.Field()
	tag := fieldErr.Tag()
	param := fieldErr.Param()

	// Custom messages in Portuguese
	messages := map[string]string{
		"required": "is required",
		"email":    "must be a valid email",
		"min":      fmt.Sprintf("must have at least %s characters", param),
		"max":      fmt.Sprintf("must have at most %s characters", param),
		"len":      fmt.Sprintf("must have exactly %s characters", param),
		"gt":       fmt.Sprintf("must be greater than %s", param),
		"gte":      fmt.Sprintf("must be greater than or equal to %s", param),
		"lt":       fmt.Sprintf("must be less than %s", param),
		"lte":      fmt.Sprintf("must be less than or equal to %s", param),
		"alpha":    "must contain only letters",
		"alphanum": "must contain only letters and numbers",
		"numeric":  "must be a number",
		"url":      "must be a valid URL",
		"phone":    "must be a valid phone number",
		"cpf":      "must be a valid CPF",
		"cnpj":     "must be a valid CNPJ",
		"datetime": "must be a valid date/time",
		"uuid":     "must be a valid UUID",
		"json":     "must be a valid JSON",
	}

	if message, exists := messages[tag]; exists {
		return fmt.Sprintf("Campo '%s' %s", field, message)
	}

	return fmt.Sprintf("Field '%s' is invalid (%s)", field, tag)
}

// validateParamValue validates parameter value based on rule
func validateParamValue(value, rule string) bool {
	switch rule {
	case "numeric":
		_, err := strconv.Atoi(value)
		return err == nil
	case "uuid":
		return len(value) == 36 && strings.Count(value, "-") == 4
	case "alpha":
		for _, r := range value {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return false
			}
		}
		return true
	case "email":
		return strings.Contains(value, "@") && strings.Contains(value, ".")
	default:
		return true
	}
}

// Custom validators

// validatePhone validates Brazilian phone number
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// Remove non-numeric characters
	cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(phone, "(", ""), ")", ""), "-", "")
	cleaned = strings.ReplaceAll(strings.ReplaceAll(cleaned, " ", ""), "+55", "")

	// Brazilian phone should have 10 or 11 digits
	return len(cleaned) == 10 || len(cleaned) == 11
}

// validateCPF validates Brazilian CPF
func validateCPF(fl validator.FieldLevel) bool {
	cpf := fl.Field().String()
	// Remove non-numeric characters
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cpf, ".", ""), "-", "")

	if len(cleaned) != 11 {
		return false
	}

	// Check if all digits are equal
	if cleaned == strings.Repeat(string(cleaned[0]), 11) {
		return false
	}

	// CPF validation algorithm
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cleaned[i]))
		sum += digit * (10 - i)
	}

	remainder := sum % 11
	var checkDigit1 int
	if remainder < 2 {
		checkDigit1 = 0
	} else {
		checkDigit1 = 11 - remainder
	}

	if checkDigit1 != int(cleaned[9]-'0') {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cleaned[i]))
		sum += digit * (11 - i)
	}

	remainder = sum % 11
	var checkDigit2 int
	if remainder < 2 {
		checkDigit2 = 0
	} else {
		checkDigit2 = 11 - remainder
	}

	return checkDigit2 == int(cleaned[10]-'0')
}

// validateCNPJ validates Brazilian CNPJ
func validateCNPJ(fl validator.FieldLevel) bool {
	cnpj := fl.Field().String()
	// Remove non-numeric characters
	cleaned := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cnpj, ".", ""), "/", ""), "-", "")

	if len(cleaned) != 14 {
		return false
	}

	// Check if all digits are equal
	if cleaned == strings.Repeat(string(cleaned[0]), 14) {
		return false
	}

	// CNPJ validation algorithm
	weights1 := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 12; i++ {
		digit, _ := strconv.Atoi(string(cleaned[i]))
		sum += digit * weights1[i]
	}

	remainder := sum % 11
	var checkDigit1 int
	if remainder < 2 {
		checkDigit1 = 0
	} else {
		checkDigit1 = 11 - remainder
	}

	if checkDigit1 != int(cleaned[12]-'0') {
		return false
	}

	weights2 := []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	sum = 0
	for i := 0; i < 13; i++ {
		digit, _ := strconv.Atoi(string(cleaned[i]))
		sum += digit * weights2[i]
	}

	remainder = sum % 11
	var checkDigit2 int
	if remainder < 2 {
		checkDigit2 = 0
	} else {
		checkDigit2 = 11 - remainder
	}

	return checkDigit2 == int(cleaned[13]-'0')
}

// validateDateTime validates date/time format
func validateDateTime(fl validator.FieldLevel) bool {
	dateTime := fl.Field().String()

	// Formatos aceitos
	formats := []string{
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, dateTime); err == nil {
			return true
		}
	}

	return false
}

// GetValidatedData extracts validated data from context
func GetValidatedData(c *gin.Context) (interface{}, bool) {
	data, exists := c.Get("validated_data")
	return data, exists
}

// GetValidatedQuery extracts validated query from context
func GetValidatedQuery(c *gin.Context) (interface{}, bool) {
	data, exists := c.Get("validated_query")
	return data, exists
}
