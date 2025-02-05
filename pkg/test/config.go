package test

import (
	"os"
	"testing"

	"github.com/kubesaw/ksctl/pkg/configuration"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

type ClusterDefinitionWithName struct {
	configuration.ClusterAccessDefinition
	ClusterName string
}

// ConfigOption an option on the configuration generated for a test
type ConfigOption func(*ClusterDefinitionWithName)

// NoToken deletes the default token set for the cluster
func NoToken() ConfigOption {
	return func(content *ClusterDefinitionWithName) {
		content.Token = ""
	}
}

// ServerAPI specifies the ServerAPI to use (default is `https://cool-server.com`)
func ServerAPI(serverAPI string) ConfigOption {
	return func(content *ClusterDefinitionWithName) {
		content.ServerAPI = serverAPI
	}
}

// ClusterName specifies the name of the server (default is `host` or `member1`)
func ClusterName(clusterName string) ConfigOption {
	return func(content *ClusterDefinitionWithName) {
		content.ClusterName = clusterName
	}
}

// ServerName specifies the name of the server (default is `cool-server.com`)
func ServerName(serverName string) ConfigOption {
	return func(content *ClusterDefinitionWithName) {
		content.ServerName = serverName
	}
}

// ClusterType specifies the cluster type (`host` or `member`)
func ClusterType(clusterType string) ConfigOption {
	return func(content *ClusterDefinitionWithName) {
		content.ClusterType = configuration.ClusterType(clusterType)
	}
}

// Host defines the configuration for the host cluster
func Host(options ...ConfigOption) ClusterDefinitionWithName {
	clusterDef := ClusterDefinitionWithName{
		ClusterName: "host",
		ClusterAccessDefinition: configuration.ClusterAccessDefinition{
			ClusterDefinition: configuration.ClusterDefinition{
				ServerAPI:   "https://cool-server.com",
				ServerName:  "cool-server.com",
				ClusterType: configuration.Host,
			},
			Token: "cool-token",
		},
	}
	return WithValues(clusterDef, options...)
}

// Member defines the configuration for a member cluster
func Member(options ...ConfigOption) ClusterDefinitionWithName {
	clusterDef := ClusterDefinitionWithName{
		ClusterName: "member1",
		ClusterAccessDefinition: configuration.ClusterAccessDefinition{
			ClusterDefinition: configuration.ClusterDefinition{
				ServerAPI:   "https://cool-server.com",
				ServerName:  "cool-server.com",
				ClusterType: configuration.Member,
			},
			Token: "cool-token",
		},
	}
	return WithValues(clusterDef, options...)
}

// WithValues applies the options on the given parameters
func WithValues(clusterDef ClusterDefinitionWithName, options ...ConfigOption) ClusterDefinitionWithName {
	for _, modify := range options {
		modify(&clusterDef)
	}
	return clusterDef
}

// NewSandboxUserConfig creates SandboxUserConfig object with the given cluster definitions
func NewSandboxUserConfig(clusterDefs ...ClusterDefinitionWithName) configuration.SandboxUserConfig {
	sandboxUserConfig := configuration.SandboxUserConfig{
		Name:                     "john",
		ClusterAccessDefinitions: map[string]configuration.ClusterAccessDefinition{},
	}
	for _, clusterDefWithName := range clusterDefs {
		sandboxUserConfig.ClusterAccessDefinitions[clusterDefWithName.ClusterName] = clusterDefWithName.ClusterAccessDefinition
	}
	return sandboxUserConfig
}

// SetFileConfig generates the configuration file to use during a test
// The file is automatically cleanup at the end of the test.
func SetFileConfig(t *testing.T, clusterDefs ...ClusterDefinitionWithName) {
	tmpFile, err := os.CreateTemp(os.TempDir(), "configFile-*.yaml")
	require.NoError(t, err)
	fileName := tmpFile.Name()
	t.Cleanup(func() {
		err := os.Remove(fileName)
		require.NoError(t, err)
		configuration.ConfigFileFlag = ""
	})

	sandboxUserConfig := NewSandboxUserConfig(clusterDefs...)
	out, err := yaml.Marshal(sandboxUserConfig)
	require.NoError(t, err)
	err = os.WriteFile(fileName, out, 0600)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())
	configuration.ConfigFileFlag = fileName
	t.Logf("config file: %s: \n%s", fileName, string(out))
}
