package network

// Config represents configuration for the application to inform which interfaces
// to listen and broadcast to and also which port to bind to.
type Config interface {
	GetInterfaces() []string
	GetPort() int
}

type _Config struct {
	Interfaces []string
	Port       int
}

func (conf _Config) GetInterfaces() []string {
	return conf.Interfaces
}

func (conf _Config) GetPort() int {
	return conf.Port
}

// NewConfig initializes and returns a network configuration to be used for listening and
// broadcasting
func NewConfig(port int) Config {
	return _Config{Port: port, Interfaces: make([]string, 0, 0)}
}

// MessageEvent is the event interface that MessageListener should expect
type MessageEvent interface {
	GetMessage() string
}

type iListener interface {
}

// MessageListener is the Interface that Communication accepts to notify of messages received
type MessageListener interface {
	iListener
	HandleMessageReceived(event MessageEvent)
	HandleEndOfMessages()
}

// BroadcastEvent is the event interface that BroadcastListener should expect
type BroadcastEvent interface {
	GetBroadcastMessage() string
}

// BroadcastListener is the Interface that Communication accepts to notify of broadcasts received
type BroadcastListener interface {
	iListener
	HandleBroadcastReceived(event BroadcastEvent)
	HandleEndOfBroadcasts()
}

type _MessageEvent struct {
	message string
}

func (me _MessageEvent) GetMessage() string {
	return me.message
}

type _BroadcastEvent struct {
	broadcastMessage string
}

func (be _BroadcastEvent) GetBroadcastMessage() string {
	return be.broadcastMessage
}

// Communication defines the interface the application uses to communicate between
// nodes
type Communication interface {
	SetupCommunication(config Config)
	AddMessageListener(listener MessageListener) bool
	RemoveMessageListener(listener MessageListener) bool
	AddBroadcastListener(listener BroadcastListener) bool
	RemoveBroadcastListener(listener BroadcastListener) bool
	CloseCommunication()
}

func containsBroadcastListener(bList []BroadcastListener, bItem BroadcastListener) int {
	oldLen := len(bList)
	listeners := make([]iListener, oldLen, oldLen)
	for index, bListener := range bList {
		listeners[index] = bListener
	}
	return contains(listeners, bItem)
}

func containsMessageListener(mList []MessageListener, mItem MessageListener) int {
	oldLen := len(mList)
	listeners := make([]iListener, oldLen, oldLen)
	for index, mListener := range mList {
		listeners[index] = mListener
	}
	return contains(listeners, mItem)
}

func contains(list []iListener, item iListener) int {
	itemIndex := -1
	for index, iItem := range list {
		if iItem == item {
			itemIndex = index
			break
		}
	}
	return itemIndex
}
