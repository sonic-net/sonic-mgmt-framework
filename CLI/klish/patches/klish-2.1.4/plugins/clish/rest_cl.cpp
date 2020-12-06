/*
###########################################################################
#
# Copyright 2019 Dell, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
###########################################################################
*/

extern "C" {
#include "private.h"
#include "lub/dump.h"
#include "nos_extn.h"
#include "logging.h"

#include <stdio.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <curl/curl.h>
#include <string.h>
#include <stdlib.h>
#include <pwd.h>
#include <cJSON.h>
#include <pthread.h>
#include <unistd.h>
#include <stdbool.h>
}

#include <string>

std::string REST_API_ROOT;

typedef struct {
    int size;
    std::string body;
} RestResponse;

typedef struct {
    const char* data;
    size_t length;
} PayloadData;

static ssize_t read_callback(void *ptr, size_t size, size_t nmemb, void *userp)
{
  PayloadData *upload = reinterpret_cast<PayloadData *>(userp);
  size_t max = size*nmemb;

  size_t copy_size = max;
  if (upload->length < max) {
      copy_size = upload->length;
  }

  memcpy(ptr, upload->data, copy_size);

  upload->length -= copy_size;
  upload->data += copy_size;

  return copy_size; 
}

static size_t write_callback(void *data, size_t size,
                                    size_t nmemb, void *userdata) {
  size_t realsize = size * nmemb;
  RestResponse *mem = reinterpret_cast<RestResponse *>(userdata);

  mem->body.append(reinterpret_cast<char*>(data), realsize);
  mem->size += realsize;

  return realsize;

}

int print_error(const char *str) {

    cJSON *ret_json = cJSON_Parse(str);
    if (!ret_json) {
        syslog(LOG_DEBUG, "clish_restcl: Failed parsing error string\r\n");
        return 0;
    }

    cJSON *ietf_err = cJSON_GetObjectItemCaseSensitive(ret_json, "ietf-restconf:errors");
    if (!ietf_err) {
        syslog(LOG_DEBUG, "clish_restcl: No errors\r\n");
        return 0;
    }

    cJSON *errors = cJSON_GetObjectItemCaseSensitive(ietf_err, "error");
    if (!errors) {
        syslog(LOG_DEBUG, "clish_restcl: No error\r\n");
        return 0;
    }

    cJSON *error;
    cJSON_ArrayForEach(error, errors) {
        cJSON *err_msg = cJSON_GetObjectItemCaseSensitive(error, "error-message");
        if (err_msg) {
            lub_dump_printf("%% Error: %s\r\n", err_msg->valuestring);
        /* Since error-message is an optional attribute, we need to check for "error-tag"
           and print the error-message accordingly */
        } else {
            std::string err_msg = "operation failed";

            cJSON* err_tag = cJSON_GetObjectItemCaseSensitive(error, "error-tag");
            if(err_tag == NULL) {
                lub_dump_printf("%% Error: %s\r\n", err_msg.c_str());
                return 1;
            }
            std::string err_tag_str = std::string {err_tag->valuestring};

            if(err_tag_str == "invalid-value") {
                err_msg = "validation failed";
            } else if (err_tag_str == "operation-not-supported") {
                err_msg = "not supported";
            } else if (err_tag_str == "access-denied") {
                err_msg = "not authorized";
            } else {
                err_msg = "operation failed";
            }
            lub_dump_printf("%% Error: %s\r\n", err_msg.c_str());
        }
        return 1;
    }
    return 0;
}


std::string rest_token;

CURL *curl =  NULL;

static int rest_set_curl_headers(bool use_token) {
    struct curl_slist* headerList = NULL;
    headerList = curl_slist_append(headerList, "accept: application/yang-data+json");
    headerList = curl_slist_append(headerList, "Content-Type: application/yang-data+json");

    if (rest_token.size() && use_token) {
        std::string auth_hdr = "Authorization: Bearer ";
        auth_hdr += rest_token;
        headerList = curl_slist_append(headerList, auth_hdr.c_str());
    }

    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headerList);

    return 0;
}

static int _init_curl() {

    curl_global_init(CURL_GLOBAL_ALL);

    curl = curl_easy_init();
    if (!curl) {
        return 1;
    }
    
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "CLI");

    curl_easy_setopt(curl, CURLOPT_READFUNCTION, read_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);

    if (REST_API_ROOT.find("https://") == 0) {
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
        curl_easy_setopt(curl, CURLOPT_SSL_VERIFYHOST, 0L);
    } else {
        curl_easy_setopt(curl, CURLOPT_UNIX_SOCKET_PATH, "/var/run/rest-local.sock");
    }

    return 0;
}

void rest_client_init() {
    char *root = getenv("REST_API_ROOT");

    REST_API_ROOT.assign(root ? root : "http://localhost");

    _init_curl();

    rest_set_curl_headers(true);
}

