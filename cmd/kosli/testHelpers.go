package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/utils"
	"github.com/stretchr/testify/require"
)

const ImageName = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"

func PullExampleImage(t *testing.T) {
	err := utils.PullDockerImage(ImageName)
	require.NoError(t, err, fmt.Sprintf("pulling example image %s should work without error", ImageName))
}

// CreateFlow creates a flow on the server
func CreateFlow(flowName string, t *testing.T) {
	o := &createFlowOptions{
		payload: FlowPayload{
			Name:        flowName,
			Description: "test flow",
			Visibility:  "private",
		},
	}

	err := o.run()
	require.NoError(t, err, "flow should be created without error")
}

// CreateArtifact creates an artifact on the server
func CreateArtifact(flowName, artifactFingerprint, artifactName string, t *testing.T) {
	o := &reportArtifactOptions{
		srcRepoRoot: "../..",
		flowName:    flowName,
		payload: ArtifactPayload{
			Fingerprint: artifactFingerprint,
			GitCommit:   "0fc1ba9876f91b215679f3649b8668085d820ab5",
			BuildUrl:    "www.yr.no",
			CommitUrl:   " www.nrk.no",
		},
	}

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}

// CreateEnv creates an env on the server
func CreateEnv(owner, envName, envType string, t *testing.T) {
	o := &createEnvOptions{
		payload: CreateEnvironmentPayload{
			Owner:       owner,
			Name:        envName,
			Type:        envType,
			Description: "test env",
		},
	}

	err := o.run()
	require.NoError(t, err, "env should be created without error")
}

func ExpectDeployment(flowName, fingerprint, envName string, t *testing.T) {
	o := &expectDeploymentOptions{
		flowName: flowName,
		payload: ExpectDeploymentPayload{
			Fingerprint: fingerprint,
			Environment: envName,
			BuildUrl:    "https://example.com",
		},
	}
	err := o.run([]string{})
	require.NoError(t, err, "deployment should be expected without error")
}

func ReportServerArtifactToEnv(paths []string, envName string, t *testing.T) {
	o := &snapshotServerOptions{
		paths: paths,
	}
	err := o.run([]string{envName})
	require.NoError(t, err, "server env should be reported without error")
}
