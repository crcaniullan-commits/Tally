package accesskeys

type AccessKeyService struct {
	store StoreAccessKey
}

func NewAccessKeyService(store StoreAccessKey) *AccessKeyService {
	return &AccessKeyService{store: store}
}
