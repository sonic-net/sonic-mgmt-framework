import os
import sys
import unittest
import warnings
from unittest import mock

import requests
from urllib3.exceptions import InsecureRequestWarning


ACTIONER_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', 'actioner'))
sys.path.insert(0, ACTIONER_DIR)

import cli_client


def _response():
    response = requests.Response()
    response.status_code = 200
    response._content = b''
    return response


class ApiClientCertificateVerificationTest(unittest.TestCase):

    def test_loopback_endpoint_detection(self):
        loopback_urls = (
            'https://localhost',
            'https://localhost:443',
            'https://127.0.0.1',
            'https://127.0.0.1:8443',
            'https://127.0.0.2',
            'https://127.255.255.254',
            'https://[::1]',
            'https://[::1]:8443',
        )
        remote_urls = (
            'https://rest.example.com',
            'https://localhost.example.com',
            'https://192.0.2.1',
            'https://[::1',
        )

        for url in loopback_urls:
            with self.subTest(url=url):
                self.assertTrue(cli_client._is_loopback_endpoint(url))

        for url in remote_urls:
            with self.subTest(url=url):
                self.assertFalse(cli_client._is_loopback_endpoint(url))

    def test_loopback_request_skips_certificate_verification(self):
        with mock.patch.object(
                cli_client.ApiClient,
                '_ApiClient__api_root',
                'https://127.0.0.1'):
            with mock.patch.object(
                    cli_client.ApiClient,
                    '_ApiClient__session') as session:
                session.request.return_value = _response()

                cli_client.ApiClient().get('/restconf/data')

                self.assertFalse(session.request.call_args.kwargs['verify'])

    def test_remote_request_verifies_server_certificate(self):
        with mock.patch.object(
                cli_client.ApiClient,
                '_ApiClient__api_root',
                'https://rest.example.com'):
            with mock.patch.object(
                    cli_client.ApiClient,
                    '_ApiClient__session') as session:
                session.request.return_value = _response()

                cli_client.ApiClient().get('/restconf/data')

                self.assertTrue(session.request.call_args.kwargs['verify'])

    def test_loopback_warning_suppression_is_scoped_to_request(self):
        def request_with_warning(*args, **kwargs):
            warnings.warn('loopback request', InsecureRequestWarning)
            return _response()

        with mock.patch.object(
                cli_client.ApiClient,
                '_ApiClient__api_root',
                'https://localhost'):
            with mock.patch.object(
                    cli_client.ApiClient,
                    '_ApiClient__session') as session:
                session.request.side_effect = request_with_warning

                with warnings.catch_warnings():
                    warnings.simplefilter('error', InsecureRequestWarning)
                    cli_client.ApiClient().get('/restconf/data')
                    with self.assertRaises(InsecureRequestWarning):
                        warnings.warn('outside request', InsecureRequestWarning)


if __name__ == '__main__':
    unittest.main()
