# Generic REST Client for CLI

Actioner scripts can use the generic rest client **cli_client.py** to communicate to REST Server.
This client is tailor made to connect to local REST Server and handle the CLI actioner usecases.
All future CLI-REST integration enhancements (for RBAC) can be handled in this tool without
affecting individual actioner scripts.

To use this tool, first create an `ApiClient` object.

```python
import cli_client as cc

api = cc.ApiClient()
```

Create a path object for target REST resource. It accepts parameterized path template and parameter
values. Path template is similar to the template used by swagger. Parameter values will be URL-encoded
and substituted in the template to get REST resource path.

```python
path = cc.Path('/restconf/data/openconfig-acl:acl/acl-sets/acl-set={name},{type}/acl-entries',
            name='MyNewAccessList', type='ACL_IPV4')
```

Invoke REST API.. `ApiClient` object supports get, post, put, patch and delete operations.
All these operations send a REST request and return a response object wrapping the REST response data.

```python
response = api.get(path)
```

Check API status through `response.ok()` function, which returns true if API was success - REST server
returned HTTP 2xx status code. `ok()` function returns false if server returned any other status response.

```python
if response.ok() {
    respJson = response.content
    # invoke renderer to display respJson
} else {
    print(response.error_message())
}
```

If request was successful, `response.content` will hold the response data as JSON dictionary object.
If request failed, `response.content` will hold error JSON returned by the server. CLI displayable
error message can be extracted using `response.error_message()` function.

Examples of other REST API calls.

```python
jsonDict = {}
jsonDict["acl-set"]=[{ "name":the_acl_name, .... }] # construct your request data json

# POST request
response = api.post(path, data=jsonDict)

# PUT request
reponse = api.put(path, data=jsonDict)

# PATCH request
response = api.patch(path, data=jsonDict)

# DELETE request
response = api.delete(path)
```

Above example used `response.error_message()` function to get ser displayable error message from
the REST response. REST server always returns error information in standard RESTCONF error format.
The `error_message()` function looks for the **error-message** attribute to get user displayable message.
A generic message will be returned based on **error-tag** attribute value if there was no **error-message**
attribute. This can be customized through an error formatter function as shown below.
Default error formatter is sufficient for most of the cases.

```python
if not response.ok():
     print(response.error_message(formatter_func=my_new_error_message_formatter))
```

The custom formtter function would receive RESTCONF error json and should return a string.
Below is a sample implementation which uses error-type attribute to format an error message.

```python
def my_new_error_message_formatter(error_entry):
    if err_entry['error-type'] == 'protocol':
        return "System error.. Please reboot!!"
    return "Application error.. Please retry."
```

