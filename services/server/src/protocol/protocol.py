import logger
import safe_socket
from lottery import Bet

ID_LENGTH = 4
BET_NUMBER_LENGTH = 4
AGENCY_ID_LENGTH = 1
TOTAL_LENGTH_BYTES = 4

def read_bet_line(client_socket):
    action = "read-bet-line"
    try:
        # logger.info(action, logger.LogResult.in_progress)
        bet_line_size_bytes = safe_socket.recv_all(client_socket, 4)
        if bet_line_size_bytes == b"" or bet_line_size_bytes == b"\x00\x00\x00\x00":
            return b""
        bet_line_size = int.from_bytes(bytes=bet_line_size_bytes, byteorder="big")
        bet_line_bytes = safe_socket.recv_all(client_socket, bet_line_size)
        return bet_line_bytes
    except Exception as e:
        logger.error(action, logger.LogResult.fail)
        raise e

def deserialize_bet_line(bet_line_bytes):
    pointer = 0
    client_name, pointer = deserialize_string(bet_line_bytes, pointer)
    client_surname, pointer = deserialize_string(bet_line_bytes, pointer)
    client_id = int.from_bytes(bet_line_bytes[pointer : pointer + ID_LENGTH], byteorder="big")
    pointer += ID_LENGTH
    date, pointer = deserialize_string(bet_line_bytes, pointer)
    bet_number = int.from_bytes(bet_line_bytes[pointer : pointer + BET_NUMBER_LENGTH], byteorder="big")
    pointer += BET_NUMBER_LENGTH
    agency_id = int.from_bytes(bet_line_bytes[pointer : pointer + AGENCY_ID_LENGTH], byteorder="big")
    return Bet(
                agency_id,
                client_name,
                client_surname,
                client_id,
                date,
                bet_number,
            )

def serialize_bet(bet: Bet) -> bytes:
    action = "serialize-bet"
    try:
        # logger.info(action, logger.LogResult.in_progress)
        first_name_bytes = serialize_string(bet.first_name)
        last_name_bytes = serialize_string(bet.last_name)
        birthdate_bytes = serialize_string(bet.birthdate)
        agency_id_bytes = bet.agency_id.to_bytes(AGENCY_ID_LENGTH, byteorder="big")
        document_bytes = bet.document.to_bytes(ID_LENGTH, byteorder="big")
        number_bytes = bet.number.to_bytes(BET_NUMBER_LENGTH, byteorder="big")
        serialized_bet = first_name_bytes + last_name_bytes + document_bytes + birthdate_bytes + number_bytes + agency_id_bytes
        total_length_bytes = len(serialized_bet).to_bytes(TOTAL_LENGTH_BYTES, byteorder="big")
        serialized_bet = total_length_bytes + serialized_bet
        # logger.info(action, logger.LogResult.success)
        return serialized_bet
    except Exception as e:
        logger.error(action, logger.LogResult.fail)
        raise e

def serialize_string(serialized_string: str) -> bytes:
    string_bytes = serialized_string.encode("utf-8")
    string_length_bytes = len(string_bytes).to_bytes(1, byteorder="big")
    return string_length_bytes + string_bytes

def deserialize_string(data: bytes, pointer: int) -> tuple[str, int]:
    string_len = data[pointer]
    pointer += 1
    string_bytes = data[pointer : pointer + string_len]
    pointer += string_len
    return string_bytes.decode("utf-8"), pointer

def send_winner_line(client_socket, bet):
    safe_socket.send_all(client_socket, serialize_bet(bet))

def send_separator(client_socket):
    safe_socket.send_all(client_socket, b"\x00\x00\x00\x00")

def read_bet_batch(client_socket):
    action = "read-bet-batch"
    try:
        # logger.info(action, logger.LogResult.in_progress)
        bet_batch_size_bytes = safe_socket.recv_all(client_socket, TOTAL_LENGTH_BYTES)
        if bet_batch_size_bytes == b"" or bet_batch_size_bytes == b"\x00\x00\x00\x00":
            return []
        bet_batch_size = int.from_bytes(bytes=bet_batch_size_bytes, byteorder="big")
        bet_batch = []
        whole_bets = safe_socket.recv_all(client_socket, bet_batch_size)
        pointer = 0
        while pointer < len(whole_bets):
            bet_line_size = int.from_bytes(whole_bets[pointer:pointer + TOTAL_LENGTH_BYTES], byteorder="big")
            pointer += TOTAL_LENGTH_BYTES
            bet_line_bytes = whole_bets[pointer:pointer + bet_line_size]
            pointer += bet_line_size
            deserialized_bet = deserialize_bet_line(bet_line_bytes)
            bet_batch.append(deserialized_bet)
        return bet_batch
    except Exception as e:
        logger.error(action, logger.LogResult.fail)
        raise e