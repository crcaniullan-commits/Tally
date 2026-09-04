package service

type AccessKeyService struct {
	db string
}

func NewAccessKeyService(db string) *AccessKeyService {
	return &AccessKeyService{db: db}
}
