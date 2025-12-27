// Copyright (c) 2025 dklesev
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ========================================================================
// Config Types
// ========================================================================

// PiholeConfig represents the top-level Pi-hole configuration.
type PiholeConfig struct {
	DNS       *DNSConfig       `json:"dns,omitempty"`
	DHCP      *DHCPConfig      `json:"dhcp,omitempty"`
	NTP       *NTPConfig       `json:"ntp,omitempty"`
	Resolver  *ResolverConfig  `json:"resolver,omitempty"`
	Database  *DatabaseConfig  `json:"database,omitempty"`
	Webserver *WebserverConfig `json:"webserver,omitempty"`
	Files     *FilesConfig     `json:"files,omitempty"`
	Misc      *MiscConfig      `json:"misc,omitempty"`
	Debug     *DebugConfig     `json:"debug,omitempty"`
}

// PiholeConfigResponse represents the response from the config endpoint.
type PiholeConfigResponse struct {
	Config PiholeConfig `json:"config"`
	Took   float64      `json:"took"`
}

// ========================================================================
// DNS Config
// ========================================================================

// DNSConfig represents DNS server configuration.
type DNSConfig struct {
	Upstreams           []string            `json:"upstreams,omitempty"`
	Hosts               []string            `json:"hosts,omitempty"`
	CNAMERecords        []string            `json:"cnameRecords,omitempty"`
	RevServers          []string            `json:"revServers,omitempty"`
	Interface           string              `json:"interface,omitempty"`
	ListeningMode       string              `json:"listeningMode,omitempty"`
	Port                int                 `json:"port,omitempty"`
	DNSSEC              bool                `json:"dnssec"`
	QueryLogging        bool                `json:"queryLogging"`
	DomainNeeded        bool                `json:"domainNeeded"`
	ExpandHosts         bool                `json:"expandHosts"`
	BogusPriv           bool                `json:"bogusPriv"`
	Localise            bool                `json:"localise"`
	CNAMEDeepInspect    bool                `json:"CNAMEdeepInspect"`
	BlockESNI           bool                `json:"blockESNI"`
	EDNS0ECS            bool                `json:"EDNS0ECS"`
	IgnoreLocalhost     bool                `json:"ignoreLocalhost"`
	ShowDNSSEC          bool                `json:"showDNSSEC"`
	AnalyzeOnlyAandAAAA bool                `json:"analyzeOnlyAandAAAA"`
	PiholePTR           string              `json:"piholePTR,omitempty"`
	ReplyWhenBusy       string              `json:"replyWhenBusy,omitempty"`
	BlockTTL            int                 `json:"blockTTL,omitempty"`
	HostRecord          string              `json:"hostRecord,omitempty"`
	Domain              *DNSDomainConfig    `json:"domain,omitempty"`
	Cache               *DNSCacheConfig     `json:"cache,omitempty"`
	Blocking            *DNSBlockingConfig  `json:"blocking,omitempty"`
	SpecialDomains      *DNSSpecialDomains  `json:"specialDomains,omitempty"`
	Reply               *DNSReplyConfig     `json:"reply,omitempty"`
	RateLimit           *DNSRateLimitConfig `json:"rateLimit,omitempty"`
}

type DNSDomainConfig struct {
	Name  string `json:"name,omitempty"`
	Local bool   `json:"local"`
}

type DNSCacheConfig struct {
	Size               int `json:"size,omitempty"`
	Optimizer          int `json:"optimizer,omitempty"`
	UpstreamBlockedTTL int `json:"upstreamBlockedTTL,omitempty"`
}

type DNSBlockingConfig struct {
	Active bool   `json:"active"`
	Mode   string `json:"mode,omitempty"`
	EDNS   string `json:"edns,omitempty"`
}

type DNSSpecialDomains struct {
	MozillaCanary      bool `json:"mozillaCanary"`
	ICloudPrivateRelay bool `json:"iCloudPrivateRelay"`
	DesignatedResolver bool `json:"designatedResolver"`
}

type DNSReplyConfig struct {
	Host     *DNSReplyIPConfig `json:"host,omitempty"`
	Blocking *DNSReplyIPConfig `json:"blocking,omitempty"`
}

