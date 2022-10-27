#!/usr/bin/env python
#
# SPDX-FileCopyrightText: 2022 Buoyant Inc.
# SPDX-License-Identifier: Apache-2.0
#
# Copyright 2022 Buoyant Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.  You may obtain
# a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import io
import json
import os
import random
import requests
import threading
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


class RateCounter:
    """
    RateCounter counts events in an N-second window and averages them. It does this
    by maintaining a (thread-safe) set of buckets, one per second, and providing a way
    to increment the bucket's counters.
    """

    def __init__(self, number_of_buckets):
        self.number_of_buckets = number_of_buckets
        self.first_bucket = None
        self.buckets = [0] * number_of_buckets
        self.lock = threading.Lock()

    def __str__(self) -> str:
        with self.lock:
            return f"RateCounter@{self.first_bucket}: {self.buckets}"

    def current_rate(self):
        """
        Returns the current rate as a float.
        """
        with self.lock:
            return sum(self.buckets) / self.number_of_buckets

    def mark(self, now=None):
        """
        Mark that a request has happened.
        """

        if not now:
            now = time.time()

        with self.lock:
            if not self.first_bucket:
                self.first_bucket = now
            
            bucket = now - self.first_bucket

            if bucket >= self.number_of_buckets:
                # We've moved past the end of the buckets, so slide the whole
                # window over.
                number_past = bucket - self.number_of_buckets + 1

                self.first_bucket += number_past

                if number_past >= self.number_of_buckets:
                    self.buckets = [0] * self.number_of_buckets
                else:
                    self.buckets = self.buckets[number_past:] + [0] * number_past

                bucket = now - self.first_bucket

            self.buckets[bucket] += 1


class BaseServer(BaseHTTPRequestHandler):
    delay_buckets = []
    error_fraction = 0
    max_rate = 0.0

    @classmethod
    def setup_from_environment(cls, *args, **kwargs):
        delay_buckets = os.environ.get("DELAY_BUCKETS", None)
        error_fraction = int(os.environ.get("ERROR_FRACTION", 0))
        max_rate = float(os.environ.get("MAX_RATE", 0.0))

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

        cls.max_rate = max(max_rate, 0.0)

        print(f"{cls.__name__}: max_rate env {max_rate} => {cls.max_rate}")

        if cls.max_rate >= 0.1:
            print(f"{cls.__name__}: max_rate is {cls.max_rate} requests per second, setting up rate counter")
            cls.rate_counter = RateCounter(10)
        else:
            cls.rate_counter = None

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


class FaceServer(BaseServer):
    # The face server is a bit more complicated. It makes requests to the
    # color service and the smiley service, and coalesces the results into a
    # single response.
    # 
    # We have defaults for all the services.

    defaults = {
        "color": "grey",
        "smiley": Smileys["Cursing"],

        "color-504": "pink",
        "smiley-504": Smileys["Sleeping"],

        "color-ratelimit": "pink",
        "smiley-ratelimit": Smileys["Kaboom"],

        "shape": """M14,15 m-5,-5 l10,10 m0,-10 l-10,10
                    M36,15 m-5,-5 l10,10 m0,-10 l-10,10
                    M10,34.5 c0,-7.5 30,-7.5 30,0""",
        "quote": "You fell victim to one of the classic blunders!",
    }

    def do_GET(self):
        start = time.time()

        rdict = {}
        errors = []

        if self.path == "/rl":
            self.standard_response({"rl": self.__class__.rate_counter.current_rate()})
            return

        ratestr = ""
        ratelimited = False

        if self.__class__.rate_counter:
            self.__class__.rate_counter.mark(int(start))
            rate = self.__class__.rate_counter.current_rate()
            ratestr = ", %.1f RPS" % rate

            if rate >= self.__class__.max_rate:
                ratelimited = True

        if ratelimited:
                rdict = {
                    "color": self.__class__.defaults["color-ratelimit"],
                    "smiley": self.__class__.defaults["smiley-ratelimit"],
                }
                errors.append("ratelimit")
        else:
            for svc, key in [ ( "color", "color" ), ( "smiley", "smiley" ) ]:
                stat, value = self.make_request(svc, key)

                if stat != 200:
                    errors.append(f"{svc}: {stat}")

                    for errkey in [ f"{svc}-{stat}", f"{svc}-{stat // 100}xx", svc ]:
                        errval = self.defaults.get(errkey, None)

                        if errval is not None:
                            value = errval
                            break

                rdict[key] = value

        end = time.time()
        latency_ms = delta_ms(start, end)

        if errors:
            rdict["errors"] = errors

        print(f"face ({latency_ms}{ratestr}): errors {errors} rdict {rdict}")

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
        "face": FaceServer,
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
