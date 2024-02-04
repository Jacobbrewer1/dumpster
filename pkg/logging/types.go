package logging

// Name represents the name of the application for the logging.
type Name string

// String returns the string representation of the Name.
func (n Name) String() string {
	return string(n)
}
