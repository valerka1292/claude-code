package version

const (
	Name    = "nanocode"
	Current = "0.0.2"
)

// Display возвращает строку версии. Примечание: функция не используется в текущей версии.
func Display() string {
	return Name + " v" + Current
}
