// shout/shout.go
package shout

// Message is a single shoutbox message.
type Message struct {
	ID        uint   `storm:"id,increment" json:"id"`
	Author    string `json:"author"`
	Body      string `json:"body"`
	CreatedAt int64  `json:"createdAt"`
}
