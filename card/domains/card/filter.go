package card

type Filter struct {
	IDs       []string `json:"ids"`
	PublicIDs []string `json:"public_ids"`
	CardIDs   []string `json:"card_ids"`
	ListIDs   []string `json:"list_ids"`
	BoardIDs  []string `json:"board_ids"`
	UserIDs   []string `json:"user_ids"`
}

func (t Filter) IsEmpty() bool {
	return len(t.IDs) == 0 && len(t.PublicIDs) == 0 && len(t.CardIDs) == 0 && len(t.ListIDs) == 0 && len(t.BoardIDs) == 0 && len(t.UserIDs) == 0
}
