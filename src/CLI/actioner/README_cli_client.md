# Generic RESTful Python client

A generic RESTful Python client for interacting with JSON APIs.

## Usage

To use this client you just need to import ApiClient and initialize it with an URL endpoint

    from cli_client import ApiClient
    api = ApiClient('#your_api_endpoint') #your_api_endpoint is optional default is https://localhost:443

Now that you have a RESTful API object you can start sending requests.


## Making a request

The framework supports GET, PUT, POST, PATCH and DELETE requests:

    from cli_client import ApiClient
    api = ApiClient()
    response = api.get('/authors/')
    response = api.post('/authors/', {'title': 'Broadcom', 'author': 'Faraaz Mohammed'})
    response = api.put('/author/faraaz/', {'dob': '06/09/2006'})
    response = api.delete('/author/faraaz/')

## To get the Response Data
response = api.get('/authors/')
Use response.content object

For Successful request response.content will contain valid JSON data
For Non-Successful request response.content will contain errors object returned by REST Server

## Verifying Requests

Two helpers are built in to verify the success of requests made. `ok()` checks for a 20x status code and returns a boolean, `errors()` returns the body content as a dict object if the status code is not 20x:

    response = api.get('/books/')
    if response.ok():
        print 'Success!'
    else:
        print req.errors()

