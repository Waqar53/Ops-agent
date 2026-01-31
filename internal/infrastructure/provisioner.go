package infrastructure
import (
	"context"
	"fmt"
)
type CloudProvider string
const (
	CloudAWS          CloudProvider = "aws"
	CloudGCP          CloudProvider = "gcp"
	CloudAzure        CloudProvider = "azure"
	CloudDigitalOcean CloudProvider = "digitalocean"
)
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
type ComputeConfig struct {
	Type          string
	InstanceType  string
	MinInstances  int
	MaxInstances  int
	CPU           string
	Memory        string
	GPU           bool
	SpotInstances bool
}
type DatabaseConfig struct {
	Engine          string
	Version         string
	InstanceClass   string
	Storage         int
	MultiAZ         bool
	BackupRetention int
	ReadReplicas    int
	Encryption      bool
}
type CacheConfig struct {
	Engine       string
	NodeType     string
	NumNodes     int
	Encryption   bool
	AutoFailover bool
}
type StorageConfig struct {
	Type       string
	Buckets    []BucketConfig
	CDN        bool
	Versioning bool
}
type BucketConfig struct {
	Name       string
	Public     bool
	Encryption bool
	Lifecycle  *LifecyclePolicy
}
type LifecyclePolicy struct {
	TransitionToIA      int
	TransitionToGlacier int
	Expiration          int
}
type NetworkConfig struct {
	VPC            *VPCConfig
	LoadBalancer   *LoadBalancerConfig
	CDN            *CDNConfig
	DNS            *DNSConfig
	WAF            bool
	DDoSProtection bool
}
type VPCConfig struct {
	CIDR           string
	PublicSubnets  []string
	PrivateSubnets []string
	NATGateways    int
	VPNGateway     bool
	FlowLogs       bool
}
type LoadBalancerConfig struct {
	Type        string
	Internal    bool
	CrossZone   bool
	SSL         bool
	HealthCheck string
}
type CDNConfig struct {
	Enabled         bool
	Provider        string
	CacheTTL        int
	GzipCompression bool
	HTTPSOnly       bool
}
type DNSConfig struct {
	Provider string
	Zone     string
	Records  []DNSRecord
}
type DNSRecord struct {
	Name  string
	Type  string
	Value string
	TTL   int
}
type AutoScalingConfig struct {
	Enabled           bool
	MinCapacity       int
	MaxCapacity       int
	TargetCPU         float64
	TargetMemory      float64
	ScaleUpCooldown   int
	ScaleDownCooldown int
	Predictive        bool
	Scheduled         []ScheduledScaling
}
type ScheduledScaling struct {
	Name        string
	MinCapacity int
	MaxCapacity int
	Recurrence  string
}
type MonitoringConfig struct {
	Enabled      bool
	Metrics      []string
	Alarms       []AlarmConfig
	LogRetention int
}
type AlarmConfig struct {
	Name       string
	Metric     string
	Threshold  float64
	Comparison string
	Period     int
	Actions    []string
}
type InfrastructureProvisioner struct {
	terraformGenerator *TerraformGenerator
	awsProvider        *AWSProvider
	gcpProvider        *GCPProvider
	azureProvider      *AzureProvider
}
func NewInfrastructureProvisioner() *InfrastructureProvisioner {
	return &InfrastructureProvisioner{
		terraformGenerator: NewTerraformGenerator(),
		awsProvider:        NewAWSProvider(),
		gcpProvider:        NewGCPProvider(),
		azureProvider:      NewAzureProvider(),
	}
}
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
func (ip *InfrastructureProvisioner) GenerateTerraform(config *InfrastructureConfig) (string, error) {
	return ip.terraformGenerator.Generate(config)
}
func (ip *InfrastructureProvisioner) EstimateCost(config *InfrastructureConfig) (*CostEstimate, error) {
	estimate := &CostEstimate{
		Monthly:   0,
		Yearly:    0,
		Breakdown: make(map[string]float64),
	}
	if config.Compute != nil {
		computeCost := ip.estimateComputeCost(config.Provider, config.Compute)
		estimate.Breakdown["compute"] = computeCost
		estimate.Monthly += computeCost
	}
	if config.Database != nil {
		dbCost := ip.estimateDatabaseCost(config.Provider, config.Database)
		estimate.Breakdown["database"] = dbCost
		estimate.Monthly += dbCost
	}
	if config.Cache != nil {
		cacheCost := ip.estimateCacheCost(config.Provider, config.Cache)
		estimate.Breakdown["cache"] = cacheCost
		estimate.Monthly += cacheCost
	}
	if config.Storage != nil {
		storageCost := ip.estimateStorageCost(config.Provider, config.Storage)
		estimate.Breakdown["storage"] = storageCost
		estimate.Monthly += storageCost
	}
	if config.Network != nil {
		networkCost := ip.estimateNetworkCost(config.Provider, config.Network)
		estimate.Breakdown["network"] = networkCost
		estimate.Monthly += networkCost
	}
	estimate.Yearly = estimate.Monthly * 12
	if config.Compute != nil && !config.Compute.SpotInstances {
		estimate.ReservedInstanceSavings = estimate.Monthly * 0.3
	}
	if config.Compute != nil && config.Compute.SpotInstances {
		estimate.SpotInstanceSavings = estimate.Monthly * 0.7
	}
	return estimate, nil
}
func (ip *InfrastructureProvisioner) estimateComputeCost(provider CloudProvider, compute *ComputeConfig) float64 {
	baseCost := 0.0
	switch provider {
	case CloudAWS:
		switch compute.InstanceType {
		case "t3.micro":
			baseCost = 0.0104 * 730
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
		baseCost += float64(db.Storage) * 0.115
		if db.MultiAZ {
			baseCost *= 2
		}
		baseCost += baseCost * float64(db.ReadReplicas) * 0.5
	}
	return baseCost
}
func (ip *InfrastructureProvisioner) estimateCacheCost(provider CloudProvider, cache *CacheConfig) float64 {
	baseCost := 0.0
	switch provider {
	case CloudAWS:
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
	return 100.0 * 0.023
}
func (ip *InfrastructureProvisioner) estimateNetworkCost(provider CloudProvider, network *NetworkConfig) float64 {
	baseCost := 0.0
	if network.LoadBalancer != nil {
		baseCost += 16.20
	}
	if network.VPC != nil {
		baseCost += float64(network.VPC.NATGateways) * 32.40
	}
	if network.CDN != nil && network.CDN.Enabled {
		baseCost += 50.0
	}
	return baseCost
}
type ProvisioningResult struct {
	Provider  CloudProvider
	Resources map[string]string
	Outputs   map[string]string
	Error     error
}
type CostEstimate struct {
	Monthly                 float64
	Yearly                  float64
	Breakdown               map[string]float64
	ReservedInstanceSavings float64
	SpotInstanceSavings     float64
}
