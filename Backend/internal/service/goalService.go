package service

type GoalService struct {
	db string
}

func NewGoalService(db string) *GoalService {
	return &GoalService{db: db}
}
