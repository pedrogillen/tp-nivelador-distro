package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/logger"
	"github.com/7574-sistemas-distribuidos/tp-nivelador/src/protocol"
)

const CONNECTION_ATTEMPTS_MAX = 6
const CONNECTION_ATTEMPS_DELAY_MS = 200

const TOTAL_BET_COMPONENTS = 5
const NAME_INDEX = 0
const SURNAME_INDEX = 1
const ID_INDEX = 2
const DATE_INDEX = 3
const BET_NUMBER_INDEX = 4

type ClientConfig struct {
	ServerHost string
	ServerPort string
	AgencyId   string
	InputFile  string
	OutputFile string
	BatchSize  int
}

type Client struct {
	conn   net.Conn
	config ClientConfig
}

type Bet struct {
	Name      string
	Surname   string
	Id        int
	Date      string
	BetNumber int
	AgencyId  int
}

func NewClient(config ClientConfig) (*Client, error) {
	conn, err := connectToServer(config.ServerHost, config.ServerPort)
	if err != nil {
		logger.Warn("connect-to-server", logger.Fail)
		return nil, err
	}

	client := &Client{conn: conn, config: config}
	return client, nil
}

func connectToServer(host, port string) (net.Conn, error) {
	const action = "connect-to-server"
	var err error
	var conn net.Conn

	logger.Info(action, logger.InProgress)
	for i := range CONNECTION_ATTEMPTS_MAX {
		conn, err = net.Dial("tcp", host+":"+port)
		if err != nil {
			logger.Warn(action, logger.Fail, "attempt", i)
			time.Sleep(CONNECTION_ATTEMPS_DELAY_MS * time.Millisecond)
			continue
		}

		logger.Info(action, logger.Success)
		break
	}

	return conn, err
}

func (client *Client) Run() error {
	const mainAction = "test-echo-server"
	defer client.conn.Close()

	agencyIdInt, err := strconv.Atoi(client.config.AgencyId)
	if err != nil {
		logger.Error("agency-id-parse", logger.Fail, "err", err)
		return err
	}

	err = protocol.SendAgencyId(client.conn, agencyIdInt)
	if err != nil {
		return err
	}

	err = client.SendBetLines(agencyIdInt, mainAction)
	if err != nil {
		return err
	}

	err = client.ReceiveBetWinners(agencyIdInt)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) ReceiveBetWinners(agencyIdInt int) error {
	outputFile, err := os.Create(client.config.OutputFile)
	if err != nil {
		logger.Error("create-output-file", logger.Fail)
		return err
	}
	defer outputFile.Close()

	bet_winner, err := protocol.ReceiveBetWinner(client.conn)
	if err != nil {
		return err
	}

	for bet_winner != nil {
		winner_line := fmt.Sprintf("%s,%s,%d,%s,%d", bet_winner.Name, bet_winner.Surname, bet_winner.Id, bet_winner.Date, bet_winner.BetNumber)

		outputFile.WriteString(winner_line + "\n")
		logger.Info("recv-winner", logger.Success)
		bet_winner, err = protocol.ReceiveBetWinner(client.conn)
		if err != nil {
			return err
		}
		// logger.Info("recv-bet-len", logger.InProgress, "bet-len", bet_len_int)
	}
	return nil
}

// func (client *Client) SendBetLines(agencyIdInt int, mainAction string) error {
// 	file, err := os.Open(client.config.InputFile)
// 	if err != nil {
// 		logger.Error("open-input-file", logger.Fail)
// 		return err
// 	}
// 	defer file.Close()

// 	scanner := bufio.NewScanner(file)
// 	messageId := 0

// 	for scanner.Scan() {
// 		// messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
// 		// logger.Info(mainAction, logger.InProgress, messageArgs...)

// 		line := scanner.Text()
// 		err := protocol.SendBetLine(line, client.conn, agencyIdInt)
// 		if err != nil {
// 			logger.Error("send-bet-line", logger.Fail)
// 			return err
// 		}
// 		messageId++
// 	}
// 	if err := scanner.Err(); err != nil {
// 		logger.Error("read-file", logger.Fail)
// 		return err
// 	}
// 	protocol.SendSeparator(client.conn)
// 	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId, "message-count", messageId)
// 	return nil
// }

func (client *Client) SendBetLines(agencyIdInt int, mainAction string) error {
	file, err := os.Open(client.config.InputFile)
	if err != nil {
		logger.Error("open-input-file", logger.Fail)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	messageId := 0
	batchedLines := []string{}
	counter := 0
	for scanner.Scan() {
		// messageArgs := []any{"agency-id", client.config.AgencyId, "message-id", messageId}
		// logger.Info(mainAction, logger.InProgress, messageArgs...)
		batchedLines = append(batchedLines, scanner.Text())
		messageId++
		counter++
		counter = counter % client.config.BatchSize
		if counter == 0 {
			err := protocol.SendBatchedLines(client.conn, batchedLines, agencyIdInt)
			if err != nil {
				logger.Error("send-batched-lines", logger.Fail, "err", err)
				return err
			}
			batchedLines = []string{}
		}
	}
	if err := scanner.Err(); err != nil {
		logger.Error("read-file", logger.Fail)
		return err
	}
	if len(batchedLines) > 0 {
		err := protocol.SendBatchedLines(client.conn, batchedLines, agencyIdInt)
		if err != nil {
			logger.Error("send-batched-lines", logger.Fail, "err", err)
			return err
		}
	}
	protocol.SendSeparator(client.conn)
	logger.Info(mainAction, logger.Success, "agency-id", client.config.AgencyId, "message-count", messageId)
	return nil
}
