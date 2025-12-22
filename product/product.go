package product

func (s Services) ProductByID(id string) (Product, error) {
	return NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99), nil
}

func (s Services) ProductsByIDs(ids []string) ([]Product, error) {
	return []Product{
		NewProduct("1", "essa", "/public/products/essa/1.webp", 1.99),
		NewProduct("2", "essa", "/public/products/essa/1.webp", 1.99),
		NewProduct("3", "essa", "/public/products/essa/1.webp", 1.99),
	}, nil
}
