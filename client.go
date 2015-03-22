package pusher

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Client struct {
	AppId, Key, Secret string
}

func (c *Client) trigger(channels []string, event string, _data interface{}, socket_id string) (error, string) {
	data, _ := json.Marshal(_data)

	payload, _ := json.Marshal(&Event{
		Name:     event,
		Channels: channels,
		Data:     string(data),
		SocketId: socket_id})

	path := "/apps/" + c.AppId + "/" + "events"

	u := createRequestUrl("POST", path, c.Key, c.Secret, auth_timestamp(), payload, nil)

	err, response := Request("POST", u, payload)

	return err, string(response)
}

func (c *Client) Trigger(channels []string, event string, _data interface{}) (error, string) {
	return c.trigger(channels, event, _data, "")
}

func (c *Client) TriggerExclusive(channels []string, event string, _data interface{}, socket_id string) (error, string) {
	return c.trigger(channels, event, _data, socket_id)
}

func (c *Client) Channels(additional_queries map[string]string) (error, *ChannelsList) {
	path := "/apps/" + c.AppId + "/channels"
	u := createRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)
	err, response := Request("GET", u, nil)
	return err, unmarshalledChannelsList(response)
}

func (c *Client) Channel(name string, additional_queries map[string]string) (error, *Channel) {
	path := "/apps/" + c.AppId + "/channels/" + name
	u := createRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, additional_queries)
	err, response := Request("GET", u, nil)
	return err, unmarshalledChannel(response, name)
}

func (c *Client) GetChannelUsers(name string) (error, *Users) {
	path := "/apps/" + c.AppId + "/channels/" + name + "/users"
	u := createRequestUrl("GET", path, c.Key, c.Secret, auth_timestamp(), nil, nil)
	err, response := Request("GET", u, nil)
	return err, unmarshalledChannelUsers(response)
}

func (c *Client) AuthenticatePrivateChannel(_params []byte) string {

	channel_name, socket_id := parseAuthRequestParams(_params)

	string_to_sign := strings.Join([]string{socket_id, channel_name}, ":")

	auth_signature := HMACSignature(string_to_sign, c.Secret)

	auth_string := strings.Join([]string{c.Key, auth_signature}, ":")

	_response := map[string]string{"auth": auth_string}

	response, _ := json.Marshal(_response)

	return string(response)
}

func (c *Client) AuthenticatePresenceChannel(_params []byte, presence_data interface{}) string {

	channel_name, socket_id := parseAuthRequestParams(_params)

	string_to_sign := strings.Join([]string{socket_id, channel_name}, ":")

	is_presence_channel := strings.HasPrefix(channel_name, "presence-")

	var json_user_data string
	_response := make(map[string]string)

	if is_presence_channel {
		_json_user_data, _ := json.Marshal(presence_data)
		json_user_data = string(_json_user_data)
		string_to_sign += ":" + json_user_data

		_response["channel_data"] = json_user_data
	}

	auth_signature := HMACSignature(string_to_sign, c.Secret)
	_response["auth"] = c.Key + ":" + auth_signature
	response, _ := json.Marshal(_response)

	return string(response)
}

func (c *Client) Webhook(header http.Header, body []byte) *Webhook {
	webhook := &Webhook{Key: c.Key, Secret: c.Secret, Header: header, RawBody: string(body)}
	json.Unmarshal(body, &webhook)
	return webhook
}
