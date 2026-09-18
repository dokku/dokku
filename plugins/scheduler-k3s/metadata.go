package scheduler_k3s

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dokku/dokku/plugins/common"
)

// metadataField selects which of the scheduler-k3s metadata maps a command
// operates on, so the annotations and labels commands can share one
// implementation the way the storage plugin does.
type metadataField struct {
	// Name is the plural noun used in user-facing messages.
	Name string
	// Singular is the singular noun used in user-facing messages.
	Singular string
	// ClearCommand is the command an empty --replace list points the user at.
	ClearCommand string
	// PropertyPrefix namespaces the property names this field owns. Annotations
	// are stored unprefixed, which is why annotations alone must skip the
	// property names reserved by other scheduler-k3s features.
	PropertyPrefix string
	// SkipReserved drops property names owned by other scheduler-k3s features.
	SkipReserved bool
	// ResourceTypes lists the kubernetes resource types this field supports.
	ResourceTypes []string
}

var annotationsField = metadataField{
	Name:          "annotations",
	Singular:      "annotation",
	ClearCommand:  "scheduler-k3s:annotations:clear",
	SkipReserved:  true,
	ResourceTypes: AnnotationResourceTypes,
}

var labelsField = metadataField{
	Name:           "labels",
	Singular:       "label",
	ClearCommand:   "scheduler-k3s:labels:clear",
	PropertyPrefix: "labels.",
	ResourceTypes:  LabelResourceTypes,
}

// MetadataSetInput captures the inputs accepted by annotations:set and labels:set.
// Key/Value and Pairs are mutually exclusive: Pairs is populated only under
// --replace, where the positional arguments are key=value pairs rather than a
// single key and value.
type MetadataSetInput struct {
	AppName      string
	ProcessType  string
	ResourceType string
	Key          string
	Value        string
	Pairs        []string
	Replace      bool
}

// MetadataClearInput captures the inputs accepted by annotations:clear and labels:clear.
// An empty ProcessType or ResourceType filters nothing, matching the report commands.
type MetadataClearInput struct {
	AppName      string
	ProcessType  string
	ResourceType string
}

// metadataPropertyName returns the property a field stores one scope's map under.
func metadataPropertyName(field metadataField, processType string, resourceType string) string {
	return fmt.Sprintf("%s%s.%s", field.PropertyPrefix, processType, resourceType)
}

// verifyMetadataAppName applies the standard app name check, exempting the global scope.
func verifyMetadataAppName(appName string) error {
	if appName == "--global" {
		return nil
	}

	return common.VerifyAppName(appName)
}

// validateResourceType reports whether a resource type is one the field supports.
// An empty resource type passes, as the clear commands treat it as an absent filter;
// callers that require one check for it separately.
func validateResourceType(field metadataField, resourceType string) error {
	if resourceType == "" {
		return nil
	}

	for _, valid := range field.ResourceTypes {
		if valid == resourceType {
			return nil
		}
	}

	valid := make([]string, len(field.ResourceTypes))
	copy(valid, field.ResourceTypes)
	sort.Strings(valid)
	return fmt.Errorf("Invalid resource-type specified, valid resource types include: %s", strings.Join(valid, ", "))
}

// parseMetadataPairs turns key=value arguments into a map. A pair without a '='
// or with an empty key is rejected, as is a key specified more than once, since
// a repeated key leaves the declared map ambiguous. A pair ending in '=' stores
// an empty value, which the per-key form cannot express.
func parseMetadataPairs(pairs []string) (map[string]string, error) {
	parsed := map[string]string{}
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("Invalid key=value pair: %s", pair)
		}

		if _, ok := parsed[parts[0]]; ok {
			return nil, fmt.Errorf("Duplicate key specified: %s", parts[0])
		}

		parsed[parts[0]] = parts[1]
	}

	return parsed, nil
}

// setMetadataKey writes a single key on one of the metadata maps. An empty value
// deletes just that key, leaving its siblings in place.
func setMetadataKey(field metadataField, input MetadataSetInput) error {
	property := metadataPropertyName(field, input.ProcessType, input.ResourceType)
	if input.Value == "" {
		if err := common.PropertyMapDelete("scheduler-k3s", input.AppName, property, input.Key); err != nil {
			return fmt.Errorf("Unable to delete property map entry: %w", err)
		}
		return nil
	}

	if err := common.PropertyMapSet("scheduler-k3s", input.AppName, property, input.Key, input.Value); err != nil {
		return fmt.Errorf("Unable to set property map entry: %w", err)
	}

	return nil
}

