package ownership

import "sync"

type PreparedDrink struct {
	OrderID  int
	WorkerID string
	Drink    string
}

type PreparationSink struct {
	mu     sync.Mutex
	drinks []PreparedDrink
}

func (s *PreparationSink) Prepare(orderID int, workerID, drink string) PreparedDrink {
	s.mu.Lock()
	defer s.mu.Unlock()

	prepared := PreparedDrink{OrderID: orderID, WorkerID: workerID, Drink: drink}
	s.drinks = append(s.drinks, prepared)
	return prepared
}

func (s *PreparationSink) Drinks() []PreparedDrink {
	s.mu.Lock()
	defer s.mu.Unlock()

	return append([]PreparedDrink(nil), s.drinks...)
}
