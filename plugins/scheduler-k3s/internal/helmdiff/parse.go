// Derived from github.com/databus23/helm-diff (Apache License 2.0).
// See LICENSE and NOTICE.md in this directory.

package helmdiff

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	hookAnnotation           = "helm.sh/hook"
	resourcePolicyAnnotation = "helm.sh/resource-policy"
)

var yamlSeparator = []byte("\n---\n")

// MappingResult stores a single Kubernetes object rendered from a Helm chart.
type MappingResult struct {
	Name           string
	Kind           string
	Content        string
	ResourcePolicy string
}

type metadata struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string
	Metadata   struct {
		Namespace   string
		Name        string
		Annotations map[string]string
	}
}

func (m metadata) String() string {
	apiBase := m.APIVersion
	sp := strings.Split(apiBase, "/")
	if len(sp) > 1 {
		apiBase = strings.Join(sp[:len(sp)-1], "/")
	}
	name := m.Metadata.Name
	if a := m.Metadata.Annotations; a != nil {
		if baseName, ok := a["helm-diff/base-name"]; ok {
			name = baseName
		}
	}
	return fmt.Sprintf("%s, %s, %s (%s)", m.Metadata.Namespace, name, m.Kind, apiBase)
}

func scanYamlSpecs(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, yamlSeparator); i >= 0 {
		return i + len(yamlSeparator), data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// Parse parses manifest bytes into a map of MappingResults keyed by
// "<namespace>, <name>, <kind> (<apiVersion>)".
func Parse(manifest []byte, defaultNamespace string, normalizeManifests bool, excludedHooks ...string) map[string]*MappingResult {
	scanner := bufio.NewScanner(io.MultiReader(strings.NewReader("\n"), bytes.NewReader(manifest)))
	scanner.Split(scanYamlSpecs)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 10485760)

	result := make(map[string]*MappingResult)

	for scanner.Scan() {
		content := bytes.TrimSpace(scanner.Bytes())
		if len(content) == 0 {
			continue
		}

		parsed, err := parseContent(content, defaultNamespace, normalizeManifests, excludedHooks...)
		if err != nil {
			log.Fatalf("%v", err)
		}

		for _, p := range parsed {
			name := p.Name

			if _, ok := result[name]; ok {
				log.Printf("Error: Found duplicate key %#v in manifest", name)
			} else {
				result[name] = p
			}
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %s", err)
	}
	return result
}

func parseContent(content []byte, defaultNamespace string, normalizeManifests bool, excludedHooks ...string) ([]*MappingResult, error) {
	var parsedMetadata metadata
	if err := yaml.Unmarshal(content, &parsedMetadata); err != nil {
		log.Fatalf("YAML unmarshal error: %s\nCan't unmarshal %s", err, content)
	}

	if parsedMetadata.APIVersion == "" && parsedMetadata.Kind == "" {
		return nil, nil
	}

	if strings.HasSuffix(parsedMetadata.Kind, "List") {
		type ListV1 struct {
			Items []yaml.MapSlice `yaml:"items"`
		}

		var list ListV1

		if err := yaml.Unmarshal(content, &list); err != nil {
			log.Fatalf("YAML unmarshal error: %s\nCan't unmarshal %s", err, content)
		}

		var result []*MappingResult

		for _, item := range list.Items {
			subcontent, err := yaml.Marshal(item)
			if err != nil {
				log.Printf("YAML marshal error: %s\nCan't marshal %v", err, item)
			}

			subs, err := parseContent(subcontent, defaultNamespace, normalizeManifests, excludedHooks...)
			if err != nil {
				return nil, fmt.Errorf("parsing YAML list item: %w", err)
			}

			result = append(result, subs...)
		}

		return result, nil
	}

	if normalizeManifests {
		var object map[interface{}]interface{}
		if err := yaml.Unmarshal(content, &object); err != nil {
			log.Fatalf("YAML unmarshal error: %s\nCan't unmarshal %s", err, content)
		}
		normalizedContent, err := yaml.Marshal(object)
		if err != nil {
			log.Fatalf("YAML marshal error: %s\nCan't marshal %v", err, object)
		}
		content = normalizedContent
	}

	if isHook(parsedMetadata, excludedHooks...) {
		return nil, nil
	}

	if parsedMetadata.Metadata.Namespace == "" {
		parsedMetadata.Metadata.Namespace = defaultNamespace
	}

	name := parsedMetadata.String()
	return []*MappingResult{
		{
			Name:           name,
			Kind:           parsedMetadata.Kind,
			Content:        string(content),
			ResourcePolicy: parsedMetadata.Metadata.Annotations[resourcePolicyAnnotation],
		},
	}, nil
}

func isHook(metadata metadata, hooks ...string) bool {
	for _, hook := range hooks {
		if metadata.Metadata.Annotations[hookAnnotation] == hook {
			return true
		}
	}
	return false
}
