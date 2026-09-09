package mcpserver

import (
	"context"
	"fmt"

	"github.com/ars1364/warden-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type hostAddressArg struct {
	Label   string `json:"label"`
	Address string `json:"address"`
}

type listHostsArgs struct {
	HostType string `json:"host_type,omitempty" jsonschema:"optional filter: physical, vm, vps, container, switch, router, or other"`
	Tag      string `json:"tag,omitempty" jsonschema:"optional filter: only hosts carrying this tag"`
}

type hostSummary struct {
	Name             string           `json:"name"`
	HostType         string           `json:"host_type"`
	Status           string           `json:"status"`
	Tags             []string         `json:"tags"`
	ParentHostName   string           `json:"parent_host_name,omitempty"`
	LocationKind     string           `json:"location_kind,omitempty"`
	CloudProvider    string           `json:"cloud_provider,omitempty"`
	PhysicalLocation string           `json:"physical_location,omitempty"`
	Addresses        []hostAddressArg `json:"addresses,omitempty"`
}

type listHostsOutput struct {
	Hosts []hostSummary `json:"hosts"`
}

type getHostArgs struct {
	Name string `json:"name" jsonschema:"the host's name"`
}

// hostDetail deliberately never carries the SSH credential itself — SSHSecretName only
// points at where to find it, matching how the inventory stores it: fetch that secret
// separately via get_secret/get_totp_code once you actually need the value.
type hostDetail struct {
	Name             string           `json:"name"`
	HostType         string           `json:"host_type"`
	Status           string           `json:"status"`
	Description      string           `json:"description"`
	Tags             []string         `json:"tags"`
	ParentHostName   string           `json:"parent_host_name,omitempty"`
	LocationKind     string           `json:"location_kind,omitempty"`
	CloudProvider    string           `json:"cloud_provider,omitempty"`
	CloudAccount     string           `json:"cloud_account,omitempty"`
	PhysicalLocation string           `json:"physical_location,omitempty"`
	Addresses        []hostAddressArg `json:"addresses,omitempty"`
	SSHPort          int              `json:"ssh_port,omitempty"`
	SSHUsername      string           `json:"ssh_username,omitempty"`
	SSHSecretName    string           `json:"ssh_secret_name,omitempty" jsonschema:"name of a secret in this vault holding the SSH credential — call get_secret with this name to fetch it"`
	SSHJumpHostName  string           `json:"ssh_jump_host_name,omitempty"`
}

type setHostArgs struct {
	Name             string           `json:"name" jsonschema:"the host's name"`
	HostType         string           `json:"host_type,omitempty" jsonschema:"physical, vm, vps, container, switch, router, or other"`
	Status           string           `json:"status,omitempty" jsonschema:"active, maintenance, or decommissioned"`
	Description      string           `json:"description,omitempty"`
	Tags             []string         `json:"tags,omitempty"`
	ParentHostName   string           `json:"parent_host_name,omitempty" jsonschema:"name of the physical host this VM/container lives inside, if any"`
	LocationKind     string           `json:"location_kind,omitempty" jsonschema:"on_prem, cloud, or colo"`
	CloudProvider    string           `json:"cloud_provider,omitempty"`
	CloudAccount     string           `json:"cloud_account,omitempty"`
	PhysicalLocation string           `json:"physical_location,omitempty"`
	Addresses        []hostAddressArg `json:"addresses,omitempty" jsonschema:"labeled IPs/hostnames, e.g. public/private/mgmt. Replaces the full set on every call — omitting this on an update to an existing host clears its addresses, it does not leave them untouched"`
	SSHPort          int              `json:"ssh_port,omitempty"`
	SSHUsername      string           `json:"ssh_username,omitempty"`
	SSHSecretName    string           `json:"ssh_secret_name,omitempty" jsonschema:"name of a secret already in this vault holding the SSH key/password"`
	SSHJumpHostName  string           `json:"ssh_jump_host_name,omitempty" jsonschema:"name of a bastion host to connect through, if this host isn't directly reachable"`
}

type setHostOutput struct {
	Name string `json:"name"`
}

func toHostSummary(db *store.DB, h store.Host) hostSummary {
	addrs, _ := db.GetHostAddresses(h.ID)
	addrArgs := make([]hostAddressArg, len(addrs))
	for i, a := range addrs {
		addrArgs[i] = hostAddressArg{Label: a.Label, Address: a.Address}
	}
	return hostSummary{
		Name: h.Name, HostType: h.HostType, Status: h.Status, Tags: h.Tags,
		ParentHostName: resolveHostName(db, h.ParentHostID),
		LocationKind:   h.LocationKind, CloudProvider: h.CloudProvider, PhysicalLocation: h.PhysicalLocation,
		Addresses: addrArgs,
	}
}

func resolveHostName(db *store.DB, id *string) string {
	if id == nil {
		return ""
	}
	h, err := db.GetHostByID(*id)
	if err != nil {
		return ""
	}
	return h.Name
}

