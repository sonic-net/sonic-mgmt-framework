################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or  #
#  its subsidiaries.                                                          #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");            #
#  you may not use this file except in compliance with the License.           #
#  You may obtain a copy of the License at                                    #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                              #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software        #
#  distributed under the License is distributed on an "AS IS" BASIS,          #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and        #
#  limitations under the License.                                             #
#                                                                              #
################################################################################
#
# Tests for the local REST Unix domain socket transport added to
# cli_client.py for sonic-buildimage#29504. Runs a real HTTP server bound
# to a temporary AF_UNIX socket and drives ApiClient/UnixSocketAdapter
# against it, rather than mocking the transport.

import json
import os
import shutil
import socketserver
import tempfile
import threading
import unittest
from http.server import BaseHTTPRequestHandler

import requests

import cli_client


class _RecordingHandler(BaseHTTPRequestHandler):
    """Minimal HTTP handler that records the request it received and
    returns a canned JSON response."""

    protocol_version = 'HTTP/1.1'

    def log_message(self, fmt, *args):
        pass  # keep test output quiet

    def _respond(self):
        self.server.last_method = self.command
        self.server.last_path = self.path
        self.server.last_headers = dict(self.headers)

        body = json.dumps({'ietf-restconf:restconf-state': {}}).encode('utf-8')
        self.send_response(200)
        self.send_header('Content-Type', 'application/yang-data+json')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        self._respond()

    def do_PATCH(self):
        length = int(self.headers.get('Content-Length', 0))
        self.server.last_body = self.rfile.read(length) if length else b''
        self._respond()


class _UnixHTTPServer(socketserver.UnixStreamServer):
    allow_reuse_address = True


class UnixSocketTransportTest(unittest.TestCase):
    """Exercises UnixSocketAdapter directly (independent of ApiClient's
    hardcoded socket path), against a real AF_UNIX HTTP server."""

    def setUp(self):
        self.tmpdir = tempfile.mkdtemp(prefix='uds-test-')
        self.sock_path = os.path.join(self.tmpdir, 'test.sock')

        self.httpd = _UnixHTTPServer(self.sock_path, _RecordingHandler)
        self.httpd.last_headers = None
        self.thread = threading.Thread(target=self.httpd.serve_forever)
        self.thread.daemon = True
        self.thread.start()

    def tearDown(self):
        self.httpd.shutdown()
        self.httpd.server_close()
        self.thread.join(timeout=5)
        shutil.rmtree(self.tmpdir, ignore_errors=True)

    def test_get_over_unix_socket(self):
        session = requests.Session()
        session.mount('http://localhost', cli_client.UnixSocketAdapter(self.sock_path))

        resp = session.get('http://localhost/restconf/data/ietf-restconf-monitoring:restconf-state')

        self.assertEqual(resp.status_code, 200)
        self.assertEqual(self.httpd.last_method, 'GET')
        self.assertEqual(
            self.httpd.last_path,
            '/restconf/data/ietf-restconf-monitoring:restconf-state')
        self.assertEqual(
            json.loads(resp.content), {'ietf-restconf:restconf-state': {}})

    def test_no_authorization_header_sent(self):
        session = requests.Session()
        session.mount('http://localhost', cli_client.UnixSocketAdapter(self.sock_path))

        session.get('http://localhost/restconf/data/openconfig-system:system',
                    headers={'User-Agent': 'sonic-cli'})

        self.assertNotIn('Authorization', self.httpd.last_headers)

    def test_patch_body_is_delivered(self):
        session = requests.Session()
        session.mount('http://localhost', cli_client.UnixSocketAdapter(self.sock_path))

        payload = b'{"openconfig-interfaces:config":{"name":"Vlan10"}}'
        resp = session.request(
            'PATCH', 'http://localhost/restconf/data/x',
            data=payload,
            headers={'Content-Type': 'application/yang-data+json'})

        self.assertEqual(resp.status_code, 200)
        self.assertEqual(self.httpd.last_body, payload)


class ApiClientUnixSocketIntegrationTest(unittest.TestCase):
    """Exercises ApiClient itself, with its local-socket adapter
    temporarily repointed at a test socket instead of the real
    /var/run/rest-local.sock."""

    def setUp(self):
        self.tmpdir = tempfile.mkdtemp(prefix='uds-test-')
        self.sock_path = os.path.join(self.tmpdir, 'test.sock')

        self.httpd = _UnixHTTPServer(self.sock_path, _RecordingHandler)
        self.httpd.last_headers = None
        self.thread = threading.Thread(target=self.httpd.serve_forever)
        self.thread.daemon = True
        self.thread.start()

        # ApiClient is only ever configured once, at import time, with the
        # real socket path. Repoint its mounted local adapter at our test
        # socket for the duration of this test, then restore it.
        session = cli_client.ApiClient._ApiClient__session
        self._orig_adapter = session.get_adapter(cli_client.REST_LOCAL_API_ROOT)
        session.mount(cli_client.REST_LOCAL_API_ROOT,
                      cli_client.UnixSocketAdapter(self.sock_path))

    def tearDown(self):
        session = cli_client.ApiClient._ApiClient__session
        session.mount(cli_client.REST_LOCAL_API_ROOT, self._orig_adapter)

        self.httpd.shutdown()
        self.httpd.server_close()
        self.thread.join(timeout=5)
        shutil.rmtree(self.tmpdir, ignore_errors=True)

    def test_apiclient_get_uses_local_socket_by_default(self):
        self.assertEqual(cli_client.ApiClient._ApiClient__api_root,
                          cli_client.REST_LOCAL_API_ROOT)

        resp = cli_client.ApiClient().get('/restconf/data/x', ignore404=False)

        self.assertEqual(resp.status_code, 200)
        self.assertEqual(self.httpd.last_method, 'GET')
        self.assertNotIn('Authorization', self.httpd.last_headers)


if __name__ == '__main__':
    unittest.main()