type DNSReplyIPConfig struct {
	Force4 bool   `json:"force4"`
	IPv4   string `json:"IPv4,omitempty"`
	Force6 bool   `json:"force6"`
	IPv6   string `json:"IPv6,omitempty"`
}

type DNSRateLimitConfig struct {
	Count    int `json:"count,omitempty"`
	Interval int `json:"interval,omitempty"`
}

// ========================================================================
// DHCP Config
// ========================================================================

// DHCPConfig represents DHCP server configuration.
type DHCPConfig struct {
	Active               bool     `json:"active"`
	Start                string   `json:"start,omitempty"`
	End                  string   `json:"end,omitempty"`
	Router               string   `json:"router,omitempty"`
	Netmask              string   `json:"netmask,omitempty"`
	LeaseTime            string   `json:"leaseTime,omitempty"`
	IPv6                 bool     `json:"ipv6"`
	RapidCommit          bool     `json:"rapidCommit"`
	MultiDNS             bool     `json:"multiDNS"`
	Logging              bool     `json:"logging"`
	IgnoreUnknownClients bool     `json:"ignoreUnknownClients"`
	Hosts                []string `json:"hosts,omitempty"`
}

// ========================================================================
// NTP Config
// ========================================================================

// NTPConfig represents NTP server configuration.
type NTPConfig struct {
	IPv4 *NTPIPConfig   `json:"ipv4,omitempty"`
	IPv6 *NTPIPConfig   `json:"ipv6,omitempty"`
	Sync *NTPSyncConfig `json:"sync,omitempty"`
}

type NTPIPConfig struct {
	Active  bool   `json:"active"`
	Address string `json:"address,omitempty"`
}

type NTPSyncConfig struct {
	Active   bool          `json:"active"`
	Server   string        `json:"server,omitempty"`
	Interval int           `json:"interval,omitempty"`
	Count    int           `json:"count,omitempty"`
	RTC      *NTPRTCConfig `json:"rtc,omitempty"`
}

type NTPRTCConfig struct {
	Set    bool   `json:"set"`
	Device string `json:"device,omitempty"`
	UTC    bool   `json:"utc"`
}

// ========================================================================
// Resolver Config
// ========================================================================

// ResolverConfig represents resolver configuration.
type ResolverConfig struct {
	ResolveIPv4  bool   `json:"resolveIPv4"`
	ResolveIPv6  bool   `json:"resolveIPv6"`
	NetworkNames bool   `json:"networkNames"`
	RefreshNames string `json:"refreshNames,omitempty"`
}

// ========================================================================
// Database Config
// ========================================================================

// DatabaseConfig represents database configuration.
type DatabaseConfig struct {
	DBImport   bool                   `json:"DBimport"`
	MaxDBDays  int                    `json:"maxDBdays,omitempty"`
	DBInterval int                    `json:"DBinterval,omitempty"`
	UseWAL     bool                   `json:"useWAL"`
	Network    *DatabaseNetworkConfig `json:"network,omitempty"`
}

type DatabaseNetworkConfig struct {
	ParseARPCache bool `json:"parseARPcache"`
	Expire        int  `json:"expire,omitempty"`
}

// ========================================================================
// Webserver Config
// ========================================================================

// WebserverConfig represents webserver configuration.
type WebserverConfig struct {
	Domain       string                    `json:"domain,omitempty"`
	ACL          string                    `json:"acl,omitempty"`
	Port         string                    `json:"port,omitempty"`
	Threads      int                       `json:"threads,omitempty"`
	Headers      []string                  `json:"headers,omitempty"`
	ServeAll     bool                      `json:"serve_all"`
	AdvancedOpts []string                  `json:"advancedOpts,omitempty"`
	Session      *WebserverSessionConfig   `json:"session,omitempty"`
	TLS          *WebserverTLSConfig       `json:"tls,omitempty"`
	Paths        *WebserverPathsConfig     `json:"paths,omitempty"`
	Interface    *WebserverInterfaceConfig `json:"interface,omitempty"`
	API          *WebserverAPIConfig       `json:"api,omitempty"`
}

