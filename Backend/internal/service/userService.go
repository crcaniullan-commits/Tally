package service

type UserService struct {
	db string
}

func NewUserService(db string) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create() {
}

func (s *UserService) Update() {
}

func (s *UserService) Delete() {
}

func (s *UserService) GetByID() {
}
