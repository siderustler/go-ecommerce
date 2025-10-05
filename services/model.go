package services

type Product struct {
	ID    string
	Name  string
	Image string
	Price float32
}

type ProductDetail struct {
	ID                  string
	Name                string
	Images              []string
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

func NewProduct(id, name, image string, price float32) Product {
	return Product{
		ID:    id,
		Name:  name,
		Price: price,
		Image: image,
	}
}
