package common

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmCLI "helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	DefaultReleaseName      = "consul"
	DefaultReleaseNamespace = "consul"
	chartFileName           = "Chart.yaml"
	valuesFileName          = "values.yaml"
	templatesDirName        = "templates"
	TopLevelChartDirName    = "consul"
)

// ReadChartFiles reads the chart files from the embedded file system, and loads their contents into
// []*loader.BufferedFile. This is a format that the Helm Go SDK functions can read from to create a chart to install
// from. The names of these files are important, as there are case statements in the Helm Go SDK looking for files named
// "Chart.yaml" or "templates/<templatename>.yaml", which is why even though the embedded file system has them named
// "consul/Chart.yaml" we have to strip the "consul" prefix out, which is done by the call to the helper method readFile.
func ReadChartFiles(chart embed.FS, chartDirName string) ([]*loader.BufferedFile, error) {
	var chartFiles []*loader.BufferedFile

	// Load Chart.yaml and values.yaml first.
	for _, f := range []string{chartFileName, valuesFileName} {
		file, err := readFile(chart, filepath.Join(chartDirName, f), chartDirName)
		if err != nil {
			return nil, err
		}
		chartFiles = append(chartFiles, file)
	}

	// Now load everything under templates/.
	dirs, err := chart.ReadDir(filepath.Join(chartDirName, templatesDirName))
	if err != nil {
		return nil, err
	}
	for _, f := range dirs {
		if f.IsDir() {
			// We only need to include files in the templates directory.
			continue
		}

		file, err := readFile(chart, filepath.Join(chartDirName, templatesDirName, f.Name()), chartDirName)
		if err != nil {
			return nil, err
		}
		chartFiles = append(chartFiles, file)
	}

	return chartFiles, nil
}

func readFile(chart embed.FS, f string, pathPrefix string) (*loader.BufferedFile, error) {
	bytes, err := chart.ReadFile(f)
	if err != nil {
		return nil, err
	}
	// Remove the path prefix.
	rel, err := filepath.Rel(pathPrefix, f)
	if err != nil {
		return nil, err
	}
	return &loader.BufferedFile{
		Name: rel,
		Data: bytes,
	}, nil
}

// Abort returns true if the raw input string is not equal to "y" or "yes".
func Abort(raw string) bool {
	confirmation := strings.TrimSuffix(raw, "\n")
	return !(strings.ToLower(confirmation) == "y" || strings.ToLower(confirmation) == "yes")
}

// InitActionConfig initializes a Helm Go SDK action configuration. This function currently uses a hack to override the
// namespace field that gets set in the K8s client set up by the SDK.
func InitActionConfig(actionConfig *action.Configuration, namespace string, settings *helmCLI.EnvSettings, logger action.DebugLog) (*action.Configuration, error) {
	getter := settings.RESTClientGetter()
	configFlags := getter.(*genericclioptions.ConfigFlags)
	configFlags.Namespace = &namespace
	err := actionConfig.Init(settings.RESTClientGetter(), namespace,
		os.Getenv("HELM_DRIVER"), logger)
	if err != nil {
		return nil, fmt.Errorf("error setting up helm action configuration to find existing installations: %s", err)
	}
	return actionConfig, nil
}
