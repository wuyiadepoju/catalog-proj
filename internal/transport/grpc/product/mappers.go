package product

import (
	"catalog-proj/internal/app/product/domain"
	"catalog-proj/internal/app/product/queries/get_product"
	"catalog-proj/internal/app/product/queries/list_products"
	"math/big"
	"time"

	pb "catalog-proj/proto/product/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ProtoMoneyToDomain converts proto Money to domain Money
func ProtoMoneyToDomain(pbMoney *pb.Money) *domain.Money {
	if pbMoney == nil {
		return nil
	}
	// Money is stored as cents, convert to big.Rat (amount/100)
	money := domain.NewMoney(pbMoney.Amount)
	return &money
}

// DomainMoneyToProto converts domain Money to proto Money
func DomainMoneyToProto(domainMoney *domain.Money) *pb.Money {
	if domainMoney == nil {
		return nil
	}
	// domain.Money is *big.Rat
	// Convert to cents: multiply by 100
	rat := *domainMoney
	cents := new(big.Rat).Mul(rat, big.NewRat(100, 1))
	
	// Use exact conversion when possible (when denominator is 1)
	// Otherwise, use Float64() and round to nearest int
	var amount int64
	if cents.Denom().IsInt64() && cents.Denom().Int64() == 1 {
		// Exact conversion
		amount = cents.Num().Int64()
	} else {
		// Approximate conversion - round to nearest
		amountFloat, _ := cents.Float64()
		amount = int64(amountFloat + 0.5) // Round to nearest
	}
	
	return &pb.Money{
		Amount: amount,
	}
}

// BigRatToProtoMoney converts *big.Rat to proto Money
func BigRatToProtoMoney(rat *big.Rat) *pb.Money {
	if rat == nil {
		return nil
	}
	// Convert to cents: multiply by 100
	cents := new(big.Rat).Mul(rat, big.NewRat(100, 1))
	
	// Use exact conversion when possible (when denominator is 1)
	// Otherwise, use Float64() and round to nearest int
	var amount int64
	if cents.Denom().IsInt64() && cents.Denom().Int64() == 1 {
		// Exact conversion
		amount = cents.Num().Int64()
	} else {
		// Approximate conversion - round to nearest
		amountFloat, _ := cents.Float64()
		amount = int64(amountFloat + 0.5) // Round to nearest
	}
	
	return &pb.Money{
		Amount: amount,
	}
}

// BigRatToInt64 converts *big.Rat to int64 (cents)
func BigRatToInt64(rat *big.Rat) int64 {
	if rat == nil {
		return 0
	}
	// Convert to cents: multiply by 100
	cents := new(big.Rat).Mul(rat, big.NewRat(100, 1))
	
	// Use exact conversion when possible (when denominator is 1)
	// Otherwise, use Float64() and round to nearest int
	if cents.Denom().IsInt64() && cents.Denom().Int64() == 1 {
		// Exact conversion
		return cents.Num().Int64()
	}
	// Approximate conversion - round to nearest
	amountFloat, _ := cents.Float64()
	return int64(amountFloat + 0.5) // Round to nearest
}

// ProtoDiscountToDomain converts proto Discount to domain Discount
func ProtoDiscountToDomain(pbDiscount *pb.Discount) *domain.Discount {
	if pbDiscount == nil {
		return nil
	}

	var amount *domain.Money
	if pbDiscount.Amount != nil {
		money := ProtoMoneyToDomain(pbDiscount.Amount)
		amount = money
	}

	var startDate, endDate time.Time
	if pbDiscount.StartDate != nil {
		startDate = pbDiscount.StartDate.AsTime()
	}
	if pbDiscount.EndDate != nil {
		endDate = pbDiscount.EndDate.AsTime()
	}

	return &domain.Discount{
		ID:        pbDiscount.Id,
		Amount:    amount,
		StartDate: startDate,
		EndDate:   endDate,
	}
}

// DomainDiscountToProto converts domain Discount to proto Discount
func DomainDiscountToProto(domainDiscount *domain.Discount) *pb.Discount {
	if domainDiscount == nil {
		return nil
	}

	var pbAmount *pb.Money
	if domainDiscount.Amount != nil {
		pbAmount = DomainMoneyToProto(domainDiscount.Amount)
	}

	return &pb.Discount{
		Id:        domainDiscount.ID,
		Amount:    pbAmount,
		StartDate: timestamppb.New(domainDiscount.StartDate),
		EndDate:   timestamppb.New(domainDiscount.EndDate),
	}
}

// DTOToProtoProduct converts GetProduct DTO to proto Product
func DTOToProtoProduct(dto *get_product.DTO) *pb.Product {
	if dto == nil {
		return nil
	}

	product := &pb.Product{
		Id:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		BasePrice:      BigRatToProtoMoney(dto.BasePrice),
		EffectivePrice: BigRatToProtoMoney(dto.EffectivePrice),
		Status:         dto.Status,
		CreatedAt:      timestamppb.New(dto.CreatedAt),
		UpdatedAt:      timestamppb.New(dto.UpdatedAt),
	}

	if dto.DiscountID != nil {
		product.Discount = &pb.Discount{
			Id:        *dto.DiscountID,
			Amount:    BigRatToProtoMoney(dto.DiscountAmount),
			StartDate: timestamppb.New(*dto.DiscountStartDate),
			EndDate:   timestamppb.New(*dto.DiscountEndDate),
		}
	}

	if dto.ArchivedAt != nil {
		product.ArchivedAt = timestamppb.New(*dto.ArchivedAt)
	}

	return product
}

// ListProductItemToProto converts ListProducts ProductItem to proto Product
func ListProductItemToProto(item list_products.ProductItem) *pb.Product {
	product := &pb.Product{
		Id:             item.ID,
		Name:           item.Name,
		Description:    item.Description,
		Category:       item.Category,
		BasePrice:      BigRatToProtoMoney(item.BasePrice),
		EffectivePrice: BigRatToProtoMoney(item.EffectivePrice),
		Status:         item.Status,
		CreatedAt:      timestamppb.New(item.CreatedAt),
		UpdatedAt:      timestamppb.New(item.UpdatedAt),
	}

	if item.DiscountID != nil {
		product.Discount = &pb.Discount{
			Id:        *item.DiscountID,
			Amount:    BigRatToProtoMoney(item.DiscountAmount),
			StartDate: timestamppb.New(*item.DiscountStartDate),
			EndDate:   timestamppb.New(*item.DiscountEndDate),
		}
	}

	if item.ArchivedAt != nil {
		product.ArchivedAt = timestamppb.New(*item.ArchivedAt)
	}

	return product
}
