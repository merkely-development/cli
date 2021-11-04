package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"

	"github.com/merkely-development/reporter/internal/digest"
	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

type artifactOptions struct {
	artifactType string
	inputSha256  string
	pipelineName string
	metadata     ArtifactPayload
}

type ArtifactPayload struct {
	Sha256       string `json:"sha256"`
	Filename     string `json:"filename"`
	Description  string `json:"description"`
	Git_commit   string `json:"git_commit"`
	Is_compliant bool   `json:"is_compliant"`
	Build_url    string `json:"build_url"`
	Commit_url   string `json:"commit_url"`
}

func newArtifactCmd(out io.Writer) *cobra.Command {
	o := new(artifactOptions)
	cmd := &cobra.Command{
		Use:   "artifact ARTIFACT-NAME-OR-PATH",
		Short: "Report/Log an artifact to Merkely. ",
		Long:  artifactDesc(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("only one argument (docker image name or file/dir path) is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return fmt.Errorf("docker image name or file/dir path is required")
			}

			if o.artifactType == "" && o.inputSha256 == "" {
				return fmt.Errorf("either --type or --sha256 must be specified")
			}

			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return err
			}

			if o.inputSha256 != "" {
				if err := digest.ValidateDigest(o.inputSha256); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.inputSha256 != "" {
				o.metadata.Filename = args[0]
				o.metadata.Sha256 = o.inputSha256
			} else {
				var err error
				o.metadata.Sha256, err = GetSha256Digest(o.artifactType, args[0])
				if err != nil {
					return err
				}
				if o.artifactType == "dir" || o.artifactType == "file" {
					o.metadata.Filename = filepath.Base(args[0])
				} else {
					o.metadata.Filename = args[0]
				}
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/", global.Host, global.Owner, o.pipelineName)

			js, _ := json.MarshalIndent(o.metadata, "", "    ")

			return requests.SendPayload(js, url, global.ApiToken,
				global.MaxAPIRetries, global.DryRun, "PUT")
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.artifactType, "type", "t", "", "the type of the artifact to calculate its SHA256 fingerprint")
	cmd.Flags().StringVarP(&o.inputSha256, "sha256", "s", "", "the SHA256 fingerprint for the artifact. Only required if you don't specify --type")
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", "the Merkely pipeline name")
	cmd.Flags().StringVarP(&o.metadata.Description, "description", "d", "", "[optional] the artifact description")
	cmd.Flags().StringVarP(&o.metadata.Git_commit, "git-commit", "g", DefaultValue(ci, "git-commit"), "the git commit from which the artifact was created")
	cmd.Flags().StringVarP(&o.metadata.Build_url, "build-url", "b", DefaultValue(ci, "build-url"), "the url of CI pipeline that built the artifact")
	cmd.Flags().StringVarP(&o.metadata.Commit_url, "commit-url", "u", DefaultValue(ci, "commit-url"), "the url for the git commit that created the artifact")
	cmd.Flags().BoolVarP(&o.metadata.Is_compliant, "compliant", "C", true, "whether the artifact is compliant or not")

	err := RequireFlags(cmd, []string{"pipeline", "git-commit", "build-url", "commit-url"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func artifactDesc() string {
	return `
   Report an artifact to a pipeline in Merkely. 
   The artifact SHA256 fingerprint is calculated and reported 
   or,alternatively, can be provided directly. 
   ` + GetCIDefaultsTemplates(supportedCIs, []string{"git-commit", "build-url", "commit-url"})
}
