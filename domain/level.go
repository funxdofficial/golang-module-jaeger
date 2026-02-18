package domain

// LevelId merepresentasikan level log/trace.
// Nilai: Info, Error, Warning, Debug.
type LevelId string

const (
	LevelInfo    LevelId = "Info"
	LevelError   LevelId = "Error"
	LevelWarning LevelId = "Warning"
	LevelDebug   LevelId = "Debug"
)

// Valid mengembalikan true jika LevelId valid.
func (l LevelId) Valid() bool {
	switch l {
	case LevelInfo, LevelError, LevelWarning, LevelDebug:
		return true
	}
	return false
}
