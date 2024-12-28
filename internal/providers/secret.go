package providers

// Clone creates a deep copy of a secret
func (s *Secret) Clone() *Secret {
	clone := &Secret{
		Name:      s.Name,
		Value:     s.Value,
		Schema:    s.Schema,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}

	// Copy metadata
	if s.Metadata != nil {
		clone.Metadata = make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			clone.Metadata[k] = v
		}
	}

	// Copy tags
	if s.Tags != nil {
		clone.Tags = make([]string, len(s.Tags))
		copy(clone.Tags, s.Tags)
	}

	return clone
}

// String returns a string representation of the secret
func (s *Secret) String() string {
	return s.Name
}
