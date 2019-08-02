// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// V4AppCatalogsResponseItemsSpec v4 app catalogs response items spec
// swagger:model v4AppCatalogsResponseItemsSpec
type V4AppCatalogsResponseItemsSpec struct {

	// A description of the catalog.
	Description string `json:"description,omitempty"`

	// A URL to a logo representing this catalog.
	LogoURL string `json:"logoURL,omitempty"`

	// storage
	Storage *V4AppCatalogsResponseItemsSpecStorage `json:"storage,omitempty"`

	// A display friendly title for this catalog.
	Title string `json:"title,omitempty"`
}

// Validate validates this v4 app catalogs response items spec
func (m *V4AppCatalogsResponseItemsSpec) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateStorage(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *V4AppCatalogsResponseItemsSpec) validateStorage(formats strfmt.Registry) error {

	if swag.IsZero(m.Storage) { // not required
		return nil
	}

	if m.Storage != nil {
		if err := m.Storage.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("storage")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *V4AppCatalogsResponseItemsSpec) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *V4AppCatalogsResponseItemsSpec) UnmarshalBinary(b []byte) error {
	var res V4AppCatalogsResponseItemsSpec
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}