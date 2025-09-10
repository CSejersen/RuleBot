package client

import "home_automation_server/integrations/halo/client/handlers"

func (c *Client) UpdateButtonVal(id string, val int) error {
	return handlers.UpdateButtonValue(c, id, val)
}
