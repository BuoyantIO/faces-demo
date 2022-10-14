#!/usr/bin/env python

import io
import json
import os
import random
import requests
import time

from http.server import ThreadingHTTPServer, BaseHTTPRequestHandler

HOST_NAME = ""
PORT_NUMBER = 8000

Smileys = {
    "Smiling":  "&#x1F603;",
    "Sleeping": "&#x1F634;",
    "Cursing":  "&#x1F92C;",
    "Kaboom":   "&#x1F92F;",
}

SHAPES = [
    # These are faces made with make_faces.py.
    # ---- 6, 1.0, 0.0, 1.0 (11)
    """
    M14,15 m-6,0 a6,6.0 -0.0 1,0 12,0 a6,6.0 -0.0 1,0 -12,0
    M36,15 m-6,0 a6,6.0 0.0 1,0 12,0 a6,6.0 0.0 1,0 -12,0
    M10,27.0 c0,15.0 30,15.0 30,0
    """,

    # ---- 6, 0.8, -10.0, 0.8 (11)
    """
    M14,15 m-6,0 a6,4.8 10.0 1,0 12,0 a6,4.8 10.0 1,0 -12,0
    M36,15 m-6,0 a6,4.8 -10.0 1,0 12,0 a6,4.8 -10.0 1,0 -12,0
    M10,28.0 c0,12.0 30,12.0 30,0
    """,

    # ---- 2, 2.5, 15.0, 0.2 (10)
    """
    M15,15 m-2,0 a2,5.0 -15.0 1,0 4,0 a2,5.0 -15.0 1,0 -4,0
    M35,15 m-2,0 a2,5.0 15.0 1,0 4,0 a2,5.0 15.0 1,0 -4,0
    M10,31.0 c0,3.0 30,3.0 30,0
    """,

    # ---- 4, 0.5, 10.0, 1.0 (10)
    """
    M15,15 m-4,0 a4,2.0 -10.0 1,0 8,0 a4,2.0 -10.0 1,0 -8,0
    M35,15 m-4,0 a4,2.0 10.0 1,0 8,0 a4,2.0 10.0 1,0 -8,0
    M10,27.0 c0,15.0 30,15.0 30,0
    """,

    # ---- 4, 0.5, 10.0, -0.2 (10)
    """
    M15,15 m-4,0 a4,2.0 -10.0 1,0 8,0 a4,2.0 -10.0 1,0 -8,0
    M35,15 m-4,0 a4,2.0 10.0 1,0 8,0 a4,2.0 10.0 1,0 -8,0
    M10,33.0 c0,-3.0 30,-3.0 30,0
    """,

    # ---- 4, 0.7, 0.0, -0.5 (11)
    """
    M14,15 m-4,0 a4,2.8 -0.0 1,0 8,0 a4,2.8 -0.0 1,0 -8,0
    M36,15 m-4,0 a4,2.8 0.0 1,0 8,0 a4,2.8 0.0 1,0 -8,0
    M10,34.5 c0,-7.5 30,-7.5 30,0
    """,

    # ---- 6, 1.0, 0.0, -1.0 (11)
    """
    M14,15 m-6,0 a6,6.0 -0.0 1,0 12,0 a6,6.0 -0.0 1,0 -12,0
    M36,15 m-6,0 a6,6.0 0.0 1,0 12,0 a6,6.0 0.0 1,0 -12,0
    M10,37.0 c0,-15.0 30,-15.0 30,0
    """
]

# # The UI uses red, green, and grey, so don't include them here.
# COLORS = [
#     "cyan",
#     "blue",
#     "orange",
#     "purple",
# ]
COLORS = [ "green" ]

# These are the quotations from the original Quote of the Moment service.
QUOTES = [
    "Abstraction is ever present.",
    "A late night does not make any sense.",
    "A principal idea is omnipresent, much like candy.",
    "Nihilism gambles with lives, happiness, and even destiny itself!",
    "The light at the end of the tunnel is interdependent on the relatedness of motivation, subcultures, and management.",
    "Utter nonsense is a storyteller without equal.",
    "Non-locality is the driver of truth. By summoning, we vibrate.",
    "A small mercy is nothing at all?",
    "The last sentence you read is often sensible nonsense.",
    "668: The Neighbor of the Beast."
]

def delta_ms(start_time, end_time):
    # Convert latency to milliseconds, rounding normally
    return int(((end_time - start_time) * 1000) + .5)


