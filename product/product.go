package product

func (s Services) ProductByID(id string) (Product, error) {
	return NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99), nil
}
