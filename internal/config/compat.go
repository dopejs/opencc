package config

import "fmt"

// --- Provider convenience functions (delegate to DefaultStore) ---

// GetProvider returns the config for a named provider, or nil.
func GetProvider(name string) *ProviderConfig {
	return DefaultStore().GetProvider(name)
}

// SetProvider creates or updates a provider and saves.
func SetProvider(name string, p *ProviderConfig) error {
	return DefaultStore().SetProvider(name, p)
}

// DeleteProviderByName removes a provider and its references from all profiles.
func DeleteProviderByName(name string) error {
	return DefaultStore().DeleteProvider(name)
}

// ProviderNames returns sorted provider names.
func ProviderNames() []string {
	return DefaultStore().ProviderNames()
}

// ExportProviderToEnv sets ANTHROPIC_* env vars for the named provider.
func ExportProviderToEnv(name string) error {
	return DefaultStore().ExportProviderToEnv(name)
}

// --- Profile convenience functions ---

// ReadProfileOrder returns the provider list for a profile.
func ReadProfileOrder(profile string) ([]string, error) {
	names := DefaultStore().GetProfileOrder(profile)
	if names == nil {
		return nil, fmt.Errorf("profile %q not found", profile)
	}
	return names, nil
}

// WriteProfileOrder sets the provider list for a profile.
func WriteProfileOrder(profile string, names []string) error {
	return DefaultStore().SetProfileOrder(profile, names)
}

// RemoveFromProfileOrder removes a provider from a profile.
func RemoveFromProfileOrder(profile, name string) error {
	return DefaultStore().RemoveFromProfile(profile, name)
}

// DeleteProfile deletes a profile. Cannot delete the default profile.
func DeleteProfile(profile string) error {
	return DefaultStore().DeleteProfile(profile)
}

// ListProfiles returns sorted profile names.
func ListProfiles() []string {
	return DefaultStore().ListProfiles()
}

// GetProfileConfig returns the full profile configuration.
func GetProfileConfig(profile string) *ProfileConfig {
	return DefaultStore().GetProfileConfig(profile)
}

// SetProfileConfig sets the full profile configuration.
func SetProfileConfig(profile string, pc *ProfileConfig) error {
	return DefaultStore().SetProfileConfig(profile, pc)
}

// --- Backward compatibility aliases for the "default" profile ---

// ReadFallbackOrder reads the default profile's provider order.
func ReadFallbackOrder() ([]string, error) {
	return ReadProfileOrder(DefaultStore().GetDefaultProfile())
}

// WriteFallbackOrder writes the default profile's provider order.
func WriteFallbackOrder(names []string) error {
	return WriteProfileOrder(DefaultStore().GetDefaultProfile(), names)
}

// RemoveFromFallbackOrder removes a provider from the default profile.
func RemoveFromFallbackOrder(name string) error {
	return RemoveFromProfileOrder(DefaultStore().GetDefaultProfile(), name)
}

// --- Global Settings convenience functions ---

// GetDefaultProfile returns the configured default profile name.
func GetDefaultProfile() string {
	return DefaultStore().GetDefaultProfile()
}

// SetDefaultProfile sets the default profile name.
func SetDefaultProfile(profile string) error {
	return DefaultStore().SetDefaultProfile(profile)
}

// GetDefaultCLI returns the configured default CLI.
func GetDefaultCLI() string {
	return DefaultStore().GetDefaultCLI()
}

// SetDefaultCLI sets the default CLI.
func SetDefaultCLI(cli string) error {
	return DefaultStore().SetDefaultCLI(cli)
}

// GetWebPort returns the configured web UI port.
func GetWebPort() int {
	return DefaultStore().GetWebPort()
}

// SetWebPort sets the web UI port.
func SetWebPort(port int) error {
	return DefaultStore().SetWebPort(port)
}

// --- Project Bindings convenience functions ---

// BindProject binds a directory path to a profile name.
func BindProject(path string, profile string) error {
	return DefaultStore().BindProject(path, profile)
}

// UnbindProject removes the binding for a directory path.
func UnbindProject(path string) error {
	return DefaultStore().UnbindProject(path)
}

// GetProjectBinding returns the profile bound to a directory path.
func GetProjectBinding(path string) string {
	return DefaultStore().GetProjectBinding(path)
}

// GetAllProjectBindings returns all project bindings.
func GetAllProjectBindings() map[string]string {
	return DefaultStore().GetAllProjectBindings()
}
