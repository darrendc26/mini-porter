package config

type ClusterType string

const (
	LocalCluster   ClusterType = "local"
	GCPCluster     ClusterType = "gcp"
	CloudCluster   ClusterType = "cloud"
	UnknownCluster ClusterType = "unknown"
)

type ClusterInfo struct {
	Name   string
	Server string
	Type   ClusterType
}
