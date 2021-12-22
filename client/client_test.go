package client_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"heckel.io/ntfy/client"
	"heckel.io/ntfy/test"
	"testing"
	"time"
)

func TestClient_Publish_Subscribe(t *testing.T) {
	s, port := test.StartServer(t)
	defer test.StopServer(t, s, port)
	c := client.New(newTestConfig(port))

	subscriptionID := c.Subscribe("mytopic")
	time.Sleep(time.Second)

	msg, err := c.Publish("mytopic", "some message")
	require.Nil(t, err)
	require.Equal(t, "some message", msg.Message)

	msg, err = c.Publish("mytopic", "some other message",
		client.WithTitle("some title"),
		client.WithPriority("high"),
		client.WithTags([]string{"tag1", "tag 2"}))
	require.Nil(t, err)
	require.Equal(t, "some other message", msg.Message)
	require.Equal(t, "some title", msg.Title)
	require.Equal(t, []string{"tag1", "tag 2"}, msg.Tags)
	require.Equal(t, 4, msg.Priority)

	msg, err = c.Publish("mytopic", "some delayed message",
		client.WithDelay("25 hours"))
	require.Nil(t, err)
	require.Equal(t, "some delayed message", msg.Message)
	require.True(t, time.Now().Add(24*time.Hour).Unix() < msg.Time)

	msg = nextMessage(c)
	require.NotNil(t, msg)
	require.Equal(t, "some message", msg.Message)

	msg = nextMessage(c)
	require.NotNil(t, msg)
	require.Equal(t, "some other message", msg.Message)
	require.Equal(t, "some title", msg.Title)
	require.Equal(t, []string{"tag1", "tag 2"}, msg.Tags)
	require.Equal(t, 4, msg.Priority)

	msg = nextMessage(c)
	require.Nil(t, msg)

	c.Unsubscribe(subscriptionID)
	time.Sleep(200 * time.Millisecond)

	msg, err = c.Publish("mytopic", "a message that won't be received")
	require.Nil(t, err)
	require.Equal(t, "a message that won't be received", msg.Message)

	msg = nextMessage(c)
	require.Nil(t, msg)
}

func newTestConfig(port int) *client.Config {
	c := client.NewConfig()
	c.DefaultHost = fmt.Sprintf("http://127.0.0.1:%d", port)
	return c
}

func nextMessage(c *client.Client) *client.Message {
	select {
	case m := <-c.Messages:
		return m
	default:
		return nil
	}
}