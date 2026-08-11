package domain

import (
	"encoding/json"
	"strings"
)

type rawDockerInspect struct {
	Config *struct {
		Env []string `json:"Env"`
	} `json:"Config"`
	Mounts []struct {
		Type        string `json:"Type"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
	} `json:"Mounts"`
	NetworkSettings *struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
			Gateway   string `json:"Gateway"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	State *struct {
		Health *struct {
			Status        string `json:"Status"`
			FailingStreak int    `json:"FailingStreak"`
			Log           []struct {
				Start    string `json:"Start"`
				End      string `json:"End"`
				ExitCode int    `json:"ExitCode"`
				Output   string `json:"Output"`
			} `json:"Log"`
		} `json:"Health"`
	} `json:"State"`
}

func ParseInspectDetails(rawJSON string) (*ContainerInspectDetails, error) {
	rawJSON = strings.TrimSpace(rawJSON)
	if rawJSON == "" {
		return &ContainerInspectDetails{}, nil
	}

	var list []rawDockerInspect
	if err := json.Unmarshal([]byte(rawJSON), &list); err != nil || len(list) == 0 {
		var single rawDockerInspect
		if err2 := json.Unmarshal([]byte(rawJSON), &single); err2 != nil {
			return &ContainerInspectDetails{}, nil
		}
		list = []rawDockerInspect{single}
	}

	item := list[0]
	details := &ContainerInspectDetails{}

	if item.Config != nil {
		details.Env = item.Config.Env
	}

	for _, m := range item.Mounts {
		details.Mounts = append(details.Mounts, MountDetail{
			Type:        m.Type,
			Source:      m.Source,
			Destination: m.Destination,
			Mode:        m.Mode,
		})
	}

	if item.NetworkSettings != nil && item.NetworkSettings.Networks != nil {
		for name, net := range item.NetworkSettings.Networks {
			details.Networks = append(details.Networks, NetworkDetail{
				Name:      name,
				IPAddress: net.IPAddress,
				Gateway:   net.Gateway,
			})
		}
	}

	if item.State != nil && item.State.Health != nil {
		h := item.State.Health
		health := &HealthDetail{
			Status:        HealthStatus(h.Status),
			FailingStreak: h.FailingStreak,
		}
		for _, entry := range h.Log {
			health.Log = append(health.Log, HealthLogEntry{
				Start:    entry.Start,
				End:      entry.End,
				ExitCode: entry.ExitCode,
				Output:   entry.Output,
			})
		}
		details.Health = health
	}

	return details, nil
}
