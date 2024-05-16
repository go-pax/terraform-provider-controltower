package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type EmptyResultError struct {
	LastRequest interface{}
}

var ErrEmptyResult = &EmptyResultError{}

func NewEmptyResultError(lastRequest interface{}) error {
	return &EmptyResultError{
		LastRequest: lastRequest,
	}
}

func (e *EmptyResultError) Error() string {
	return "empty result"
}

func (e *EmptyResultError) Is(err error) bool {
	_, ok := err.(*EmptyResultError)
	return ok
}

func (e *EmptyResultError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

type TooManyResultsError struct {
	Count       int
	LastRequest interface{}
}

var ErrTooManyResults = &TooManyResultsError{}

func NewTooManyResultsError(count int, lastRequest interface{}) error {
	return &TooManyResultsError{
		Count:       count,
		LastRequest: lastRequest,
	}
}

func (e *TooManyResultsError) Error() string {
	return fmt.Sprintf("too many results: wanted 1, got %d", e.Count)
}

func (e *TooManyResultsError) Is(err error) bool {
	_, ok := err.(*TooManyResultsError)
	return ok
}

func (e *TooManyResultsError) As(target interface{}) bool {
	t, ok := target.(**retry.NotFoundError)
	if !ok {
		return false
	}

	*t = &retry.NotFoundError{
		Message:     e.Error(),
		LastRequest: e.LastRequest,
	}

	return true
}

// FoundFunc is function that returns false if the specified value causes a `NotFound` error to be returned.
type FoundFunc[T any] func(*T) bool

// AssertSinglePtrResult returns the single non-nil pointer value in the specified slice.
// Returns a `NotFound` error otherwise.
// If any of the specified functions return false for the value, a `NotFound` error is returned.
func AssertSinglePtrResult[T any](a []*T, fs ...FoundFunc[T]) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	} else if v := a[0]; v == nil {
		return nil, NewEmptyResultError(nil)
	} else {
		for _, f := range fs {
			if !f(v) {
				return nil, NewEmptyResultError(nil)
			}
		}
		return v, nil
	}
}

// AssertSingleValueResult returns a pointer to the single value in the specified slice of values.
// Returns a `NotFound` error otherwise.
func AssertSingleValueResult[T any](a []T, fs ...FoundFunc[T]) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	} else if l > 1 {
		return nil, NewTooManyResultsError(l, nil)
	} else if v := &a[0]; v == nil {
		return nil, NewEmptyResultError(nil)
	} else {
		for _, f := range fs {
			if !f(v) {
				return nil, NewEmptyResultError(nil)
			}
		}
		return v, nil
	}
}

// AssertFirstValueResult returns a pointer to the first value in the specified slice of values.
// Returns a `NotFound` error otherwise.
func AssertFirstValueResult[T any](a []T) (*T, error) {
	if l := len(a); l == 0 {
		return nil, NewEmptyResultError(nil)
	}
	return &a[0], nil
}
