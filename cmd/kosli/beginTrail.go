package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const beginTrailShortDesc = `Begin or update a Kosli flow trail.`

const beginTrailLongDesc = beginTrailShortDesc + `

You can optionally associate the trail to a git commit using ^--commit^ (requires access to a git repo). And you  
can optionally redact some of the git commit data sent to Kosli using ^--redact-commit-info^.

^TRAIL-NAME^s must start with a letter or number, and only contain letters, numbers, ^.^, ^-^, ^_^, and ^~^.
`

const beginTrailExample = `
# begin/update a Kosli flow trail:
kosli begin trail yourTrailName \
	--flow yourFlowName \
	--description yourTrailDescription \
	--template-file /path/to/your/template/file.yml \
	--user-data /path/to/your/user-data/file.json \
	--api-token yourAPIToken \
	--org yourOrgName
`

type beginTrailOptions struct {
	payload              TrailPayload
	templateFile         string
	userDataFile         string
	flow                 string
	commitSHA            string
	redactedCommitInfo   []string
	srcRepoRoot          string
	externalURLs         map[string]string
	externalFingerprints map[string]string
}

type TrailPayload struct {
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	UserData     interface{}              `json:"user_data"`
	Commit       *gitview.BasicCommitInfo `json:"git_commit_info,omitempty"`
	ExternalURLs map[string]*URLInfo      `json:"external_urls,omitempty"`
	OriginURL    string                   `json:"origin_url,omitempty"`
}

func newBeginTrailCmd(out io.Writer) *cobra.Command {
	o := new(beginTrailOptions)
	cmd := &cobra.Command{
		Use:     "trail TRAIL-NAME",
		Short:   beginTrailShortDesc,
		Long:    beginTrailLongDesc,
		Example: beginTrailExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateSliceValues(o.redactedCommitInfo, allowedCommitRedactionValues)
			if err != nil {
				return fmt.Errorf("%s for --redact-commit-info", err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVar(&o.flow, "flow", "", flowNameFlag)
	cmd.Flags().StringVar(&o.payload.Description, "description", "", trailDescriptionFlag)
	cmd.Flags().StringVarP(&o.templateFile, "template-file", "f", "", templateFileSimpleFlag)
	cmd.Flags().StringVarP(&o.userDataFile, "user-data", "u", "", trailUserDataFlag)
	cmd.Flags().StringVarP(&o.commitSHA, "commit", "g", DefaultValueForCommit(ci, false), beginTrailCommitFlag)
	cmd.Flags().StringSliceVar(&o.redactedCommitInfo, "redact-commit-info", []string{}, attestationRedactCommitInfoFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", attestationRepoRootFlag)
	cmd.Flags().StringVarP(&o.payload.OriginURL, "origin-url", "o", DefaultValue(ci, "build-url"), attestationOriginUrlFlag)
	cmd.Flags().StringToStringVar(&o.externalFingerprints, "external-fingerprint", map[string]string{}, externalFingerprintFlag)
	cmd.Flags().StringToStringVar(&o.externalURLs, "external-url", map[string]string{}, externalURLFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *beginTrailOptions) run(args []string) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s", global.Host, global.Org, o.flow)

	o.payload.Name = args[0]

	var err error
	o.payload.UserData, err = LoadJsonData(o.userDataFile)
	if err != nil {
		return err
	}

	if o.commitSHA != "" {
		gv, err := gitview.New(o.srcRepoRoot)
		if err != nil {
			return err
		}
		commitInfo, err := gv.GetCommitInfoFromCommitSHA(o.commitSHA, false, o.redactedCommitInfo)
		if err != nil {
			return err
		}
		o.payload.Commit = &commitInfo.BasicCommitInfo
	}

	// process external urls
	o.payload.ExternalURLs, err = processExternalURLs(o.externalURLs, o.externalFingerprints)
	if err != nil {
		return err
	}

	form, err := newFlowForm(o.payload, o.templateFile, false)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodPut,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}

	res, err := kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		verb := "begun"
		if res.Resp.StatusCode == 200 {
			verb = "updated"
		}
		logger.Info("trail '%s' was %s", o.payload.Name, verb)
	}
	return err
}
