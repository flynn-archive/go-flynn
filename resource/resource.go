package resource

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flynn/go-discoverd"
	"github.com/flynn/go-discoverd/balancer"
)

func NewServer(service, path string) (*Server, error) {
	set, err := discoverd.NewServiceSet(service)
	if err != nil {
		return nil, err
	}
	s := &Server{
		path: path,
		set:  set,
		lb:   balancer.Random(set, nil),
	}
	return s, err
}

type Server struct {
	path string
	set  discoverd.ServiceSet
	lb   balancer.LoadBalancer
}

type Resource struct {
	ID  string            `json:"id"`
	Env map[string]string `json:"env"`
}

func (s *Server) Provision() (*Resource, error) {
	server, err := s.lb.Next()
	if err != nil {
		return nil, err
	}

	res, err := http.Post(fmt.Sprintf("http://%s%s", server.Addr, s.path), "", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("resource: unexpected status code %s", res.StatusCode)
	}

	resource := &Resource{}
	if err := json.NewDecoder(res.Body).Decode(resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *Server) Close() error {
	return s.set.Close()
}
