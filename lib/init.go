package lib

func Init() {
	// Initialize the map inside an init function
	genericMessage = make(map[string]string)
	genericMessage["joined"] = "I have joined the chat."
	genericMessage["welcome"] = "Welcome to chatroom."
	genericMessage["welcomeBack"] = "Welcome back!"
}
