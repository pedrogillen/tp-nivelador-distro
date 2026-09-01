package protocol

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/safe_socket"
)

const TOTAL_BET_COMPONENTS = 5
const NAME_INDEX = 0
const SURNAME_INDEX = 1
const ID_INDEX = 2
const DATE_INDEX = 3
const BET_NUMBER_INDEX = 4

type Bet struct {
	Name      string
	Surname   string
	Id        int
	Date      string
	BetNumber int
	AgencyId  int
}

func SendSeparator(conn net.Conn) error {
	separator := []byte{0, 0, 0, 0}
	err := safe_socket.SendAll(conn, separator)
	if err != nil {
		logger.Error("send-separator", logger.Fail, "error", err)
		return err
	}
	return nil
}

func SendBetLine(line string, conn net.Conn, agencyId int) error {
	total_packet, err := SerializeBetLine(line, agencyId)
	if err != nil {
		return err
	}

	//logger.Info("send-bet-lines", logger.InProgress, "name", name, "surname", surname, "id", id, "date", date, "bet_number", bet_number)

	err = safe_socket.SendAll(conn, total_packet)
	if err != nil {
		logger.Error("send-bet-lines", logger.Fail, "error", err)
		return err
	}
	return nil
}

func SerializeBetLine(line string, agencyId int) ([]byte, error) {
	components := strings.Split(line, ",")
	if len(components) != TOTAL_BET_COMPONENTS {
		logger.Error("send-bet-lines", logger.Fail, "invalid-line", line)
		return nil, fmt.Errorf("invalid line: %s", line)
	}

	s_name := SerializeString(components[NAME_INDEX])
	s_surname := SerializeString(components[SURNAME_INDEX])
	s_id, err := SerializeUint32(components[ID_INDEX])
	if err != nil {
		return nil, err
	}
	s_date := SerializeString(components[DATE_INDEX])
	s_bet_number, err := SerializeUint32(components[BET_NUMBER_INDEX])
	if err != nil {
		return nil, err
	}
	s_agency_id := SerializeAgencyId(agencyId)
	data_packet := append(s_name, s_surname...)
	data_packet = append(data_packet, s_id...)
	data_packet = append(data_packet, s_date...)
	data_packet = append(data_packet, s_bet_number...)
	data_packet = append(data_packet, s_agency_id...)
	data_packet = SerializeTotalPacketSize(data_packet)
	return data_packet, nil
}

func SerializeUint32(serialized string) ([]byte, error) {
	arr := make([]byte, 4)

	converted, err := strconv.Atoi(serialized)
	if err != nil {
		logger.Error("send-bet-lines", logger.Fail, "invalid-bet-number", serialized)
		return nil, fmt.Errorf("invalid bet number: %s", serialized)
	}

	binary.BigEndian.PutUint32(arr, uint32(converted))

	return arr, nil
}

func SerializeAgencyId(agencyId int) []byte {
	bytes := uint8(agencyId)

	return []byte{bytes}
}

func SerializeTotalPacketSize(data_packet []byte) []byte {
	arr := make([]byte, 4)
	var total_packet_size uint32 = uint32(len(data_packet))

	binary.BigEndian.PutUint32(arr, total_packet_size)
	total_packet := append(arr, data_packet...)
	return total_packet
}

func SerializeString(serialized string) []byte {
	data_packet := make([]byte, 0)
	data_packet = append(data_packet, byte(len(serialized)))
	data_packet = append(data_packet, serialized...)
	return data_packet
}

func DeserializeBet(data []byte) (*Bet, error) {
	pointer := 0
	name, err := DeserializeString(data, &pointer)
	if err != nil {
		return nil, err
	}
	surname, err := DeserializeString(data, &pointer)
	if err != nil {
		return nil, err
	}
	id, err := DeserializeUint32(data, &pointer)
	if err != nil {
		return nil, err
	}
	date, err := DeserializeString(data, &pointer)
	if err != nil {
		return nil, err
	}
	bet_number, err := DeserializeUint32(data, &pointer)
	if err != nil {
		return nil, err
	}
	agency_id, err := DeserializeAgencyId(data, &pointer)
	if err != nil {
		return nil, err
	}

	bet := Bet{
		Name:      name,
		Surname:   surname,
		Id:        id,
		Date:      date,
		BetNumber: bet_number,
		AgencyId:  agency_id,
	}
	return &bet, nil
}

func DeserializeString(data []byte, pointer *int) (string, error) {
	if len(data) < *pointer+1 {
		return "", fmt.Errorf("invalid data")
	}
	string_len := int(data[*pointer])
	*pointer++
	if len(data) < *pointer+string_len {
		return "", fmt.Errorf("invalid data")
	}
	*pointer += string_len
	return string(data[*pointer-string_len : *pointer]), nil
}

func DeserializeUint32(data []byte, pointer *int) (int, error) {
	if len(data) < *pointer+4 {
		return 0, fmt.Errorf("invalid data")
	}
	number := int(binary.BigEndian.Uint32(data[*pointer : *pointer+4]))
	*pointer += 4
	return number, nil
}

func DeserializeAgencyId(data []byte, pointer *int) (int, error) {
	if len(data) < *pointer+1 {
		return 0, fmt.Errorf("invalid data")
	}
	agency_id := int(data[*pointer])
	*pointer++
	return agency_id, nil
}

func ReceiveBetWinner(conn net.Conn) (*Bet, error) {
	bet_len_bytes, err := safe_socket.RecvAll(conn, 4)
	if err != nil {
		logger.Error("recv-bet-len", logger.Fail, "error", err)
		return nil, err
	}

	bet_len_int := int(binary.BigEndian.Uint32(bet_len_bytes))
	if bet_len_int == 0 {
		return nil, nil
	}

	bet_bytes, err := safe_socket.RecvAll(conn, bet_len_int)
	if err != nil {
		logger.Error("recv-bet", logger.Fail, "error", err)
		return nil, err
	}

	bet, err := DeserializeBet(bet_bytes)
	if err != nil {
		logger.Error("deserialize-bet", logger.Fail, "error", err)
		return nil, err
	}
	return bet, nil
}

func SendBatchedLines(conn net.Conn, lines []string, agencyId int) error {
	total_packet := make([]byte, 0)
	for _, line := range lines {
		serialized_line, err := SerializeBetLine(line, agencyId)
		if err != nil {
			return err
		}
		total_packet = append(total_packet, serialized_line...)
	}
	serialized_total_packet := SerializeTotalPacketSize(total_packet)
	err := safe_socket.SendAll(conn, serialized_total_packet)
	if err != nil {
		logger.Error("send-batched-lines", logger.Fail, "error", err)
		return err
	}
	return nil
}
