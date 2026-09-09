package goals

type GoalService struct {
	store StoreGoal
}

func NewGoalService(store StoreGoal) *GoalService {
	return &GoalService{store: store}
}