package save1090

import (
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type LatLon struct {
	Latitude  float64
	Longitude float64
}

// http://woodair.net/sbs/article/barebones42_socket_data.htm
type SBSMessage struct {
	MessageType      string // Always be MSG for us.							0
	TransmissionType int    //                                                  1
	// SessionID - dummy field for us                                           2
	// AircraftID - dummy field for us                                          3
	HexIdent string // Mode S hex code                                          4
	// FlightID - dummy field for us                                            5
	Transmitted  time.Time //                                                   6 + 7
	Received     time.Time //                                                   8 + 9
	Callsign     string    //                                                  10
	Altitude     int       // Mode C altitiude (Flight Level). not AMSL        11
	Groundspeed  int       // Not airspeed. knots.                             12
	Track        int       // Degrees, 0-359                                   13
	LatLon       *LatLon   //                                                  14+15
	VerticalRate int       // feet per minute                                  16
	Squawk       int       //                                                  17
	SquawkChange bool      //                                                  18
	Emergency    bool      //                                                  19
	Ident        bool      //                                                  20
	OnGround     bool      //                                                  21
}

func decode(line string) *SBSMessage {
	parts := strings.Split(line, ",")
	l := len(parts)
	if l < 5 {
		return nil
	}

	msg := &SBSMessage{
		MessageType:      decodeString(parts, 0),
		TransmissionType: decodeInt(parts, 1),
		HexIdent:         decodeString(parts, 4),
		Transmitted:      decodeTimestamp(parts, 6, 7),
		Received:         decodeTimestamp(parts, 6, 7),
		Callsign:         decodeString(parts, 10),
		Altitude:         decodeInt(parts, 11),
		Groundspeed:      decodeInt(parts, 12),
		Track:            decodeInt(parts, 13),
		VerticalRate:     decodeInt(parts, 16),
		Squawk:           decodeInt(parts, 17),
		SquawkChange:     decodeBool(parts, 18),
		Emergency:        decodeBool(parts, 19),
		Ident:            decodeBool(parts, 20),
		OnGround:         decodeBool(parts, 21),
	}

	if parts[14] != "" && parts[15] != "" {
		msg.LatLon = &LatLon{Latitude: decodeFloat(parts, 14), Longitude: decodeFloat(parts, 15)}
	}

	return msg
}

func decodeString(parts []string, ind int) string {
	if ind >= len(parts) {
		return ""
	}

	return parts[ind]
}

func decodeBool(parts []string, ind int) bool {
	if ind >= len(parts) {
		return false
	}
	return parts[ind] == "-1"
}

func decodeInt(parts []string, ind int) int {
	if ind >= len(parts) {
		return 0
	}
	i, _ := strconv.ParseInt(parts[ind], 10, 64)
	return int(i)
}

func decodeFloat(parts []string, ind int) float64 {
	if ind >= len(parts) {
		return 0
	}
	f, _ := strconv.ParseFloat(parts[ind], 64)
	return f
}

func decodeTimestamp(parts []string, ind1, ind2 int) time.Time {
	if ind2 >= len(parts) {
		return time.Time{}
	}

	t, _ := time.Parse("2006/01/02T15:04:05.000", parts[ind1]+"T"+parts[ind2])
	return t
}
