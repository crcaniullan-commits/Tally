package categories

type CategorieService struct {
	store StoreCategorie
}

func NewCategorieService(store StoreCategorie) *CategorieService {
	return &CategorieService{store: store}
}