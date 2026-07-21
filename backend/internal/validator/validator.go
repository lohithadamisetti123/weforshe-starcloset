// Package validator wraps go-playground/validator with translation into the
// API's error envelope.
//
// The point of the wrapper is the translation. Raw validator errors read like
// "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed
// on the 'min' tag" — accurate, and useless to put in front of a person. What
// the frontend needs is a field name it can attach to an input and a sentence
// it can render next to it.
package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/myntra/bharat-ai-backend/internal/utils"
)

// Validator validates request structs.
type Validator struct {
	v *validator.Validate
}

// New constructs a Validator.
func New() *Validator {
	v := validator.New()

	// Report fields by their JSON name. Without this the error says "Password"
	// while the frontend's form field is "password", and the client has to
	// lower-case-and-hope to match them up.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" {
			name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		}
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	return &Validator{v: v}
}

// Struct validates s and returns a 422 AppError carrying per-field messages.
func (val *Validator) Struct(s any) error {
	err := val.v.Struct(s)
	if err == nil {
		return nil
	}

	var invalid *validator.InvalidValidationError
	if ok := asInvalid(err, &invalid); ok {
		// The struct itself is unvalidatable (a nil pointer, a non-struct).
		// That is a programming error in the handler, not bad input, and it
		// must not be reported to the client as their fault.
		return utils.NewInternal(err)
	}

	var verrs validator.ValidationErrors
	if !asValidationErrors(err, &verrs) {
		return utils.NewInternal(err)
	}

	details := make(map[string]string, len(verrs))
	for _, fe := range verrs {
		details[fe.Field()] = message(fe)
	}

	return utils.NewValidation("Some fields need attention", details)
}

// Var validates a single value against a tag.
func (val *Validator) Var(field any, tag string) error {
	if err := val.v.Var(field, tag); err != nil {
		return utils.NewBadRequest("Invalid value", err)
	}
	return nil
}

// message renders one field error as a sentence a person can act on.
//
// Each case answers "what do I type instead?" rather than restating the rule
// that fired. "Password must be at least 8 characters" tells the shopper what
// to do; "failed on the min tag" tells them the developer's business.
func message(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return "Enter a valid email address"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be %s or more", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be %s or less", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, humanList(fe.Param()))
	case "uuid", "uuid4":
		return fmt.Sprintf("%s must be a valid identifier", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	default:
		return fmt.Sprintf("%s is not valid", field)
	}
}

// humanList turns validator's space-separated oneof parameter into readable
// prose, respecting the single quotes that protect multi-word members like
// 'Date Night'.
func humanList(param string) string {
	parts := splitOneOf(param)
	return strings.Join(parts, ", ")
}

// splitOneOf splits on spaces except inside single quotes.
func splitOneOf(param string) []string {
	var (
		out     []string
		current strings.Builder
		quoted  bool
	)
	for _, r := range param {
		switch {
		case r == '\'':
			quoted = !quoted
		case r == ' ' && !quoted:
			if current.Len() > 0 {
				out = append(out, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		out = append(out, current.String())
	}
	return out
}

// asInvalid and asValidationErrors are small type-assertion helpers. validator's
// error types are concrete, not wrapped, so errors.As is overkill and a plain
// assertion states the intent more directly.
func asInvalid(err error, target **validator.InvalidValidationError) bool {
	v, ok := err.(*validator.InvalidValidationError)
	if ok {
		*target = v
	}
	return ok
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	v, ok := err.(validator.ValidationErrors)
	if ok {
		*target = v
	}
	return ok
}
