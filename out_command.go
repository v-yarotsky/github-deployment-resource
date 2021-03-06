package resource

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/google/go-github/v28/github"
)

type OutCommand struct {
	github GitHub
	writer io.Writer
}

func NewOutCommand(github GitHub, writer io.Writer) *OutCommand {
	return &OutCommand{
		github: github,
		writer: writer,
	}
}

func (c *OutCommand) Run(sourceDir string, request OutRequest) (OutResponse, error) {
	if request.Params.ID == nil {
		return OutResponse{}, errors.New("id is a required parameter")
	}
	if request.Params.State == nil {
		return OutResponse{}, errors.New("state is a required parameter")
	}

	idInt, err := strconv.ParseInt(*request.Params.ID, 10, 64)
	if err != nil {
		return OutResponse{}, err
	}
	fmt.Fprintln(c.writer, "getting deployment")
	deployment, err := c.github.GetDeployment(idInt)
	if err != nil {
		return OutResponse{}, err
	}

	concourseMetadata := GetConcourseMetadata()
	newStatus := &github.DeploymentStatusRequest{
		State:          request.Params.State,
		Description:    request.Params.Description,
		LogURL:         &concourseMetadata.BuildURL,
		EnvironmentURL: request.Params.EnvironmentURL,
	}

	fmt.Fprintln(c.writer, "creating deployment status")
	_, err = c.github.CreateDeploymentStatus(*deployment.ID, newStatus)
	if err != nil {
		return OutResponse{}, err
	}

	fmt.Fprintln(c.writer, "getting deployment statuses list")
	statuses, err := c.github.ListDeploymentStatuses(*deployment.ID)
	if err != nil {
		return OutResponse{}, err
	}

	return OutResponse{
		Version: Version{
			ID: *request.Params.ID,
		},
		Metadata: metadataFromDeployment(deployment, statuses),
	}, nil
}

func (c *OutCommand) fileContents(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contents)), nil
}
