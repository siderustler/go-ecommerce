package services

import (
	"errors"
)

type address struct {
	City       string
	Address    string
	PostalCode string
	Local      string
}

func newAddress(city, addr, postalCode, local string) address {
	return address{
		City:       city,
		Address:    addr,
		PostalCode: postalCode,
		Local:      local,
	}
}

type ShippingAddress struct {
	Address address
}

func NewShippingAddress(city, address, postalCode, local string) ShippingAddress {
	shipping := newAddress(city, address, postalCode, local)
	return ShippingAddress{
		Address: shipping,
	}
}

func (s ShippingAddress) isZero() bool {
	return s == ShippingAddress{}
}

type Billing struct {
	NIPCode string
	Company string
	Address address
}

func NewBilling(nipCode string, company string, city, address, postalCode, local string) Billing {
	billingAddress := newAddress(city, address, postalCode, local)
	return Billing{
		NIPCode: nipCode,
		Company: company,
		Address: billingAddress,
	}
}

type Customer struct {
	Name     string
	Email    string
	Phone    string
	Billing  Billing
	shipping ShippingAddress
}

func NewCustomer(name, email, phone string, billing Billing, shipping ShippingAddress) Customer {
	return Customer{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Billing:  billing,
		shipping: shipping,
	}
}

func (c Customer) Shipping() (ShippingAddress, error) {
	if !c.shipping.isZero() {
		return c.shipping, nil
	}

	return ShippingAddress{}, errors.New("shipping address is not specified")
}

type BasketProduct struct {
	Product Product
	Count   int
}

type Product struct {
	ID          string
	Name        string
	Image       string
	PriceBefore float32
	Price       float32
}

type Filter struct {
	PriceFrom              float32
	PriceTo                float32
	IncludeMachines        bool
	IncludeGardening       bool
	IncludeParts           bool
	IncludeElectro         bool
	IncludeElectroMachines bool
	Sort                   string
	Search                 string
}

func NewFilter(
	priceFrom, priceTo float32,
	includeMachines, includeGardening, includeParts, includeElectro, includeElectroMachines bool,
	sort, search string,
) Filter {
	return Filter{
		PriceFrom:              priceFrom,
		PriceTo:                priceTo,
		IncludeMachines:        includeMachines,
		IncludeGardening:       includeGardening,
		IncludeParts:           includeParts,
		IncludeElectro:         includeElectro,
		IncludeElectroMachines: includeElectroMachines,
		Sort:                   sort,
		Search:                 search,
	}
}

type ProductDetail struct {
	ID                  string
	Name                string
	Images              []string
	PriceBefore         float32
	Price               float32
	ProductInfo         []string
	TechnicalParameters []string
}

func NewProductDetail(id, name string, images, productInfo, technicalParameters []string, price float32) ProductDetail {
	return ProductDetail{
		ID:                  id,
		Name:                name,
		Images:              images,
		ProductInfo:         productInfo,
		TechnicalParameters: technicalParameters,
		Price:               price,
	}
}

func NewPromoProductDetail(id, name string, images, productInfo, technicalParameters []string, price float32, priceBefore float32) ProductDetail {
	return ProductDetail{
		ID:                  id,
		Name:                name,
		Images:              images,
		ProductInfo:         productInfo,
		TechnicalParameters: technicalParameters,
		Price:               price,
		PriceBefore:         priceBefore,
	}
}

func NewProduct(id, name, image string, price float32) Product {
	return Product{
		ID:    id,
		Name:  name,
		Price: price,
		Image: image,
	}
}

func NewPromoProduct(id, name, image string, priceBefore float32, price float32) Product {
	return Product{
		ID:          id,
		Name:        name,
		Price:       price,
		Image:       image,
		PriceBefore: priceBefore,
	}
}

func NewBasketProduct(id, name, image string, price float32, priceBefore float32, count int) BasketProduct {
	return BasketProduct{
		Product: Product{
			ID:          id,
			Name:        name,
			Image:       image,
			PriceBefore: priceBefore,
			Price:       price,
		},
		Count: count,
	}
}
