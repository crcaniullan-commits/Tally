package service

type CategorieService struct {
	db string
}

func NewCategorieService(db string) *CategorieService {
	return &CategorieService{db: db}
}
