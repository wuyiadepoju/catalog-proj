package product

import (
	"catalog-proj/internal/app/product/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MapDomainError maps domain errors to gRPC status codes
func MapDomainError(err error) error {
	if err == nil {
		return nil
	}

	domainErr, ok := err.(*domain.DomainError)
	if !ok {
		// Unknown error, return as internal error
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}

	switch domainErr.Code {
	case domain.ErrProductNotFound.Code:
		return status.Errorf(codes.NotFound, domainErr.Message)
	case domain.ErrProductNotActive.Code:
		return status.Errorf(codes.FailedPrecondition, domainErr.Message)
	case domain.ErrInvalidDiscountPeriod.Code:
		return status.Errorf(codes.InvalidArgument, domainErr.Message)
	case domain.ErrProductAlreadyArchived.Code:
		return status.Errorf(codes.FailedPrecondition, domainErr.Message)
	case domain.ErrDiscountAlreadyActive.Code:
		return status.Errorf(codes.AlreadyExists, domainErr.Message)
	case domain.ErrInvalidPrice.Code:
		return status.Errorf(codes.InvalidArgument, domainErr.Message)
	case domain.ErrProductHasActiveDiscount.Code:
		return status.Errorf(codes.FailedPrecondition, domainErr.Message)
	default:
		return status.Errorf(codes.Internal, "unexpected error: %s", domainErr.Message)
	}
}
