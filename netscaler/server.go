package netscaler

type Server struct {
	Name string `json:"name"`
	IP   string `json:"ipaddress"`
}

func NewServer(name, ip string) Server {
	return Server{name, ip}
}

func (s Server) ResourceName() string {
	return s.Name
}

func (s Server) ResourceType() string {
	return "server"
}

func (c *Client) GetServers(filter string) ([]Server, error) {
	servers := []Server{}
	err := c.query("server", filter, &servers)
	return servers, err
}

func (c *Client) AddServers(servers []Server) error {
	for _, server := range servers {
		err := c.create(server)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) RemoveServers(names []string) error {
	for _, name := range names {
		err := c.delete(name, "server")
		if err != nil {
			return err
		}
	}

	return nil
}