type WebserverSessionConfig struct {
	Timeout int  `json:"timeout,omitempty"`
	Restore bool `json:"restore"`
}

type WebserverTLSConfig struct {
	Cert     string `json:"cert,omitempty"`
	Validity int    `json:"validity,omitempty"`
}

type WebserverPathsConfig struct {
	Webroot string `json:"webroot,omitempty"`
	Webhome string `json:"webhome,omitempty"`
	Prefix  string `json:"prefix,omitempty"`
}

type WebserverInterfaceConfig struct {
	Boxed bool   `json:"boxed"`
	Theme string `json:"theme,omitempty"`
}

type WebserverAPIConfig struct {
	MaxSessions            int                     `json:"max_sessions,omitempty"`
	PrettyJSON             bool                    `json:"prettyJSON"`
	ExcludeClients         []string                `json:"excludeClients,omitempty"`
	ExcludeDomains         []string                `json:"excludeDomains,omitempty"`
	MaxHistory             int                     `json:"maxHistory,omitempty"`
	MaxClients             int                     `json:"maxClients,omitempty"`
	ClientHistoryGlobalMax bool                    `json:"client_history_global_max"`
	AllowDestructive       bool                    `json:"allow_destructive"`
	Temp                   *WebserverAPITempConfig `json:"temp,omitempty"`
}

type WebserverAPITempConfig struct {
	Limit int    `json:"limit,omitempty"`
	Unit  string `json:"unit,omitempty"`
}

// ========================================================================
// Files Config
// ========================================================================

// FilesConfig represents file paths configuration.
type FilesConfig struct {
	PID        string          `json:"pid,omitempty"`
	Database   string          `json:"database,omitempty"`
	Gravity    string          `json:"gravity,omitempty"`
	GravityTmp string          `json:"gravity_tmp,omitempty"`
	MACVendor  string          `json:"macvendor,omitempty"`
	PCAP       string          `json:"pcap,omitempty"`
	Log        *FilesLogConfig `json:"log,omitempty"`
}

type FilesLogConfig struct {
	FTL       string `json:"ftl,omitempty"`
	DNSmasq   string `json:"dnsmasq,omitempty"`
	Webserver string `json:"webserver,omitempty"`
}

// ========================================================================
// Misc Config
// ========================================================================

// MiscConfig represents miscellaneous configuration.
type MiscConfig struct {
	PrivacyLevel    int              `json:"privacylevel,omitempty"`
	DelayStartup    int              `json:"delay_startup,omitempty"`
	Nice            int              `json:"nice,omitempty"`
	Addr2Line       bool             `json:"addr2line"`
	EtcDnsmasqD     bool             `json:"etc_dnsmasq_d"`
	DnsmasqLines    []string         `json:"dnsmasq_lines,omitempty"`
	ExtraLogging    bool             `json:"extraLogging"`
	ReadOnly        bool             `json:"readOnly"`
	NormalizeCPU    bool             `json:"normalizeCPU"`
	HideDnsmasqWarn bool             `json:"hide_dnsmasq_warn"`
	Check           *MiscCheckConfig `json:"check,omitempty"`
}

type MiscCheckConfig struct {
	Load  bool `json:"load"`
	Shmem int  `json:"shmem,omitempty"`
	Disk  int  `json:"disk,omitempty"`
}

// ========================================================================
// Debug Config
// ========================================================================

// DebugConfig represents debug flags configuration.
type DebugConfig struct {
	Database     bool `json:"database"`
	Networking   bool `json:"networking"`
	Locks        bool `json:"locks"`
	Queries      bool `json:"queries"`
	Flags        bool `json:"flags"`
	Shmem        bool `json:"shmem"`
	GC           bool `json:"gc"`
	ARP          bool `json:"arp"`
	Regex        bool `json:"regex"`
	API          bool `json:"api"`
	TLS          bool `json:"tls"`
	Overtime     bool `json:"overtime"`
	Status       bool `json:"status"`
	Caps         bool `json:"caps"`
	DNSSEC       bool `json:"dnssec"`
	Vectors      bool `json:"vectors"`
	Resolver     bool `json:"resolver"`
	EDNS0        bool `json:"edns0"`
	Clients      bool `json:"clients"`
	AliasClients bool `json:"aliasclients"`
	Events       bool `json:"events"`
	Helper       bool `json:"helper"`
	Config       bool `json:"config"`
	Inotify      bool `json:"inotify"`
	Webserver    bool `json:"webserver"`
	Extra        bool `json:"extra"`
	Reserved     bool `json:"reserved"`
	NTP          bool `json:"ntp"`
	Netlink      bool `json:"netlink"`
	Timing       bool `json:"timing"`
	All          bool `json:"all"`
}

