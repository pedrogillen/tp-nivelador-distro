import socket

# TODO: Complete with a short-read/short-write tolerant implementation


def recv_all(socket: socket.socket, size):
    data = b""
    while len(data) < size:
        chunk = socket.recv(size - len(data))
        if chunk == b'':
            raise RuntimeError("socket connection broken")
        data += chunk
    return data


def send_all(socket: socket.socket, bytes):
    total_sent = 0
    while total_sent < len(bytes):
        sent = socket.send(bytes[total_sent:])
        total_sent += sent
    return total_sent