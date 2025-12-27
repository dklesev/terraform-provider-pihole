// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

// Group represents a Pi-hole group.
type Group struct {
	ID           int64  `json:"id,omitempty"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	Description  string `json:"comment,omitempty"`
	DateAdded    int64  `json:"date_added,omitempty"`
	DateModified int64  `json:"date_modified,omitempty"`
}

// GroupsResponse represents the response from the groups endpoint.
type GroupsResponse struct {
	Groups []Group `json:"groups"`
	Took   float64 `json:"took"`
}

// Domain represents a Pi-hole domain entry (allow/deny, exact/regex).
type Domain struct {
	ID           int64   `json:"id,omitempty"`
	Domain       string  `json:"domain"`
	Type         string  `json:"type"` // "allow" or "deny"
	Kind         string  `json:"kind"` // "exact" or "regex"
	Enabled      bool    `json:"enabled"`
	Comment      string  `json:"comment,omitempty"`
	Groups       []int64 `json:"groups,omitempty"`
	DateAdded    int64   `json:"date_added,omitempty"`
	DateModified int64   `json:"date_modified,omitempty"`
}

// DomainsResponse represents the response from the domains endpoint.
type DomainsResponse struct {
	Domains []Domain `json:"domains"`
	Took    float64  `json:"took"`
}

// Client represents a Pi-hole client configuration.
type PiholeClient struct {
	ID           int64   `json:"id,omitempty"`
	Client       string  `json:"client"`
	Comment      string  `json:"comment,omitempty"`
	Groups       []int64 `json:"groups,omitempty"`
	DateAdded    int64   `json:"date_added,omitempty"`
	DateModified int64   `json:"date_modified,omitempty"`
}

// ClientsResponse represents the response from the clients endpoint.
type ClientsResponse struct {
	Clients []PiholeClient `json:"clients"`
	Took    float64        `json:"took"`
}

// List represents a Pi-hole blocklist/allowlist.
type List struct {
	ID             int64   `json:"id,omitempty"`
	Address        string  `json:"address"`
	Type           string  `json:"type"` // "block" or "allow"
	Enabled        bool    `json:"enabled"`
	Comment        string  `json:"comment,omitempty"`
	Groups         []int64 `json:"groups,omitempty"`
	DateAdded      int64   `json:"date_added,omitempty"`
	DateModified   int64   `json:"date_modified,omitempty"`
	Number         int64   `json:"number,omitempty"` // Number of domains in the list
	InvalidDomains int64   `json:"invalid_domains,omitempty"`
	Status         int     `json:"status,omitempty"`
	ABPEntries     int64   `json:"abp_entries,omitempty"`
}

// ListsResponse represents the response from the lists endpoint.
type ListsResponse struct {
	Lists []List  `json:"lists"`
	Took  float64 `json:"took"`
}

// DNSBlocking represents the DNS blocking status.
type DNSBlocking struct {
	Blocking string   `json:"blocking"` // "enabled", "disabled", "failed", "unknown"
	Timer    *float64 `json:"timer"`    // nil if no timer, otherwise seconds remaining
}

// DNSBlockingResponse represents the response from the dns/blocking endpoint.
type DNSBlockingResponse struct {
	Blocking string   `json:"blocking"`
	Timer    *float64 `json:"timer"`
	Took     float64  `json:"took"`
}

// DNSBlockingRequest represents a request to change blocking status.
type DNSBlockingRequest struct {
	Blocking bool     `json:"blocking"`
	Timer    *float64 `json:"timer,omitempty"`
}

// CNAMERecord represents a local CNAME record.
type CNAMERecord struct {
	Domain string `json:"domain"`
	Target string `json:"target"`
	TTL    int    `json:"ttl,omitempty"`
}

// LocalDNS represents a local DNS record (A/AAAA).
type LocalDNS struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

// FTLInfo represents Pi-hole FTL information.
type FTLInfo struct {
	Version  string `json:"version"`
	Branch   string `json:"branch"`
	Tag      string `json:"tag"`
	Hash     string `json:"hash"`
	Date     string `json:"date"`
	Database struct {
		Gravity int64 `json:"gravity"`
		Groups  int64 `json:"groups"`
		Clients int64 `json:"clients"`
		Lists   struct {
			Allow int64 `json:"allow"`
			Block int64 `json:"block"`
		} `json:"lists"`
		Domains struct {
			Allow struct {
				Exact int64 `json:"exact"`
				Regex int64 `json:"regex"`
			} `json:"allow"`
			Deny struct {
				Exact int64 `json:"exact"`
				Regex int64 `json:"regex"`
			} `json:"deny"`
		} `json:"domains"`
	} `json:"database"`
}

// SystemInfo represents system information.
type SystemInfo struct {
	Uptime int64 `json:"uptime"`
	Memory struct {
		RAM struct {
			Total       int64   `json:"total"`
			PercentUsed float64 `json:"perc_used"`
		} `json:"ram"`
	} `json:"memory"`
	Load    []float64 `json:"load"`
	Sensors struct {
		CPUTemp float64 `json:"cpu_temp"`
	} `json:"sensors"`
}

// InfoResponse represents the combined info response.
type InfoResponse struct {
	FTL    FTLInfo    `json:"ftl"`
	System SystemInfo `json:"system"`
	Took   float64    `json:"took"`
}

// VersionInfo represents version information.
type VersionInfo struct {
	Core struct {
		Local struct {
			Version string `json:"version"`
			Branch  string `json:"branch"`
			Hash    string `json:"hash"`
		} `json:"local"`
	} `json:"core"`
	Web struct {
		Local struct {
			Version string `json:"version"`
			Branch  string `json:"branch"`
			Hash    string `json:"hash"`
		} `json:"local"`
	} `json:"web"`
	FTL struct {
		Local struct {
			Version string `json:"version"`
			Branch  string `json:"branch"`
			Hash    string `json:"hash"`
		} `json:"local"`
	} `json:"ftl"`
}
