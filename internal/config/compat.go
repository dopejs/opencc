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

// DeleteProfile deletes a profile. Cannot delete "default".
func DeleteProfile(profile string) error {
	return DefaultStore().DeleteProfile(profile)
}

// ListProfiles returns sorted profile names.
func ListProfiles() []string {
	return DefaultStore().ListProfiles()
}

// --- Backward compatibility aliases for the "default" profile ---

// ReadFallbackOrder reads the default profile's provider order.
func ReadFallbackOrder() ([]string, error) {
	return ReadProfileOrder("default")
}

// WriteFallbackOrder writes the default profile's provider order.
func WriteFallbackOrder(names []string) error {
	return WriteProfileOrder("default", names)
}

// RemoveFromFallbackOrder removes a provider from the default profile.
func RemoveFromFallbackOrder(name string) error {
	return RemoveFromProfileOrder("default", name)
}