int rest_token_fetch(int *interval) {

    CURLcode res;
    std::string url;

    if (!curl) {
        syslog(LOG_WARNING, "curl handle is not yet initialized.");
        return 1;
    }
    
    url  = REST_API_ROOT;
    url.append("/authenticate");

    rest_set_curl_headers(false);
   
    RestResponse ret = {};
    ret.size = 0;

    if(curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());

        curl_easy_setopt(curl, CURLOPT_UPLOAD, 0L);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &ret);
        
        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, "GET");

        res = curl_easy_perform(curl);
        /* Check for errors */
        if(res != CURLE_OK) {
            syslog(LOG_WARNING, "curl_easy_perform() for rest_token_fetch failed: %s\n",
                    curl_easy_strerror(res));
        } else {
            if (ret.size) {
                cJSON *ret_json = cJSON_Parse(ret.body.c_str());
                if (ret_json) {
                    cJSON *token = cJSON_GetObjectItemCaseSensitive(ret_json, "access_token");
                    if (token) {
                        rest_token.assign(token->valuestring);

                        setenv("REST_API_TOKEN", rest_token.c_str(), 1);

                        pyobj_set_rest_token(rest_token.c_str());

                        rest_set_curl_headers(true);

                        cJSON  *expiry = cJSON_GetObjectItemCaseSensitive(ret_json, "expires_in");
                        if (expiry) {
                            *interval = expiry->valueint;
                        }
                    } else {
                        syslog(LOG_DEBUG, "rest_token_fetch: No access_token");
                    }
                } else {
                    syslog(LOG_DEBUG, "rest_token_fetch: Failed parsing return string");
                }
            } else {
                syslog(LOG_DEBUG, "rest_token_fetch: No response received");
            }
        }
    }
    return 0;
}

std::string& rtrim(std::string& str, const std::string& chars = "\t\n\v\f\r ")
{
    str.erase(str.find_last_not_of(chars) + 1);
    return str;
}

int _parse_args (std::string &input, std::string &oper_s, std::string &url_s, std::string &body_s){

    size_t oper_pos = input.find("oper=");
    if (oper_pos != std::string::npos) {
        oper_pos += strlen("oper=");
    }

    size_t url_pos = input.find("url=");
    if (url_pos == std::string::npos) {
        syslog(LOG_ERR, "url= missing in %s\r\n", input.c_str());
        return 1;
    } else {
        /* 'oper' terminates just before 'url=' */
        oper_s = input.substr(oper_pos, url_pos - oper_pos);
        rtrim(oper_s);
    
        url_pos += strlen("url=");
    }

    url_s = REST_API_ROOT;
    
    size_t body_pos = input.find("body=");

    /* If "body=" doesnt exist, 'url' extends till end of string */
    if (body_pos == std::string::npos) {
        url_s += input.substr(url_pos);
    } else {
        /* 'url' terminates just before 'body=' */
        url_s += input.substr(url_pos, body_pos - url_pos);

        body_pos += strlen("body=");
        body_s = input.substr(body_pos);

        rtrim(body_s);
    }

    rtrim(url_s);
    
    return 0;
}

int rest_cl(char *cmd, const char *buff)
{
    CURLcode res;

    int ret_code = 1;
    std::string url, body, oper;
    std::string arg = buff;

    setenv("USER_COMMAND", cmd, 1);
    syslog(LOG_DEBUG, "clish_restcl: cmd=%s", cmd);

    _parse_args(arg, oper, url, body);
    syslog(LOG_DEBUG, "clish_restcl: [oper:%s][path:%s][body:%s]", oper.c_str(), url.c_str(), body.c_str());

    RestResponse ret = {};
    ret.size = 0;

    if(curl) {

        curl_easy_setopt(curl, CURLOPT_CUSTOMREQUEST, oper.c_str());

        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());

        PayloadData up_obj = {};
        if (body.size()) {

            up_obj.data = body.c_str();
            up_obj.length = body.size(); 
            curl_easy_setopt(curl, CURLOPT_READDATA, &up_obj);

            curl_easy_setopt(curl, CURLOPT_INFILESIZE_LARGE,
                    (curl_off_t)up_obj.length);

            curl_easy_setopt(curl, CURLOPT_UPLOAD, 1L);
        } else {
            curl_easy_setopt(curl, CURLOPT_UPLOAD, 0L);
        }

        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &ret);

        res = curl_easy_perform(curl);
        /* Check for errors */
        if(res != CURLE_OK) {
            lub_dump_printf("%%Error: Could not connect to Management REST Server\n");
            syslog(LOG_WARNING, "curl_easy_perform() failed: %s\n",
                    curl_easy_strerror(res));
        } else {
            int64_t http_code = 0;
            curl_easy_getinfo(curl, CURLINFO_RESPONSE_CODE, &http_code);
            if (ret.size) {
                syslog(LOG_DEBUG, "clish_restcl: http_code:%ld [%d:%s]", http_code, ret.size, ret.body.c_str());
                print_error(ret.body.c_str());
            } else {
                ret_code = 0;
            }
        }
    } else {
        lub_dump_printf("%%Error: Could not connect to Management REST Server\n");
        syslog(LOG_WARNING, "Couldn't initialize curl handle");
    }

    return ret_code;
}
