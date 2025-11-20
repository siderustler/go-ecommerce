package services

import (
	"errors"
	"fmt"
	"strings"
)

type address struct {
	City       string
	Address    string
	PostalCode string
	Local      string
}

func newAddress(city, addr, postalCode, local string) (address, error) {
	if strings.Trim(city, " ") == "" {
		return address{}, errors.New("city is empty")
	}
	if strings.Trim(addr, " ") == "" {
		return address{}, errors.New("address is empty")
	}
	if strings.Trim(postalCode, " ") == "" {
		return address{}, errors.New("postal code is empty")
	}
	return address{
		City:       city,
		Address:    addr,
		PostalCode: postalCode,
		Local:      local,
	}, nil
}

type ShippingAddress struct {
	Address address
}

func NewShippingAddress(city, address, postalCode, local string) (ShippingAddress, error) {
	shipping, err := newAddress(city, address, postalCode, local)
	if err != nil {
		return ShippingAddress{}, fmt.Errorf("creating address: %w", err)
	}
	return ShippingAddress{
		Address: shipping,
	}, nil
}

func (s ShippingAddress) isZero() bool {
	return s == ShippingAddress{}
}

type Billing struct {
	NIPCode string
	Company string
	Address address
}

func NewBilling(nipCode string, company string, city, address, postalCode, local string) (Billing, error) {
	if strings.Trim(nipCode, " ") == "" {
		return Billing{}, errors.New("nip code is empty")
	}
	if strings.Trim(company, " ") == "" {
		return Billing{}, errors.New("company is empty")
	}
	billingAddress, err := newAddress(city, address, postalCode, local)
	if err != nil {
		return Billing{}, fmt.Errorf("creating address: %w", err)
	}

	return Billing{
		NIPCode: nipCode,
		Company: company,
		Address: billingAddress,
	}, nil
}

type Customer struct {
	Name     string
	Email    string
	Phone    string
	Billing  Billing
	shipping ShippingAddress
}

func NewCustomer(name, email, phone string, billing Billing, shipping ShippingAddress) (Customer, error) {
	if strings.Trim(name, " ") == "" {
		return Customer{}, errors.New("name is empty")
	}
	if strings.Trim(email, " ") == "" {
		return Customer{}, errors.New("email is empty")
	}
	if strings.Trim(phone, " ") == "" {
		return Customer{}, errors.New("phone is empty")
	}

	return Customer{
		Name:     name,
		Email:    email,
		Phone:    phone,
		Billing:  billing,
		shipping: shipping,
	}, nil
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
