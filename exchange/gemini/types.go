package gemini

type Event struct {
	Type      string `json:"type"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Remaining string `json:"remaining"`
	Delta     string `json:"delta"`
	Reason    string `json:"reason"`
}

type Message struct {
	Type           string  `json:"type"`
	SocketSequence int     `json:"socket_sequence"`
	EventID        int     `json:"eventId"`
	Timestampms    int     `json:"timestampms"`
	Events         []Event `json:"events"`
}
