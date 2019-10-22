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
import json
import urllib3
from six.moves.urllib.parse import quote

class ApiClient(object):
	"""
	A client for accessing a RESTful API
	"""
	def __init__(self, api_uri=None):
		"""
		Create a RESTful API client.
		"""
		api_uri="https://localhost:443"
		self.api_uri	= api_uri
		
		self.checkCertificate = False
		
		if not self.checkCertificate:
			urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

		self.version = "0.0.1"

	def set_headers(self, nonce = None):
		from base64 import b64encode
		from hashlib import sha256
		from platform import platform, python_version
		from hmac import new

		if not nonce:
			from time import time
			nonce = int(time())

		return {
			'User-Agent': "PythonClient/{0} ({1}; Python {2})".format(self.version, 
										 platform(True), 
										 python_version())
		}

	@staticmethod
	def merge_dicts(*dict_args):
		result = {}
		for dictionary in dict_args:
			result.update(dictionary)

		return result

	def request(self, method, path, data = {}, headers = {}):
		from requests import request

		url = '{0}{1}'.format(self.api_uri, path)
		headers = self.merge_dicts(self.set_headers(), headers)

		if method == "GET":
			params = {}
			if data is not None:
				params.update(data)
			return request(method, url, headers=headers, params=params, verify=self.checkCertificate)
		else:
			body = None
			if data is not None:
				body = json.dumps(data)
			return request(method, url, headers=headers, data=body, verify=self.checkCertificate)

	def post(self, path, data = {}):
		return Response(self.request("POST", path, data, {'Content-Type': 'application/yang-data+json'}))

	def get(self, path, params = {}):
		return Response(self.request("GET", path, params))

	def put(self, path, data = {}):
		return Response(self.request("PUT", path, data, {'Content-Type': 'application/yang-data+json'}))

	def patch(self, path, data = {}):
		return Response(self.request("PATCH", path, data, {'Content-Type': 'application/yang-data+json'}))

	def delete(self, path):
		return Response(self.request("DELETE", path, None))

class Path(object):
	def __init__(self, template, **kwargs):
		self.template = template
		self.params = kwargs
		self.path = template
		for k, v in kwargs.items():
			self.path = self.path.replace('{%s}'%k, quote(v, safe=''))

	def __str__(self):
		return self.path

class Response(object):
	def __init__(self, response):
		self.response = response

		try:
			if len(response.content) == 0:
				self.content = None
			else:
				self.content = self.response.json()
		except ValueError:
			self.content = self.response.text

	def ok(self):
		return self.response.status_code >= 200 and self.response.status_code <= 299

	def errors(self):
		if self.ok():
			return {}

		errors = self.content

		if(not isinstance(errors, dict)):
			errors = {"error": errors} # convert to dict for consistency
		elif('ietf-restconf:errors' in errors):
			errors = errors['ietf-restconf:errors']

		return errors

	def error_message(self, formatter_func=None):
		err = self.errors().get('error')
		if err == None:
			return None
		if isinstance(err, list):
			err = err[0]
		if isinstance(err, dict):
			if formatter_func is not None:
				return formatter_func(err)
			return default_error_message_formatter(err)
		return str(err)

	def __getitem__(self, key):
		return self.content[key]

def default_error_message_formatter(err_entry):
	if 'error-message' in err_entry:
		return err_entry['error-message']
	err_tag = err_entry.get('error-tag')
	if err_tag == 'invalid-value':
		return '%Error: validation failed'
	if err_tag == 'operation-not-supported':
		return '%Error: not supported'
	if err_tag == 'access-denied':
		return '%Error: not authorized'
	return '%Error: operation failed'