class BaseServer(BaseHTTPRequestHandler):
    delay_buckets = []
    error_fraction = 0

    @classmethod
    def setup_from_environment(cls, *args, **kwargs):
        delay_buckets = os.environ.get("DELAY_BUCKETS", None)
        error_fraction = int(os.environ.get("ERROR_FRACTION", 0))

        if delay_buckets:
            for bucket_str in delay_buckets.split(","):
                bucket = None

                try:
                    bucket = int(bucket_str)
                except ValueError:
                    pass

                if bucket is not None:
                    bucket = max(bucket, 0)
                    cls.delay_buckets.append(bucket)

        print(f"{cls.__name__}: delay_buckets env {delay_buckets} => {cls.delay_buckets}")

        cls.error_fraction = min(max(error_fraction, 0), 100)

        print(f"{cls.__name__}: error_fraction env {error_fraction} => {cls.error_fraction}")

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

    def do_HEAD(self):
        self.standard_headers()

    def do_POST(self):
        self.send_error(405, "Method not allowed")

    def do_GET(self):
        raise NotImplementedError("GET must be provided by a subclass")

    def standard_response(self, data):
        if self.__class__.delay_buckets:
            delay_ms = random.choice(self.__class__.delay_buckets)
            time.sleep(delay_ms / 1000)

        if self.__class__.error_fraction > 0:
            if random.randint(0, 99) <= self.__class__.error_fraction:
                self.send_error(500, "Error fraction triggered")
                return

        response = {
            "path": self.path,
            "client_address": self.client_address,
            "method": self.command,
            "headers": dict(self.headers),
            "status": 200,
        }
        response.update(data)

        self.standard_headers()
        self.wfile.write(json.dumps(response).encode("utf-8"))

    def standard_headers(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()


class ShapeServer(BaseServer):
    def do_GET(self):
        self.standard_response({"shape": random.choice(SHAPES)})


class ColorServer(BaseServer):
    def do_GET(self):
        self.standard_response({"color": random.choice(COLORS)})


class QuoteServer(BaseServer):
    def do_GET(self):
        self.standard_response({"quote": random.choice(QUOTES)})


class SmileyServer(BaseServer):
    def do_GET(self):
        self.standard_response({"smiley": Smileys["Smiling"]})


class CompositeServer(BaseServer):
    # The composite server is a bit more complicated. It makes requests to the
    # color service and the shape service _and_ the quote service, and coalesces
    # the responses into a single response.
    # 
    # We have defaults for all the services.

    defaults = {
        "color": "grey",
        "smiley": Smileys["Cursing"],
        "shape": """M14,15 m-5,-5 l10,10 m0,-10 l-10,10
                    M36,15 m-5,-5 l10,10 m0,-10 l-10,10
                    M10,34.5 c0,-7.5 30,-7.5 30,0""",
        "quote": "You fell victim to one of the classic blunders!",
    }

    def do_GET(self):
        start = time.time()

        cstat, color = self.make_request("color", "color")
        sstat, smiley = self.make_request("smiley", "smiley")
        # qstat, quote = self.make_request("quote", "quote")

        end = time.time()

        latency_ms = delta_ms(start, end)

        errors = []

        if cstat != 200:
            color = self.defaults["color"]
            errors.append(f"color: {cstat}")

        if sstat != 200:
            smiley = self.defaults["smiley"]
            errors.append(f"smiley: {sstat}")

        # if qstat != 200:
        #     quote = self.defaults["quote"]
        #     errors.append(f"quote: {qstat}")

        rdict = {
            "color": color,
            "smiley": smiley,
            # "quote": quote,
        }

        if errors:
            rdict["errors"] = errors

        print(f"composite ({latency_ms}): errors {errors}")

        self.standard_response(rdict)

    def make_request(self, service, keyword) -> tuple[int, str]:
        start = time.time()

        url = f"http://{service}/"

        response = requests.get(url)

        end = time.time()

        latency_ms = delta_ms(start, end)

        if response.status_code != 200:
            # So. We got an error. Propagate it.
            print(f"...{url} ({latency_ms}ms): {response.status_code}")
            return response.status_code, ""

        # We got a response. Try to grab the key from the JSON.
        value = response.json().get(keyword, "")

        if not value:
            # This is not how this is meant to go.
            print(f"...{url} ({latency_ms}ms): no {keyword} in response")
            return 400, ""

        # We got a value. Return it.
        print(f"...{url} ({latency_ms}ms): {value}")
        return 200, value


if __name__ == '__main__':
    import sys

    # Python 3, open as binary, then wrap in a TextIOWrapper with write-through.
    sys.stdout = io.TextIOWrapper(open(sys.stdout.fileno(), 'wb', 0),
                                  encoding="utf-8", write_through=True)

    server_type = os.environ.get("FACES_SERVICE", None)

    if not server_type:
        raise ValueError("FACES_SERVICE must be set")

    servers = {
        "color": ColorServer,
        "shape": ShapeServer,
        "quote": QuoteServer,
        "smiley": SmileyServer,
        "composite": CompositeServer,
    }

    server_class = servers.get(server_type, None)

    if not server_class:
        raise ValueError(f"Invalid FACES_SERVICE: {server_type}")

    server_class.setup_from_environment()

    httpd = ThreadingHTTPServer((HOST_NAME, PORT_NUMBER), server_class)

    print(time.asctime(), 'Server UP - %s:%s' % (HOST_NAME, PORT_NUMBER))

    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass

    httpd.server_close()

    print(time.asctime(), 'Server DOWN - %s:%s' % (HOST_NAME, PORT_NUMBER))
