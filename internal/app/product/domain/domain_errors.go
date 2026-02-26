package domain

import "fmt"

type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

var (
	ErrProductNotActive = &DomainError{
		Code:    "product_not_active",
		Message: "product is not active",
	}
	ErrInvalidDiscountPeriod = &DomainError{
		Code:    "invalid_discount_period",
		Message: "discount is not valid at the specified time",
	}
	ErrProductNotFound = &DomainError{
		Code:    "product_not_found",
		Message: "product not found",
	}
	ErrProductAlreadyArchived = &DomainError{
		Code:    "product_already_archived",
		Message: "product already archived",
	}
	ErrDiscountAlreadyActive = &DomainError{
		Code:    "discount_already_active",
		Message: "discount already active",
	}
	ErrInvalidPrice = &DomainError{
		Code:    "invalid_price",
		Message: "price is invalid",
	}
	ErrProductHasActiveDiscount = &DomainError{
		Code:    "product_has_active_discount",
		Message: "cannot deactivate product with active discount",
	}
	ErrInvalidProductName = &DomainError{
		Code:    "invalid_product_name",
		Message: "product name cannot be empty",
	}
	ErrInvalidProductDescription = &DomainError{
		Code:    "invalid_product_description",
		Message: "product description cannot be empty",
	}
	ErrInvalidProductCategory = &DomainError{
		Code:    "invalid_product_category",
		Message: "product category cannot be empty",
	}
	ErrInvalidDiscountID = &DomainError{
		Code:    "invalid_discount_id",
		Message: "discount id cannot be empty",
	}
	ErrInvalidDiscountAmount = &DomainError{
		Code:    "invalid_discount_amount",
		Message: "discount amount must be between 0 and 100%",
	}
	ErrInvalidDiscountDateRange = &DomainError{
		Code:    "invalid_discount_date_range",
		Message: "discount start date must be before end date",
	}
)
