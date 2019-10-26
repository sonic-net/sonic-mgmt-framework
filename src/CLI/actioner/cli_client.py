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
		params = {}
		headers = self.merge_dicts(self.set_headers(), headers)

		if method == "GET":
			params.update(data)
			return request(method, url, headers=headers, params=params, verify=self.checkCertificate)
		else:
			return request(method, url, headers=headers, params=params, data=json.dumps(data), verify=self.checkCertificate)

	def post(self, path, data = {}):
		return Response(self.request("POST", path, data, {'Content-Type': 'application/json'}))

	def get(self, path, data = {}):
		return Response(self.request("GET", path, data))

	def put(self, path, data = {}):
		return Response(self.request("PUT", path, data, {'Content-Type': 'application/json'}))

	def patch(self, path, data = {}):
		return Response(self.request("PATCH", path, data, {'Content-Type': 'application/json'}))

	def delete(self, path, data = {}):
		return Response(self.request("DELETE", path, data))

class Response(object):
	def __init__(self, response):
		self.response = response

		try:
			self.content = self.response.json()
		except ValueError:
			self.content = self.response.text

	def ok(self):
		import requests
		return self.response.status_code == requests.codes.ok

	def errors(self):
		if self.ok():
			return {}

		errors = self.content

		if(not isinstance(errors, dict)):
			errors = {"error": errors} # convert to dict for consistency
		elif('errors' in errors):
			errors = errors['ietf-restconf:errors']

		return errors

	def __getitem__(self, key):
		return self.content[key]
