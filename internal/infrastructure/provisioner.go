package infrastructure

import (
	"context"
	"fmt"
)

// CloudProvider represents a cloud provider type
type CloudProvider string

const (
	CloudAWS          CloudProvider = "aws"
	CloudGCP          CloudProvider = "gcp"
	CloudAzure        CloudProvider = "azure"
	CloudDigitalOcean CloudProvider = "digitalocean"
)

// InfrastructureConfig holds infrastructure configuration
type InfrastructureConfig struct {
	Provider    CloudProvider
	Region      string
	Project     string
	Environment string
	Compute     *ComputeConfig
	Database    *DatabaseConfig
	Cache       *CacheConfig
	Storage     *StorageConfig
	Network     *NetworkConfig
	AutoScaling *AutoScalingConfig
	Monitoring  *MonitoringConfig
	Tags        map[string]string
}

// ComputeConfig for compute resources
type ComputeConfig struct {
	Type          string // ec2, ecs, eks, lambda, fargate
	InstanceType  string
	MinInstances  int
	MaxInstances  int
	CPU           string
	Memory        string
	GPU           bool
	SpotInstances bool
}

// DatabaseConfig for database resources
type DatabaseConfig struct {
	Engine          string // postgresql, mysql, mongodb, dynamodb
	Version         string
	InstanceClass   string
	Storage         int // GB
	MultiAZ         bool
	BackupRetention int // days
	ReadReplicas    int
	Encryption      bool
}

// CacheConfig for caching resources
type CacheConfig struct {
	Engine       string // redis, memcached, elasticache
	NodeType     string
	NumNodes     int
	Encryption   bool
	AutoFailover bool
}

// StorageConfig for object storage
type StorageConfig struct {
	Type       string // s3, gcs, azure-blob
	Buckets    []BucketConfig
	CDN        bool
	Versioning bool
}

// BucketConfig for storage buckets
type BucketConfig struct {
	Name       string
	Public     bool
	Encryption bool
	Lifecycle  *LifecyclePolicy
}

// LifecyclePolicy for storage lifecycle
type LifecyclePolicy struct {
	TransitionToIA      int // days
	TransitionToGlacier int
	Expiration          int
}

// NetworkConfig for networking
type NetworkConfig struct {
	VPC            *VPCConfig
	LoadBalancer   *LoadBalancerConfig
	CDN            *CDNConfig
	DNS            *DNSConfig
	WAF            bool
	DDoSProtection bool
}

// VPCConfig for VPC configuration
type VPCConfig struct {
	CIDR           string
	PublicSubnets  []string
	PrivateSubnets []string
	NATGateways    int
	VPNGateway     bool
	FlowLogs       bool
}

// LoadBalancerConfig for load balancer
type LoadBalancerConfig struct {
	Type        string // alb, nlb, clb
	Internal    bool
	CrossZone   bool
	SSL         bool
	HealthCheck string
}

// CDNConfig for CDN
type CDNConfig struct {
	Enabled         bool
	Provider        string // cloudfront, cloudflare, fastly
	CacheTTL        int
	GzipCompression bool
	HTTPSOnly       bool
}

// DNSConfig for DNS
type DNSConfig struct {
	Provider string // route53, cloudflare
	Zone     string
	Records  []DNSRecord
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Name  string
	Type  string
	Value string
	TTL   int
}

// AutoScalingConfig for auto-scaling
type AutoScalingConfig struct {
	Enabled           bool
	MinCapacity       int
	MaxCapacity       int
	TargetCPU         float64
	TargetMemory      float64
	ScaleUpCooldown   int // seconds
	ScaleDownCooldown int
	Predictive        bool
	Scheduled         []ScheduledScaling
}

// ScheduledScaling for scheduled scaling
type ScheduledScaling struct {
	Name        string
	MinCapacity int
	MaxCapacity int
	Recurrence  string // cron expression
}

// MonitoringConfig for monitoring
type MonitoringConfig struct {
	Enabled      bool
	Metrics      []string
	Alarms       []AlarmConfig
	LogRetention int // days
}

// AlarmConfig for alarms
type AlarmConfig struct {
	Name       string
	Metric     string
	Threshold  float64
	Comparison string // GreaterThan, LessThan, etc.
	Period     int    // seconds
	Actions    []string
}

// InfrastructureProvisioner provisions infrastructure
type InfrastructureProvisioner struct {
	terraformGenerator *TerraformGenerator
	awsProvider        *AWSProvider
	gcpProvider        *GCPProvider
	azureProvider      *AzureProvider
}

// NewInfrastructureProvisioner creates a new infrastructure provisioner
func NewInfrastructureProvisioner() *InfrastructureProvisioner {
	return &InfrastructureProvisioner{
		terraformGenerator: NewTerraformGenerator(),
		awsProvider:        NewAWSProvider(),
		gcpProvider:        NewGCPProvider(),
		azureProvider:      NewAzureProvider(),
	}
}

// Provision provisions infrastructure based on configuration
func (ip *InfrastructureProvisioner) Provision(ctx context.Context, config *InfrastructureConfig) (*ProvisioningResult, error) {
	switch config.Provider {
	case CloudAWS:
		return ip.awsProvider.Provision(ctx, config)
	case CloudGCP:
		return ip.gcpProvider.Provision(ctx, config)
	case CloudAzure:
		return ip.azureProvider.Provision(ctx, config)
	default:
		return nil, fmt.Errorf("unsupported cloud provider: %s", config.Provider)
	}
}

