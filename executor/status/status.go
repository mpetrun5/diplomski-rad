package status

type Status int

const (
	Timeout Status = iota
	NotParticipant
	Error
	Success
)

func StatusName(status Status) string {
	switch status {
	case Timeout:
		return "Timeout reached"
	case NotParticipant:
		return "Not participant"
	case Error:
		return "Error"
	case Success:
		return "Sucess"
	default:
		return "Unknown"
	}
}