// ========================================================================
// API Methods
// ========================================================================

// GetConfig retrieves the full Pi-hole configuration.
func (c *Client) GetConfig(ctx context.Context) (*PiholeConfig, error) {
	resp, err := c.Get(ctx, "config")
	if err != nil {
		return nil, err
	}

	var result PiholeConfigResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse config response: %w", err)
	}

	return &result.Config, nil
}

// UpdateConfig updates specific configuration options using PATCH.
// The body must be wrapped in {"config": {...}} format.
// Path should be the section name (e.g., "misc").
func (c *Client) UpdateConfig(ctx context.Context, section string, values map[string]interface{}) error {
	// Pi-hole v6 requires PATCH to /api/config with body {"config": {"section": {...}}}
	body := map[string]interface{}{
		"config": map[string]interface{}{
			section: values,
		},
	}
	_, err := c.Patch(ctx, "config", body)
	return err
}

// UpdateConfigValue updates a single configuration value.
func (c *Client) UpdateConfigValue(ctx context.Context, section, key string, value interface{}) error {
	return c.UpdateConfig(ctx, section, map[string]interface{}{key: value})
}

// GetDNSConfig retrieves only the DNS configuration.
func (c *Client) GetDNSConfig(ctx context.Context) (*DNSConfig, error) {
	config, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return config.DNS, nil
}

// GetDHCPConfig retrieves only the DHCP configuration.
func (c *Client) GetDHCPConfig(ctx context.Context) (*DHCPConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.DHCP, nil
}

// GetMiscConfig retrieves only the miscellaneous configuration.
func (c *Client) GetMiscConfig(ctx context.Context) (*MiscConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Misc, nil
}

// GetNTPConfig retrieves only the NTP configuration.
func (c *Client) GetNTPConfig(ctx context.Context) (*NTPConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.NTP, nil
}

// GetResolverConfig retrieves only the resolver configuration.
func (c *Client) GetResolverConfig(ctx context.Context) (*ResolverConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Resolver, nil
}

// GetDatabaseConfig retrieves only the database configuration.
func (c *Client) GetDatabaseConfig(ctx context.Context) (*DatabaseConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Database, nil
}

// GetWebserverConfig retrieves only the webserver configuration.
func (c *Client) GetWebserverConfig(ctx context.Context) (*WebserverConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Webserver, nil
}

// GetFilesConfig retrieves only the files configuration.
func (c *Client) GetFilesConfig(ctx context.Context) (*FilesConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Files, nil
}

// GetDebugConfig retrieves only the debug configuration.
func (c *Client) GetDebugConfig(ctx context.Context) (*DebugConfig, error) {
	piholeConfig, err := c.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	return piholeConfig.Debug, nil
}

// AddConfigArrayItem adds an item to a config array using PUT.
// Path should be like "dns/upstreams" and value is the item to add.
func (c *Client) AddConfigArrayItem(ctx context.Context, path, value string) error {
	// URL encode the value for the path
	encoded := url.PathEscape(value)
	endpoint := fmt.Sprintf("config/%s/%s", path, encoded)
	_, err := c.Put(ctx, endpoint, nil)
	return err
}

// DeleteConfigArrayItem removes an item from a config array using DELETE.
func (c *Client) DeleteConfigArrayItem(ctx context.Context, path, value string) error {
	encoded := url.PathEscape(value)
	endpoint := fmt.Sprintf("config/%s/%s", path, encoded)
	_, err := c.Delete(ctx, endpoint)
	return err
}