// GenerateTerraform generates Terraform configuration
func (ip *InfrastructureProvisioner) GenerateTerraform(config *InfrastructureConfig) (string, error) {
	return ip.terraformGenerator.Generate(config)
}

// EstimateCost estimates infrastructure cost
func (ip *InfrastructureProvisioner) EstimateCost(config *InfrastructureConfig) (*CostEstimate, error) {
	estimate := &CostEstimate{
		Monthly:   0,
		Yearly:    0,
		Breakdown: make(map[string]float64),
	}

	// Compute costs
	if config.Compute != nil {
		computeCost := ip.estimateComputeCost(config.Provider, config.Compute)
		estimate.Breakdown["compute"] = computeCost
		estimate.Monthly += computeCost
	}

	// Database costs
	if config.Database != nil {
		dbCost := ip.estimateDatabaseCost(config.Provider, config.Database)
		estimate.Breakdown["database"] = dbCost
		estimate.Monthly += dbCost
	}

	// Cache costs
	if config.Cache != nil {
		cacheCost := ip.estimateCacheCost(config.Provider, config.Cache)
		estimate.Breakdown["cache"] = cacheCost
		estimate.Monthly += cacheCost
	}

	// Storage costs
	if config.Storage != nil {
		storageCost := ip.estimateStorageCost(config.Provider, config.Storage)
		estimate.Breakdown["storage"] = storageCost
		estimate.Monthly += storageCost
	}

	// Network costs
	if config.Network != nil {
		networkCost := ip.estimateNetworkCost(config.Provider, config.Network)
		estimate.Breakdown["network"] = networkCost
		estimate.Monthly += networkCost
	}

	estimate.Yearly = estimate.Monthly * 12

	// Apply discounts for reserved instances
	if config.Compute != nil && !config.Compute.SpotInstances {
		estimate.ReservedInstanceSavings = estimate.Monthly * 0.3 // 30% savings
	}

	// Apply discounts for spot instances
	if config.Compute != nil && config.Compute.SpotInstances {
		estimate.SpotInstanceSavings = estimate.Monthly * 0.7 // 70% savings
	}

	return estimate, nil
}

// Helper functions for cost estimation

func (ip *InfrastructureProvisioner) estimateComputeCost(provider CloudProvider, compute *ComputeConfig) float64 {
	baseCost := 0.0

	switch provider {
	case CloudAWS:
		// Simplified AWS pricing
		switch compute.InstanceType {
		case "t3.micro":
			baseCost = 0.0104 * 730 // per hour * hours per month
		case "t3.small":
			baseCost = 0.0208 * 730
		case "t3.medium":
			baseCost = 0.0416 * 730
		case "m5.large":
			baseCost = 0.096 * 730
		case "m5.xlarge":
			baseCost = 0.192 * 730
		default:
			baseCost = 50.0
		}
	}

	return baseCost * float64(compute.MaxInstances)
}

func (ip *InfrastructureProvisioner) estimateDatabaseCost(provider CloudProvider, db *DatabaseConfig) float64 {
	baseCost := 0.0

	switch provider {
	case CloudAWS:
		// Simplified RDS pricing
		switch db.InstanceClass {
		case "db.t3.micro":
			baseCost = 0.017 * 730
		case "db.t3.small":
			baseCost = 0.034 * 730
		case "db.m5.large":
			baseCost = 0.17 * 730
		default:
			baseCost = 50.0
		}

		// Add storage cost
		baseCost += float64(db.Storage) * 0.115 // per GB per month

		// Multi-AZ doubles the cost
		if db.MultiAZ {
			baseCost *= 2
		}

		// Read replicas
		baseCost += baseCost * float64(db.ReadReplicas) * 0.5
	}

	return baseCost
}

func (ip *InfrastructureProvisioner) estimateCacheCost(provider CloudProvider, cache *CacheConfig) float64 {
	baseCost := 0.0

	switch provider {
	case CloudAWS:
		// Simplified ElastiCache pricing
		switch cache.NodeType {
		case "cache.t3.micro":
			baseCost = 0.017 * 730
		case "cache.t3.small":
			baseCost = 0.034 * 730
		case "cache.m5.large":
			baseCost = 0.136 * 730
		default:
			baseCost = 30.0
		}

		baseCost *= float64(cache.NumNodes)
	}

	return baseCost
}

func (ip *InfrastructureProvisioner) estimateStorageCost(provider CloudProvider, storage *StorageConfig) float64 {
	// Simplified S3 pricing: $0.023 per GB per month
	return 100.0 * 0.023 // Assume 100GB
}

func (ip *InfrastructureProvisioner) estimateNetworkCost(provider CloudProvider, network *NetworkConfig) float64 {
	baseCost := 0.0

	// Load balancer
	if network.LoadBalancer != nil {
		baseCost += 16.20 // ALB base cost per month
	}

	// NAT gateways
	if network.VPC != nil {
		baseCost += float64(network.VPC.NATGateways) * 32.40 // per NAT gateway
	}

	// CDN
	if network.CDN != nil && network.CDN.Enabled {
		baseCost += 50.0 // Base CDN cost
	}

	return baseCost
}

// ProvisioningResult holds the result of provisioning
type ProvisioningResult struct {
	Provider  CloudProvider
	Resources map[string]string // resource type -> resource ID
	Outputs   map[string]string // output name -> output value
	Error     error
}

// CostEstimate holds cost estimation
type CostEstimate struct {
	Monthly                 float64
	Yearly                  float64
	Breakdown               map[string]float64
	ReservedInstanceSavings float64
	SpotInstanceSavings     float64
}
