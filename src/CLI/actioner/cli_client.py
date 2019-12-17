################################################################################
#                                                                              #
#  Copyright 2019 Broadcom. The term Broadcom refers to Broadcom Inc. and/or   #
#  its subsidiaries.                                                           #
#                                                                              #
#  Licensed under the Apache License, Version 2.0 (the "License");             #
#  you may not use this file except in compliance with the License.            #
#  You may obtain a copy of the License at                                     #
#                                                                              #
#     http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                              #
#  Unless required by applicable law or agreed to in writing, software         #
#  distributed under the License is distributed on an "AS IS" BASIS,           #
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
#  See the License for the specific language governing permissions and         #
#  limitations under the License.                                              #
#                                                                              #
################################################################################

import os
import json
import urllib3
from six.moves.urllib.parse import quote

urllib3.disable_warnings()

class ApiClient(object):
    """
    A client for accessing a RESTful API
    """

    def __init__(self):
        """
        Create a RESTful API client.
        """
        self.api_uri = os.getenv('REST_API_ROOT', 'https://localhost')

        self.checkCertificate = False

        self.version = "0.0.1"

    def set_headers(self):
        from requests.structures import CaseInsensitiveDict
        return CaseInsensitiveDict({
            'User-Agent': "CLI"
        })

    @staticmethod
    def merge_dicts(*dict_args):
        result = {}
        for dictionary in dict_args:
            result.update(dictionary)
        return result

    def request(self, method, path, data=None, headers={}):
        from requests import request, RequestException

        url = '{0}{1}'.format(self.api_uri, path)

        req_headers = self.set_headers()
        req_headers.update(headers)

        body = None
        if data is not None:
            if 'Content-Type' not in req_headers:
                req_headers['Content-Type'] = 'application/yang-data+json'
            body = json.dumps(data)

        try:
            r = request(method, url, headers=req_headers, data=body, verify=self.checkCertificate)
            return Response(r)
        except RequestException:
            #TODO have more specific error message based
            return self._make_error_response('%Error: Could not connect to Management REST Server')

    def post(self, path, data={}):
        return self.request("POST", path, data)

    def get(self, path):
        return self.request("GET", path, None)

    def head(self, path):
        return self.request("HEAD", path, None)

    def put(self, path, data={}):
        return self.request("PUT", path, data)

    def patch(self, path, data={}):
        return self.request("PATCH", path, data)

    def delete(self, path):
        return self.request("DELETE", path, None)

    @staticmethod
    def _make_error_response(errMessage, errType='client', errTag='operation-failed'):
        import requests
        r = Response(requests.Response())
        r.content = {'ietf-restconf:errors':{ 'error':[ {
            'error-type':errType, 'error-tag':errTag, 'error-message':errMessage }]}}
        return r

    def cli_not_implemented(self, hint):
        return self._make_error_response('%Error: not implemented {0}'.format(hint))


class Path(object):
    def __init__(self, template, **kwargs):
        self.template = template
        self.params = kwargs
        self.path = template
        for k, v in kwargs.items():
            self.path = self.path.replace('{%s}' % k, quote(v, safe=''))

    def __str__(self):
        return self.path


class Response(object):
    def __init__(self, response):
        self.response = response
        self.status_code = response.status_code

        try:
            if response.content is None or len(response.content) == 0:
                self.content = None
            else:
                self.content = self.response.json()
        except ValueError:
            self.content = self.response.text

    def ok(self):
        return self.status_code >= 200 and self.status_code <= 299

    def errors(self):
        if self.ok():
            return {}

        errors = self.content

        if(not isinstance(errors, dict)):
            errors = {"error": errors}  # convert to dict for consistency
        elif('ietf-restconf:errors' in errors):
            errors = errors['ietf-restconf:errors']

        return errors

    def error_message(self, formatter_func=None):
        if hasattr(self, 'err_message_override'):
            return self.err_message_override

        err = self.errors().get('error')
        if err == None:
            return None
        if isinstance(err, list):
            err = err[0]
        if isinstance(err, dict):
            if formatter_func is not None:
                return formatter_func(self.status_code, err)
            return default_error_message_formatter(self.status_code, err)
        return str(err)

    def set_error_message(self, message):
        self.err_message_override = add_error_prefix(message)

    def __getitem__(self, key):
        return self.content[key]


def default_error_message_formatter(status_code, err_entry):
    if 'error-message' in err_entry:
        err_msg = err_entry['error-message']
        return add_error_prefix(err_msg)
    err_tag = err_entry.get('error-tag')
    if err_tag == 'invalid-value':
        return '%Error: validation failed'
    if err_tag == 'operation-not-supported':
        return '%Error: not supported'
    if err_tag == 'access-denied':
        return '%Error: not authorized'
    return '%Error: operation failed'

def add_error_prefix(err_msg):
    if not err_msg.startswith("%Error"):
        return '%Error: ' + err_msg
    return err_msg