// resolveHostNameToID looks up a host by name for use as a parent/jump-host reference;
// empty input means "no reference", not an error.
func resolveHostNameToID(db *store.DB, name string) (*string, error) {
	if name == "" {
		return nil, nil
	}
	h, err := db.GetHostByName(name)
	if err != nil {
		return nil, fmt.Errorf("host %q not found", name)
	}
	return &h.ID, nil
}

func listHostsHandler(db *store.DB) func(context.Context, *mcp.CallToolRequest, listHostsArgs) (*mcp.CallToolResult, listHostsOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args listHostsArgs) (*mcp.CallToolResult, listHostsOutput, error) {
		hosts, err := db.ListHosts()
		if err != nil {
			return nil, listHostsOutput{}, err
		}
		out := listHostsOutput{Hosts: []hostSummary{}}
		for _, h := range hosts {
			if args.HostType != "" && h.HostType != args.HostType {
				continue
			}
			if args.Tag != "" && !containsTag(h.Tags, args.Tag) {
				continue
			}
			out.Hosts = append(out.Hosts, toHostSummary(db, h))
		}
		return nil, out, nil
	}
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func getHostHandler(db *store.DB) func(context.Context, *mcp.CallToolRequest, getHostArgs) (*mcp.CallToolResult, hostDetail, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args getHostArgs) (*mcp.CallToolResult, hostDetail, error) {
		h, err := db.GetHostByName(args.Name)
		if err != nil {
			return nil, hostDetail{}, fmt.Errorf("host %q not found", args.Name)
		}
		addrs, err := db.GetHostAddresses(h.ID)
		if err != nil {
			return nil, hostDetail{}, err
		}
		addrArgs := make([]hostAddressArg, len(addrs))
		for i, a := range addrs {
			addrArgs[i] = hostAddressArg{Label: a.Label, Address: a.Address}
		}
		return nil, hostDetail{
			Name: h.Name, HostType: h.HostType, Status: h.Status, Description: h.Description, Tags: h.Tags,
			ParentHostName: resolveHostName(db, h.ParentHostID),
			LocationKind:   h.LocationKind, CloudProvider: h.CloudProvider, CloudAccount: h.CloudAccount,
			PhysicalLocation: h.PhysicalLocation, Addresses: addrArgs,
			SSHPort: h.SSHPort, SSHUsername: h.SSHUsername, SSHSecretName: h.SSHSecretName,
			SSHJumpHostName: resolveHostName(db, h.SSHJumpHostID),
		}, nil
	}
}

func setHostHandler(db *store.DB, key *store.APIKey) func(context.Context, *mcp.CallToolRequest, setHostArgs) (*mcp.CallToolResult, setHostOutput, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args setHostArgs) (*mcp.CallToolResult, setHostOutput, error) {
		if !hasScope(key, "write") {
			return nil, setHostOutput{}, fmt.Errorf("api key %q lacks write scope", key.Name)
		}
		if args.Name == "" {
			return nil, setHostOutput{}, fmt.Errorf("name is required")
		}
		parentID, err := resolveHostNameToID(db, args.ParentHostName)
		if err != nil {
			return nil, setHostOutput{}, err
		}
		jumpID, err := resolveHostNameToID(db, args.SSHJumpHostName)
		if err != nil {
			return nil, setHostOutput{}, err
		}

		hostType := args.HostType
		if hostType == "" {
			hostType = "other"
		}
		status := args.Status
		if status == "" {
			status = "active"
		}
		in := store.HostInput{
			Name: args.Name, HostType: hostType, Status: status, Description: args.Description, Tags: args.Tags,
			ParentHostID: parentID, LocationKind: args.LocationKind, CloudProvider: args.CloudProvider,
			CloudAccount: args.CloudAccount, PhysicalLocation: args.PhysicalLocation,
			SSHPort: args.SSHPort, SSHUsername: args.SSHUsername, SSHSecretName: args.SSHSecretName,
			SSHJumpHostID: jumpID,
		}

		existing, err := db.GetHostByName(args.Name)
		var hostID string
		if err == nil {
			hostID = existing.ID
			if updErr := db.UpdateHost(hostID, in); updErr != nil {
				return nil, setHostOutput{}, updErr
			}
		} else {
			created, createErr := db.CreateHost(in, "mcp:"+key.Name)
			if createErr != nil {
				return nil, setHostOutput{}, createErr
			}
			hostID = created.ID
		}

		addrs := make([]store.HostAddress, len(args.Addresses))
		for i, a := range args.Addresses {
			addrs[i] = store.HostAddress{Label: a.Label, Address: a.Address, Position: i}
		}
		if err := db.ReplaceHostAddresses(hostID, addrs); err != nil {
			return nil, setHostOutput{}, err
		}

		return nil, setHostOutput{Name: args.Name}, nil
	}
}
