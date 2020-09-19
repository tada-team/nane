package main

type globalSettings struct {
	Addr  string `default:"localhost:8022"`
	Kozma bool
	// "default" doesnt work with tests
	MaxMessagesInRoom  int
	MaxRoomTitleLength int
	MaxUsernameLength  int
	MaxMessageLength   int
}

func (s globalSettings) GetMaxMessagesInRoom() int {
	if s.MaxMessagesInRoom != 0 {
		return s.MaxMessagesInRoom
	}
	return 1024
}

func (s globalSettings) GetMaxRoomTitleLength() int {
	if s.MaxRoomTitleLength != 0 {
		return s.MaxRoomTitleLength
	}
	return 50
}

func (s globalSettings) GetMaxUsernameLength() int {
	if s.MaxUsernameLength != 0 {
		return s.MaxUsernameLength
	}
	return 50
}

func (s globalSettings) GetMaxMessageLength() int {
	if s.MaxMessageLength != 0 {
		return s.MaxMessageLength
	}
	return 10500
}
