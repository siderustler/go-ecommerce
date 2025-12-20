package product

type repository interface{}

type Services struct {
	repository repository
}

func NewServices(repository repository) *Services {
	return &Services{
		repository: repository,
	}
}
