package basket

import "errors"

type ProductID string
type BasketProducts map[ProductID]BasketProduct

type Basket struct {
	ID         string
	CustomerID string
	Products   BasketProducts
}

func (b Basket) IsZero() bool {
	return b.ID == ""
}

func (b *Basket) AddProduct(productID string, count int) {
	product, inBasket := b.Products[ProductID(productID)]
	if !inBasket {
		b.Products[ProductID(productID)] = NewBasketProduct(productID, 1)
		return
	}
	product.Count += count
	b.Products[ProductID(productID)] = product
}

func (b *Basket) RemoveProduct(productID string) error {
	product, inBasket := b.Products[ProductID(productID)]
	if !inBasket {
		return errors.New("product not in basket")
	}
	product.Count -= 1
	isBasketProductDeletable := product.Count == 0
	if isBasketProductDeletable {
		delete(b.Products, ProductID(productID))
		return nil
	}
	b.Products[ProductID(productID)] = product
	return nil
}

func NewBasket(id, customerID string, products BasketProducts) Basket {
	return Basket{
		ID:         id,
		CustomerID: customerID,
		Products:   products,
	}
}

type BasketProduct struct {
	ProductID string
	Count     int
}

func NewBasketProduct(id string, count int) BasketProduct {
	return BasketProduct{
		ProductID: id,
		Count:     count,
	}
}
