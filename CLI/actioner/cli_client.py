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
import requests
from requests.structures import CaseInsensitiveDict
from six.moves.urllib.parse import quote
from collections import OrderedDict
from cli_log import log_info, log_warning

urllib3.disable_warnings()


class ApiClient(object):
    """REST API client to connect to the SONiC management REST server.
    Customized for CLI actioner use.
    """

    # Initialize API root and session
    __api_root = os.getenv('REST_API_ROOT', 'https://localhost')
    __session = requests.Session()

    def request(self, method, path, data=None, headers={}, query=None, response_type=None):

        url = "{0}{1}".format(ApiClient.__api_root, path)

        req_headers = CaseInsensitiveDict({'User-Agent': 'sonic-cli'})
        req_headers.update(headers)

        body = None
        if data is not None:
            if 'Content-Type' not in req_headers:
                req_headers['Content-Type'] = 'application/yang-data+json'
            body = json.dumps(data)

        try:
            r = ApiClient.__session.request(
                method,
                url,
                headers=req_headers,
                data=body,
                params=query,
                verify=False)

            return Response(r, response_type)

        except requests.RequestException as e:
            log_info("cli_client request exception: {}", e)
            # TODO have more specific error message based
            msg = '%Error: Could not connect to Management REST Server'
            return ApiClient.__new_error_response(msg)

    def post(self, path, data=None, response_type=None):
        return self.request("POST", path, data, response_type=response_type)

    def get(self, path, depth=None, ignore404=True, response_type=None):
        q = self.prepare_query(depth=depth)
        resp = self.request("GET", path, query=q, response_type=response_type)
        if ignore404 and resp.status_code == 404:
            resp.status_code = 200
            resp.content = None
        return resp

    def head(self, path, depth=None):
        q = self.prepare_query(depth=depth)
        return self.request("HEAD", path, query=q)

    def put(self, path, data):
        return self.request("PUT", path, data)

    def patch(self, path, data):
        return self.request("PATCH", path, data)

    def delete(self, path, ignore404=True, deleteEmptyEntry=False):
        q = self.prepare_query(deleteEmptyEntry=deleteEmptyEntry)
        resp = self.request("DELETE", path, data=None, query=q)
        if ignore404 and resp.status_code == 404:
            resp.status_code = 204
            resp.content = None
        return resp

    @staticmethod
    def prepare_query(depth=None, deleteEmptyEntry=None):
        query = {}
        if depth != None and depth != "unbounded":
            query["depth"] = depth
        if deleteEmptyEntry is True:
            query["deleteEmptyEntry"] = "true"
        return query

    @staticmethod
    def __new_error_response(errMessage, errType='client', errTag='operation-failed'):
        r = Response(requests.Response())
        r.content = {'ietf-restconf:errors': {'error': [{
            'error-type': errType, 'error-tag': errTag, 'error-message': errMessage}]}}
        return r

    def cli_not_implemented(self, hint):
        return self.__new_error_response('%Error: not implemented {0}'.format(hint))


class Path(object):
    def __init__(self, template, **kwargs):
        self.template = template
        self.params = kwargs
        self.path = template
        for k, v in list(kwargs.items()):
            self.path = self.path.replace('{%s}' % k, quote(v, safe=''))

    def join(self, p, **kwargs):
        if not isinstance(p, Path):
            p = Path(p, **kwargs)
        return Path(self.path + Path._withslash(p.path))

    @staticmethod
    def _withslash(s):
        return "/"+s if s and not s.startswith("/") else s

    def __str__(self):
        return self.path


class Response(object):
    def __init__(self, response, response_type=None):
        self.response = response
        self.response_type = response_type
        self.status_code = (response.status_code if response.status_code else 0)
        self.content = response.content

        try:
            if response.content is None or len(response.content) == 0:
                self.content = None
            elif self.response_type and self.response_type.lower() == 'string':
                self.content = str(response.content).decode('string_escape')
            elif has_json_content(response):
                self.content = json.loads(response.content, object_pairs_hook=OrderedDict)
        except ValueError:
            # TODO Can we set status_code to 5XX in this case???
            # Json parsing can fail only if server returned bad json
            log_warning("Server returned invalid json for url {}", self.response.url)
            self.content = response.content

    def ok(self):
        return self.status_code >= 200 and self.status_code <= 299

    def errors(self):
        if self.ok():
            return {}

        errors = self.content
        if not isinstance(errors, dict):
            errors = {"error": errors}  # convert to dict for consistency
        elif "ietf-restconf:errors" in errors:
            errors = errors["ietf-restconf:errors"]
        return errors

    def error_message(self, formatter_func=None):
        if hasattr(self, "err_message_override"):
            return self.err_message_override

        err = self.errors().get("error")
        if err is None:
            return None
        if isinstance(err, list):
            err = err[0]
        if isinstance(err, dict):
            return format_error_message(self.status_code, err, formatter_func)
        return str(err)

    def set_error_message(self, message):
        self.err_message_override = add_error_prefix(message)

    def __getitem__(self, key):
        return self.content[key]


def has_json_content(resp):
    ctype = resp.headers.get("Content-Type")
    return ctype is not None and "json" in ctype


def format_error_message(status_code, err_entry, formatter_func=None):
    if formatter_func is not None:
        err_msg = formatter_func(status_code, err_entry)
        if err_msg:
            return add_error_prefix(err_msg)
    if "error-message" in err_entry:
        err_msg = err_entry["error-message"]
        return add_error_prefix(err_msg)
    err_tag = err_entry.get("error-tag")
    if err_tag == "invalid-value":
        return "%Error: validation failed"
    if err_tag == "operation-not-supported":
        return "%Error: not supported"
    if err_tag == "access-denied":
        return "%Error: not authorized"
    return "%Error: operation failed"


def add_error_prefix(err_msg):
    if not err_msg.startswith("%Error"):
        return "%Error: " + err_msg
    return err_msg


class ErrorData(object):
    def __init__(self, response, index=0, formatter_func=None):
        """Constructs an ErrorData object by parsing the RESTCONF error message
        from a Response object.
        """
        err_list = response.errors().get("error")
        err = err_list[index] if err_list and len(err_list) > index else {}
        err_info = err.get("error-info", {})
        self.status = response.status_code
        self.type = err.get("error-type")
        self.tag = err.get("error-tag")
        self.app_tag = err.get("error-app-tag")
        self.path = err.get("error-path")
        self.cvl_err = err_info.get("cvl-error")
        self.message = format_error_message(response.status_code, err, formatter_func)
