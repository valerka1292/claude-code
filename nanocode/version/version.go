package version

const (
	Name    = "nanocode"
	Current = "0.0.2"
)

func Display() string {
	return Name + " v" + Current
}