// replaceMetadata swaps one scope's whole map for the declared one in a single write.
// Every pair is parsed before anything is written, so a rejected pair leaves the
// stored map untouched.
func replaceMetadata(field metadataField, input MetadataSetInput) error {
	if len(input.Pairs) == 0 {
		return fmt.Errorf("Must specify at least one key=value pair, use %s to remove all %s", field.ClearCommand, field.Name)
	}

	parsed, err := parseMetadataPairs(input.Pairs)
	if err != nil {
		return err
	}

	property := metadataPropertyName(field, input.ProcessType, input.ResourceType)
	if err := common.PropertyMapWrite("scheduler-k3s", input.AppName, property, parsed); err != nil {
		return fmt.Errorf("Unable to write property map: %w", err)
	}

	return nil
}

// clearMetadata removes every metadata property matching the given filters.
func clearMetadata(field metadataField, input MetadataClearInput) error {
	properties, err := matchingMetadataProperties(field, input.AppName, input.ProcessType, input.ResourceType)
	if err != nil {
		return err
	}

	for _, property := range properties {
		if err := common.PropertyDelete("scheduler-k3s", input.AppName, property.PropertyName); err != nil {
			return fmt.Errorf("Unable to delete property: %w", err)
		}
	}

	return nil
}

// metadataProperty is one stored metadata scope, as found by scanning the property store.
type metadataProperty struct {
	// PropertyName is the name the scope is stored under, prefix included.
	PropertyName string
	ProcessType  string
	ResourceType string
}

// matchingMetadataProperties scans the property store for the scopes a field owns on
// appName, optionally filtered by processType/resourceType. Both the report commands
// and the clear commands read through this, so they cannot disagree about what is in scope.
func matchingMetadataProperties(field metadataField, appName string, processType string, resourceType string) ([]metadataProperty, error) {
	properties, err := common.PropertyGetAllByPrefix("scheduler-k3s", appName, field.PropertyPrefix)
	if err != nil {
		return nil, fmt.Errorf("Unable to get property list: %w", err)
	}

	knownResourceTypes := map[string]bool{}
	for _, rt := range field.ResourceTypes {
		knownResourceTypes[rt] = true
	}

	matching := []metadataProperty{}
	for propertyName := range properties {
		if field.SkipReserved && isReservedAnnotationProperty(propertyName) {
			continue
		}

		suffix := strings.TrimPrefix(propertyName, field.PropertyPrefix)
		dot := strings.LastIndex(suffix, ".")
		if dot <= 0 || dot == len(suffix)-1 {
			continue
		}

		propProcessType := suffix[:dot]
		propResourceType := suffix[dot+1:]
		if !knownResourceTypes[propResourceType] {
			continue
		}

		if processType != "" && propProcessType != processType {
			continue
		}
		if resourceType != "" && propResourceType != resourceType {
			continue
		}

		matching = append(matching, metadataProperty{
			PropertyName: propertyName,
			ProcessType:  propProcessType,
			ResourceType: propResourceType,
		})
	}

	return matching, nil
}

// commandMetadataSet implements the annotations:set and labels:set commands.
func commandMetadataSet(field metadataField, input MetadataSetInput) error {
	if err := verifyMetadataAppName(input.AppName); err != nil {
		return err
	}

	if input.ResourceType == "" {
		return fmt.Errorf("Missing resource-type")
	}

	if err := validateResourceType(field, input.ResourceType); err != nil {
		return err
	}

	if input.ProcessType == "" {
		input.ProcessType = GlobalProcessType
	}

	if input.Replace {
		return replaceMetadata(field, input)
	}

	return setMetadataKey(field, input)
}

// commandMetadataClear implements the annotations:clear and labels:clear commands.
// Unlike the set commands, an omitted process-type or resource-type is an absent
// filter rather than a default scope, matching the report commands.
func commandMetadataClear(field metadataField, input MetadataClearInput) error {
	if err := verifyMetadataAppName(input.AppName); err != nil {
		return err
	}

	if err := validateResourceType(field, input.ResourceType); err != nil {
		return err
	}

	return clearMetadata(field, input)
}
