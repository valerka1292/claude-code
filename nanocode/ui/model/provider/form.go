package provider

// Field represents a provider configuration field.
type Field int

const (
	FieldName Field = iota
	FieldBaseURL
	FieldModel
	FieldAPIKey
	FieldContextSize
)

// String returns the string representation of a field.
func (f Field) String() string {
	switch f {
	case FieldName:
		return "name"
	case FieldBaseURL:
		return "base_url"
	case FieldModel:
		return "model"
	case FieldAPIKey:
		return "api_key"
	case FieldContextSize:
		return "context_size"
	default:
		return ""
	}
}

// Form holds the provider form data during creation/editing.
type Form struct {
	Name        string
	BaseURL     string
	Model       string
	APIKey      string
	ContextSize string
}

// Reset clears all form fields.
func (f *Form) Reset() {
	f.Name = ""
	f.BaseURL = ""
	f.Model = ""
	f.APIKey = ""
	f.ContextSize = ""
}
