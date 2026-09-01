import socket
import logger
import safe_socket
from lottery import Lottery
import protocol

class Server:
    def __init__(self, server_host: str, server_port: int, output_file: str) -> None:
        self.server_host = server_host
        self.server_port = server_port
        self.lottery = Lottery(output_file)

    # def _handle_client(self, client_socket):
    #     action = "handle-client"
    #     message_amount = 0
    #     try:
    #         logger.info("read-bet-lines", logger.LogResult.in_progress)
    #         bet_line = protocol.read_bet_line(client_socket)
    #         while bet_line != b'':
    #             bet = protocol.deserialize_bet_line(bet_line)
    #             self.lottery.store_bets([bet])
    #             bet_line = protocol.read_bet_line(client_socket)
    #             message_amount += 1
    #         logger.info("read-bet-lines", logger.LogResult.success, "messages-amount", message_amount)
    #         logger.info("check-winners", logger.LogResult.in_progress)
    #         for bet in self.lottery.load_bets():
    #             # logger.info("check-winners", logger.LogResult.in_progress, "current", bet.first_name, "number", bet.number)
    #             if self.lottery.has_won(bet):
    #                 protocol.send_winner_line(client_socket, bet)
    #         protocol.send_separator(client_socket)
    #     except Exception as e:
    #         logger.error(
    #             action, logger.LogResult.fail, "messages-amount", message_amount, "error", e
    #         )
    #         raise e

    def run(self):
        action = "accept-connection"
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
            server_socket.bind((self.server_host, self.server_port))
            server_socket.listen()
            while True:
                try:
                    logger.info(action, logger.LogResult.in_progress)
                    client_socket, _ = server_socket.accept()
                except Exception as e:
                    logger.error(action, logger.LogResult.fail)
                    raise e
                logger.info(action, logger.LogResult.success)

                self._handle_client(client_socket)


    def _handle_client(self, client_socket):
        action = "handle-client"
        message_amount = 0
        try:
            logger.info("read-bet-lines", logger.LogResult.in_progress)
            bet_batch = protocol.read_bet_batch(client_socket)
            while len(bet_batch) > 0:
                bets = []
                for bet in bet_batch:
                    bets.append(bet)
                bet_batch = protocol.read_bet_batch(client_socket)
                self.lottery.store_bets(bets)
                message_amount += 1
            logger.info("read-bet-lines", logger.LogResult.success, "messages-amount", message_amount)
            logger.info("check-winners", logger.LogResult.in_progress)
            for bet in self.lottery.load_bets():
                #logger.info("check-winners", logger.LogResult.in_progress, "bet", bet)
                if self.lottery.has_won(bet):
                    protocol.send_winner_line(client_socket, bet)
            protocol.send_separator(client_socket)
        except Exception as e:
            logger.error(
                action, logger.LogResult.fail, "messages-amount", message_amount, "error", e
            )
            raise e